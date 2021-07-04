package pgt

import (
	"log"
	"sync"
	"time"

	"github.com/drewthor/wolves_reddit_bot/apis/nba"
	"github.com/drewthor/wolves_reddit_bot/apis/reddit"
	"github.com/drewthor/wolves_reddit_bot/pkg/gcloud"
)

func CreatePostGameThread(teamTriCode nba.TriCode, wg *sync.WaitGroup) {
	defer wg.Done()
	currentTimeUTC := time.Now().UTC()
	westCoastLocation, locationError := time.LoadLocation("America/Los_Angeles")
	if locationError != nil {
		log.Fatal(locationError)
	}
	currentTimeWestern := currentTimeUTC.In(westCoastLocation)
	currentDateWestern := currentTimeWestern.Format(nba.TimeDayFormat)
	dailyAPIPaths := nba.GetDailyAPIPaths()
	teams := nba.GetTeams(dailyAPIPaths.Teams)
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
	scheduledGames := nba.GetScheduledGames(dailyAPIPaths.TeamSchedule, team.ID)
	todaysGame, gameToday := scheduledGames[currentDateWestern]
	log.Println("checked for game")
	if gameToday {
		log.Println("game today")
		todaysGameScoreboard := nba.GetGameScoreboard(dailyAPIPaths.Scoreboard, currentDateWestern, todaysGame.GameID)
		boxscore := nba.GetBoxscore(dailyAPIPaths.Boxscore, currentDateWestern, todaysGame.GameID)
		if boxscore.GameEnded() {
			log.Println("game ended")
			// the nba api sometimes has not updated the record for the teams right at the end of games
			// see apis/nba/boxscore.go::UpdateTeamsRegularSeasonRecords
			currentGameNumber, foundCurrentGame := scheduledGames.CurrentGameNumber(todaysGame.GameID, todaysGame.SeasonStage)
			if foundCurrentGame {
				log.Println("updating records")
				boxscore.UpdateTeamsRegularSeasonRecords(currentGameNumber)
				boxscore.UpdateTeamsPlayoffsSeriesRecords()
			}
			if todaysGameScoreboard.EndTimeUTC != "" {
				gameEndTimeUTC, err := time.Parse(nba.UTCFormat, todaysGameScoreboard.EndTimeUTC)
				if err != nil {
					log.Fatal(err)
				}
				log.Println(gameEndTimeUTC)
			}
			datastore := new(gcloud.Datastore)
			gameEvent, exists := datastore.GetTeamGameEvent(todaysGame.GameID, team.ID)
			if exists && gameEvent.PostGameThread {
				log.Println("already found post")
				return
			}
			log.Println("making post")
			redditClient := reddit.Client{}
			redditClient.Authorize()
			log.Println("authorized")
			subreddit := "SeattleSockeye"
			title := boxscore.GetRedditPostGameThreadTitle(teamTriCode, teams)
			thingURLMapping := redditClient.GetThingURLs([]string{gameEvent.GameThreadRedditPostFullname}, subreddit)
			gameThreadURL := thingURLMapping[gameEvent.GameThreadRedditPostFullname]
			players, err := nba.GetPlayers(dailyAPIPaths.Players)
			if err != nil {
				log.Fatal(err)
			}
			playersMap := map[string]nba.Player{}
			for _, player := range players {
				playersMap[player.ID] = player
			}
			content := boxscore.GetRedditPostGameThreadBodyString(playersMap, gameThreadURL)
			submitResponse := redditClient.SubmitNewPost(subreddit, title, content)
			gameEvent.PostGameThreadRedditPostFullname = submitResponse.JsonNode.DataNode.Fullname

			gameEvent.CreatedTime = time.Now()
			gameEvent.GameID = todaysGame.GameID
			gameEvent.TeamID = team.ID
			gameEvent.PostGameThread = true
			datastore.SaveTeamGameEvent(gameEvent)
		}
	}
}
