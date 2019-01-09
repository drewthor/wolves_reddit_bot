package main

import (
	"fmt"
	"github.com/drewthor/wolves_reddit_bot/apis/nba"
	"time"
)

func main() {
	currentTime := time.Now()
	currentDate := currentTime.Format(nba.TimeDayFormat)
	dailyAPIPaths := nba.GetDailyAPIPaths()
	teams := nba.GetTeams(dailyAPIPaths.Teams)
	wolvesID := teams["MIN"].ID
	scheduledGames := nba.GetScheduledGames(dailyAPIPaths.TeamSchedule, wolvesID)
	todaysGame, gameToday := scheduledGames[currentDate]
	if gameToday {
		todaysGameID := todaysGame.GameID
		fmt.Println(todaysGameID)
	}
}
