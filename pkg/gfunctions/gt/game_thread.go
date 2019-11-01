package gt

import (
	"log"
	"sync"
	"time"

	"github.com/drewthor/wolves_reddit_bot/apis/nba"
	"github.com/drewthor/wolves_reddit_bot/apis/reddit"
	"github.com/drewthor/wolves_reddit_bot/pkg/gcloud"
)

func CreateGameThread(teamTriCode nba.TriCode, wg *sync.WaitGroup) {
	defer wg.Done()
	currentTimeUTC := time.Now().UTC()
	// Issues occur when using eastern time for "today's games" as games on the west coast can still be going on
	// when the eastern time rolls over into the next day
	westCoastLocation, locationError := time.LoadLocation("America/Los_Angeles")
	if locationError != nil {
		log.Fatal(locationError)
	}
	currentTimeWestern := currentTimeUTC.In(westCoastLocation)
	currentDateWestern := currentTimeWestern.Format(nba.TimeDayFormat)
	log.Println(currentDateWestern)
	dailyAPIPaths := nba.GetDailyAPIPaths()
	teams := nba.GetTeams(dailyAPIPaths.Teams)
	team, foundTeam := teams[teamTriCode]
	if !foundTeam {
		log.Println("failed to find team with TriCode: " + teamTriCode)
		return
	}
	scheduledGames := nba.GetScheduledGames(dailyAPIPaths.TeamSchedule, team.ID)
	todaysGame, gameToday := scheduledGames[currentDateWestern]

	if gameToday {
		log.Println("game today")
		boxscore := nba.GetBoxscore(dailyAPIPaths.Boxscore, currentDateWestern, todaysGame.GameID)
		datastore := new(gcloud.Datastore)
		gameEvent, exists := datastore.GetTeamGameEvent(todaysGame.GameID, team.ID)
		subreddit := "Timberwolves"

		if (boxscore.DurationUntilGameStarts().Hours() < 1) && !boxscore.GameEnded() {
			log.Println("game in progress")

			log.Println("making post")
			redditClient := reddit.Client{}
			redditClient.Authorize()
			log.Println("authorized")
			title := boxscore.GetRedditGameThreadTitle(teamTriCode, teams)
			content := boxscore.GetRedditGameThreadBodyString(nba.GetPlayers(dailyAPIPaths.Players), "" /*postGameThreadURL*/)

			if exists && gameEvent.GameThread {
				log.Println("updating post")
				redditClient.UpdateUserText(gameEvent.GameThreadRedditPostFullname, content)
			} else {
				submitResponse := redditClient.SubmitNewPost(subreddit, title, content)
				gameEvent.GameThreadRedditPostFullname = submitResponse.JsonNode.DataNode.Fullname
			}
			gameEvent.CreatedTime = time.Now()
			gameEvent.GameID = todaysGame.GameID
			gameEvent.TeamID = team.ID
			gameEvent.GameThread = true
			datastore.SaveTeamGameEvent(gameEvent)
		}

		if exists && gameEvent.GameThread && gameEvent.PostGameThread {
			log.Println("adding post game thread link to game thread")
			redditClient := reddit.Client{}
			redditClient.Authorize()
			thingURLMapping := redditClient.GetThingURLs([]string{gameEvent.PostGameThreadRedditPostFullname}, subreddit)
			postGameThreadURL := thingURLMapping[gameEvent.PostGameThreadRedditPostFullname]
			content := boxscore.GetRedditGameThreadBodyString(nba.GetPlayers(dailyAPIPaths.Players), postGameThreadURL)
			redditClient.UpdateUserText(gameEvent.GameThreadRedditPostFullname, content)
		}
	}
}
