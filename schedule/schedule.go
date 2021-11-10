package schedule

import (
	"log"
	"time"

	"github.com/drewthor/wolves_reddit_bot/apis/nba"

	"github.com/go-co-op/gocron"
)

func NewScheduler() *gocron.Scheduler {
	s := gocron.NewScheduler(time.UTC)

	s.TagsUnique()

	s.Every(5).Minutes().Do(getTodaysGamesAndAddToJobs, s)

	return s
}

func getTodaysGamesAndAddToJobs(s *gocron.Scheduler) {
	todaysScoreboard, err := nba.GetTodaysScoreboard()
	if err != nil {
		log.Println("cannot run job to get todays games and add them as jobs", err)
		return
	}

	for _, game := range todaysScoreboard.Scoreboard.Games {
		s.Every(1).Minutes().StartAt(game.GameTimeUTC).Tag(game.GameID).Do(getGameData, s, game.GameID, todaysScoreboard.Scoreboard.GameDate)
	}
}

func getGameData(s *gocron.Scheduler, gameID, gameDate string) {
	boxscore, err := nba.GetBoxscoreDetailed(gameID, time.Now().Year())
	if err != nil {
		log.Printf("could not retrieve detailed boxscore for gameID: %s\n", gameID)
	}
	_, err = nba.GetOldBoxscore(gameID, gameDate, time.Now().Year())
	if err != nil {
		log.Printf("could not retrieve old boxscore for gameID: %s\n", gameID)
	}

	if boxscore.Final() {
		s.RemoveByTag(gameID)
	}
}
