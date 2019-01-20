package main

import (
	"fmt"
	"github.com/drewthor/wolves_reddit_bot/apis/nba"
	"time"
)

func main() {
	currentTimeUTC := time.Now().UTC()
	fmt.Println(currentTimeUTC)
	currentTimeEastern := time.Now().UTC()
	currentDateEastern := currentTimeEastern.Format(nba.TimeDayFormat)
	dailyAPIPaths := nba.GetDailyAPIPaths()
	teams := nba.GetTeams(dailyAPIPaths.Teams)
	wolvesID := teams["OKC"].ID
	scheduledGames := nba.GetScheduledGames(dailyAPIPaths.TeamSchedule, wolvesID)
	todaysGame, gameToday := scheduledGames[currentDateEastern]
	if gameToday {
		todaysGameScoreboard := nba.GetGameScoreboard(dailyAPIPaths.Scoreboard, currentDateEastern, todaysGame.GameID)
		fmt.Println(todaysGameScoreboard)
	}
}
