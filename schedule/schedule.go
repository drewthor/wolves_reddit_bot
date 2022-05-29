package schedule

import (
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/drewthor/wolves_reddit_bot/apis/nba"

	"github.com/go-co-op/gocron"
)

func NewScheduler() *gocron.Scheduler {
	s := gocron.NewScheduler(time.UTC)

	s.TagsUnique()

	s.Every(5).Minutes().Do(getTodaysGamesAndAddToJobs, s)
	// 11am UTC or 3/4 am LA time
	s.Every(1).Day().At("11:00").Do(updateSeasonStartYear)

	return s
}

func updateSeasonStartYear() {
	currentTime := time.Now()
	nbaCurrentSeasonStartYear := currentTime.Year()
	if currentTime.Month() < time.July {
		nbaCurrentSeasonStartYear = currentTime.Year() - 1
	}
	dailyAPIPaths, err := nba.GetDailyAPIPaths()
	if err == nil {
		nbaCurrentSeasonStartYear = dailyAPIPaths.APISeasonInfoNode.SeasonYear
	} else {
		log.WithError(err).Error("could not get daily API to set current NBA season start year; falling back to calendar year")
	}
	nba.SetCurrentSeasonStartYear(nbaCurrentSeasonStartYear)
}

func getTodaysGamesAndAddToJobs(s *gocron.Scheduler) {
	todaysScoreboard, err := nba.GetTodaysScoreboard()
	if err != nil {
		log.WithError(err).Error("cannot run job to get todays games and add them as jobs")
		return
	}

	for _, game := range todaysScoreboard.Scoreboard.Games {
		s.Every(1).Minutes().StartAt(game.GameTimeUTC).Tag(game.GameID).Do(getGameData, s, game.GameID, todaysScoreboard.Scoreboard.GameDate)
	}
}

func getGameData(s *gocron.Scheduler, gameID, gameDate string) {
	log.Println("Getting game data for game: ", gameID, " for gameDate: ", gameDate)
	boxscore, err := nba.GetBoxscoreDetailed(gameID, nba.NBACurrentSeasonStartYear)
	if err != nil {
		log.WithError(err).Errorf("could not retrieve detailed boxscore for gameID: %s", gameID)
	}
	_, err = nba.GetOldBoxscore(gameID, gameDate, time.Now().Year())
	if err != nil {
		log.WithError(err).Errorf("could not retrieve old boxscore for gameID: %s", gameID)
	}

	if boxscore.Final() {
		log.Println("scheduler length before removing: ", s.Len())
		log.Println("removing scheduled job with tag: ", gameID)
		log.Println("scheduler length after removing: ", s.Len())
		err = s.RemoveByTag(gameID)
		if err != nil {
			log.Println("could not remove scheduled job with tag: ", gameID, " error: ", err)
		}
	}
}
