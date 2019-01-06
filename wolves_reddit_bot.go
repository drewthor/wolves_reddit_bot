package main

import (
	"fmt"
	"github.com/aturn3/wolves_reddit_bot/apis/nba"
)

func main() {
	teams := nba.GetTeams()
	wolvesID := teams["MIN"].ID
	scheduledGames := nba.GetScheduledGames(wolvesID)
	fmt.Println(scheduledGames)
}
