package pgt

import (
	"fmt"
	"github.com/drewthor/wolves_reddit_bot/apis/nba"
	"github.com/drewthor/wolves_reddit_bot/apis/reddit"
	"log"
	"time"
)

func CreatePostGameThread(teamTriCode nba.TriCode) {
	currentTimeUTC := time.Now().UTC()
	eastCoastLocation, locationError := time.LoadLocation("America/New_York")
	if locationError != nil {
		log.Fatal(locationError)
	}
	currentTimeEastern := currentTimeUTC.In(eastCoastLocation)
	currentDateEastern := currentTimeEastern.Format(nba.TimeDayFormat)
	dailyAPIPaths := nba.GetDailyAPIPaths()
	teams := nba.GetTeams(dailyAPIPaths.Teams)
	teamID := teams[teamTriCode].ID
	scheduledGames := nba.GetScheduledGames(dailyAPIPaths.TeamSchedule, teamID)
	todaysGame, gameToday := scheduledGames[currentDateEastern]
	if gameToday {
		log.Println("game today")
		todaysGameScoreboard := nba.GetGameScoreboard(dailyAPIPaths.Scoreboard, currentDateEastern, todaysGame.GameID)
		if todaysGameScoreboard.EndTimeUTC != "" {
			log.Print("current time: ")
			log.Print(currentTimeUTC)
			gameEndTimeUTC, err := time.Parse(nba.UTCFormat, todaysGameScoreboard.EndTimeUTC)
			if err != nil {
				log.Fatal(err)
			}
			log.Print("gameEndTime: ")
			log.Print(gameEndTimeUTC)
			timeSinceGameEnded := currentTimeUTC.Sub(gameEndTimeUTC)
			log.Println(fmt.Sprintf("timeSinceGameEnded: %fmin %fsec", timeSinceGameEnded.Minutes(), timeSinceGameEnded.Seconds()))
			if timeSinceGameEnded.Minutes() < 2 {
				log.Println("making post")
				redditClient := reddit.Client{}
				redditClient.Authorize()
				subreddit := "SeattleSockeye"
				boxscore := nba.GetBoxscore(dailyAPIPaths.Boxscore, currentDateEastern, todaysGame.GameID)
				title := boxscore.GetRedditPostGameThreadTitle(teamTriCode, teams)
				content := boxscore.GetRedditBodyString(nba.GetPlayers(dailyAPIPaths.Players))
				redditClient.SubmitNewPost(subreddit, title, content)
			}
		}
	}
}
