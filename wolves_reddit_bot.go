package main

import (
	"fmt"
	"github.com/aturn3/wolves_reddit_bot/apis/nba"
)

func main() {
	dailyAPIPaths := nba.GetDailyAPIPaths()
	teams := nba.GetTeams(dailyAPIPaths.Teams)
	wolvesID := teams["MIN"].ID
	scheduledGames := nba.GetScheduledGames(dailyAPIPaths.TeamSchedule, wolvesID)
	firstGameID := scheduledGames["20180929"].GameID
	fmt.Println(scheduledGames)
	fmt.Println(firstGameID)
}
