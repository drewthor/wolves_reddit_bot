package pgt

import (
	"github.com/drewthor/wolves_reddit_bot/apis/nba"
	"github.com/drewthor/wolves_reddit_bot/apis/reddit"
	"log"
	"time"
)

func CreatePostGameThread() {
	currentTimeUTC := time.Now().UTC()
	eastCoastLocation, locationError := time.LoadLocation("America/New_York")
	if locationError != nil {
		log.Fatal(locationError)
	}
	currentTimeEastern := currentTimeUTC.In(eastCoastLocation)
	currentDateEastern := currentTimeEastern.Format(nba.TimeDayFormat)
	dailyAPIPaths := nba.GetDailyAPIPaths()
	teams := nba.GetTeams(dailyAPIPaths.Teams)
	wolvesID := teams["MIN"].ID
	scheduledGames := nba.GetScheduledGames(dailyAPIPaths.TeamSchedule, wolvesID)
	todaysGame, gameToday := scheduledGames[currentDateEastern]
	if gameToday {
		todaysGameScoreboard := nba.GetGameScoreboard(dailyAPIPaths.Scoreboard, currentDateEastern, todaysGame.GameID)
		if todaysGameScoreboard.EndTimeUTC != "" {
			gameEndTime, err := time.Parse(nba.UTCFormat, todaysGameScoreboard.EndTimeUTC)
			if err != nil {
				log.Fatal(err)
			}
			timeSinceGameEnded := currentTimeUTC.Sub(gameEndTime)
			if timeSinceGameEnded.Minutes() < 2 {
				redditClient := reddit.Client{}
				redditClient.Authorize()
				subreddit := "SeattleSockeye"
				title := "test"
				boxscore := nba.GetBoxscore(dailyAPIPaths.Boxscore, currentDateEastern, todaysGame.GameID)
				content := boxscore.GetRedditBodyString(nba.GetPlayers(dailyAPIPaths.Players))
				redditClient.SubmitNewPost(subreddit, title, content)
			}
		}
	}
}
