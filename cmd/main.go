package main

import (
	"fmt"
	"log"
	"time"

	"github.com/drewthor/wolves_reddit_bot/apis/nba"
)

func main() {
	currentTimeUTC := time.Now().UTC()
	// Issues occur when using eastern time for "today's games" as games on the west coast can still be going on
	// when the eastern time rolls over into the next day
	westCoastLocation, locationError := time.LoadLocation("America/Los_Angeles")
	if locationError != nil {
		log.Fatal(locationError)
	}
	currentTimeWestern := currentTimeUTC.In(westCoastLocation)
	currentDateWestern := currentTimeWestern.Format(nba.TimeDayFormat)
	dailyAPIPaths := nba.GetDailyAPIPaths()
	teams := nba.GetTeams(dailyAPIPaths.Teams)
	team, foundTeam := teams[nba.MinnesotaTimberwolves]
	if !foundTeam {
		log.Println("failed to find team with TriCode: " + nba.MinnesotaTimberwolves)
		return
	}
	scheduledGames := nba.GetScheduledGames(dailyAPIPaths.TeamSchedule, team.ID)
	todaysGame, gameToday := scheduledGames[currentDateWestern]

	if gameToday {
		currentGameNumber, foundCurrentGame := scheduledGames.CurrentGameNumber(todaysGame.GameID, todaysGame.SeasonStage)
		if foundCurrentGame {
			fmt.Println(currentGameNumber)
		}
	}
	/*var wg sync.WaitGroup
	wg.Add(1)
	go gt.CreateGameThread(nba.MinnesotaTimberwolves, &wg)
	wg.Wait()*/
}
