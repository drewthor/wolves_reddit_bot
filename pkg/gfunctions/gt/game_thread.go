package gt

import (
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

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
	dailyAPIInfo, err := nba.GetDailyAPIPaths()
	if err != nil {
		log.Fatal(err)
	}
	dailyAPIPaths := dailyAPIInfo.APIPaths
	teams, err := nba.GetTeams(dailyAPIPaths.Teams)
	if err != nil {
		log.Fatal(err)
	}
	var team *nba.Team
	for _, t := range teams {
		if t.TriCode == teamTriCode {
			team = &t
		}
	}
	if team == nil {
		log.Println("failed to find team with TriCode: " + teamTriCode)
		return
	}
	scheduledGames, err := nba.GetCurrentTeamSchedule(dailyAPIPaths.TeamSchedule, team.ID)
	if err != nil {
		log.Fatal(err)
	}
	todaysGame, gameToday := scheduledGames[currentDateWestern]

	if gameToday {
		log.Println("game today")
		boxscore, err := nba.GetCurrentSeasonBoxscore(todaysGame.GameID, currentDateWestern)
		if err != nil {
			log.Fatal(err)
		}
		datastore := new(gcloud.Datastore)
		gameEvent, exists := datastore.GetTeamGameEvent(todaysGame.GameID, team.ID)
		subreddit := "SeattleSockeye"

		durationUntilGameStarts, err := boxscore.DurationUntilGameStarts()
		if err != nil {
			log.Fatal(err)
		}
		if (durationUntilGameStarts.Hours() < 1) && !boxscore.GameEnded() {
			log.Println("game in progress")

			log.Println("making post")
			redditClient := reddit.Client{}
			redditClient.Authorize()
			log.Println("authorized")
			title := boxscore.GetRedditGameThreadTitle(teamTriCode, teams)
			players, err := nba.GetPlayers(dailyAPIInfo.APISeasonInfoNode.SeasonYear)
			if err != nil {
				log.Fatal(err)
			}
			playersMap := map[string]nba.Player{}
			for _, player := range players {
				playersMap[player.ID] = player
			}
			content := boxscore.GetRedditGameThreadBodyString(playersMap, "" /*postGameThreadURL*/)

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

		if exists && gameEvent.GameThread && gameEvent.PostGameThread && !gameEvent.GameThreadComplete {
			log.Println("adding post game thread link to game thread")
			redditClient := reddit.Client{}
			redditClient.Authorize()
			thingURLMapping := redditClient.GetThingURLs([]string{gameEvent.PostGameThreadRedditPostFullname}, subreddit)
			postGameThreadURL := thingURLMapping[gameEvent.PostGameThreadRedditPostFullname]
			players, err := nba.GetPlayers(dailyAPIInfo.APISeasonInfoNode.SeasonYear)
			if err != nil {
				log.Fatal(err)
			}
			playersMap := map[string]nba.Player{}
			for _, player := range players {
				playersMap[player.ID] = player
			}
			content := boxscore.GetRedditGameThreadBodyString(playersMap, postGameThreadURL)
			redditClient.UpdateUserText(gameEvent.GameThreadRedditPostFullname, content)
			gameEvent.GameThreadComplete = true
			datastore.SaveTeamGameEvent(gameEvent)
		}
	}
}
