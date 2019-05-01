package gfunctions

import (
	"github.com/drewthor/wolves_reddit_bot/apis/nba"
	"github.com/drewthor/wolves_reddit_bot/apis/reddit"
	"github.com/drewthor/wolves_reddit_bot/pkg/gcloud"
	"log"
	"time"
)

// Work in progress
func CreateGameThread(teamTriCode nba.TriCode) {
	currentTimeUTC := time.Now().UTC()
	eastCoastLocation, locationError := time.LoadLocation("America/New_York")
	if locationError != nil {
		log.Fatal(locationError)
	}
	currentTimeWestern := currentTimeUTC.In(eastCoastLocation)
	currentDateWestern := currentTimeWestern.Format(nba.TimeDayFormat)
	dailyAPIPaths := nba.GetDailyAPIPaths()
	teams := nba.GetTeams(dailyAPIPaths.Teams)
	teamID := teams[teamTriCode].ID
	scheduledGames := nba.GetScheduledGames(dailyAPIPaths.TeamSchedule, teamID)
	todaysGame, gameToday := scheduledGames[currentDateWestern]
	if gameToday {
		log.Println("game today")
		todaysGameScoreboard := nba.GetGameScoreboard(dailyAPIPaths.Scoreboard, currentDateWestern, todaysGame.GameID)
		boxscore := nba.GetBoxscore(dailyAPIPaths.Boxscore, currentDateWestern, todaysGame.GameID)
		if boxscore.GameEnded() {
			log.Println("game ended")
			if todaysGameScoreboard.EndTimeUTC != "" {
				gameEndTimeUTC, err := time.Parse(nba.UTCFormat, todaysGameScoreboard.EndTimeUTC)
				if err != nil {
					log.Fatal(err)
				}
				log.Println(gameEndTimeUTC)
			}
			datastore := new(gcloud.Datastore)
			gameEvent, exists := datastore.GetTeamGameEvent(todaysGame.GameID, teamID)
			if exists && gameEvent.PostGameThread {
				log.Println("already found post")
				return
			}
			log.Println("making post")
			redditClient := reddit.Client{}
			redditClient.Authorize()
			subreddit := "SeattleSockeye"
			title := boxscore.GetRedditPostGameThreadTitle(teamTriCode, teams)
			content := boxscore.GetRedditBodyString(nba.GetPlayers(dailyAPIPaths.Players))
			redditClient.SubmitNewPost(subreddit, title, content)
			gameEvent.GameID = todaysGame.GameID
			gameEvent.PostGameThread = true
			datastore.SaveTeamGameEvent(gameEvent)
		}
	}
}
