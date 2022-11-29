package schedule

import (
	"context"
	"time"

	"github.com/drewthor/wolves_reddit_bot/apis/cloudflare"
	"github.com/drewthor/wolves_reddit_bot/internal/game"
	"github.com/drewthor/wolves_reddit_bot/internal/season"
	"github.com/drewthor/wolves_reddit_bot/util"
	log "github.com/sirupsen/logrus"

	"github.com/drewthor/wolves_reddit_bot/apis/nba"

	"github.com/go-co-op/gocron"
)

type Service interface {
	Start()
	Stop()
}

type service struct {
	scheduler *gocron.Scheduler

	gameService   game.Service
	seasonService season.Service

	r2Client cloudflare.Client
}

func NewService(gameService game.Service, seasonService season.Service, r2Client cloudflare.Client) Service {
	scheduler := gocron.NewScheduler(time.UTC)

	scheduler.TagsUnique()

	return &service{
		scheduler:     scheduler,
		gameService:   gameService,
		seasonService: seasonService,
		r2Client:      r2Client,
	}
}

func (s *service) Start() {
	s.scheduler.Every(5).Minutes().Do(s.getTodaysGamesAndAddToJobs)
	// 11am UTC or 3/4 am LA time
	s.scheduler.Every(1).Day().At("11:00").Do(s.updateSeasonStartYear)
	s.scheduler.Every(1).Day().At("11:00").Do(s.updateSeasonWeeks)

	s.scheduler.StartAsync()
}

func (s *service) Stop() {
	s.scheduler.Stop()
}

func (s *service) updateSeasonStartYear() {
	ctx := context.Background()
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

	seasonStartYear, err := s.seasonService.GetCurrentSeasonStartYear(ctx)
	if err != nil {
		log.WithError(err).Error("failed to get current season start year")
	} else {
		nbaCurrentSeasonStartYear = seasonStartYear
	}

	nba.SetCurrentSeasonStartYear(nbaCurrentSeasonStartYear)
}

func (s *service) updateSeasonWeeks() {
	ctx := context.Background()

	_, err := s.seasonService.UpdateSeasonWeeks(ctx)
	if err != nil {
		log.WithError(err).Error("failed to update season weeks during scheduled job")
	}
}

func (s *service) getTodaysGamesAndAddToJobs() {
	ctx := context.Background()
	todaysGameDate := ""
	currentTimeUTC := time.Now().UTC()
	westCoastLocation, err := time.LoadLocation("Pacific/Honolulu")
	if err != nil {
		log.WithError(err).Error("failed to load west coast location for todays game date")
	} else {
		currentTimeWestern := currentTimeUTC.In(westCoastLocation)
		todaysGameDate = currentTimeWestern.Format(nba.TimeDayFormat)
	}

	todaysScoreboard, err := nba.GetTodaysScoreboard(ctx, s.r2Client, util.NBAR2Bucket)
	if err != nil {
		log.WithError(err).Error("failed to get new todays scoreboard to add games to jobs")
	}
	todaysGameDate = nba.NormalizeGameDate(todaysScoreboard.Scoreboard.GameDate)
	// ran into problem where nba didn't update old scoreboard for todays games 2022-09-30 they are still publishing 2022-09-29
	// even though the warriors were playing washington in japan
	//if todaysScoreboard.Scoreboard.GameDate != "" {
	//	todaysGameDate = nba.NormalizeGameDate(todaysScoreboard.Scoreboard.GameDate)
	//}
	//todaysScoreboardOld, err := nba.GetGameScoreboards(ctx, s.r2Client, util.NBAR2Bucket, todaysGameDate)
	//if err != nil {
	//	currentUser, userErr := user.Current()
	//	log.WithError(err).WithFields(log.Fields{"user": currentUser, "user_err": userErr}).Warn("failed to get old todays scoreboard to add games to jobs")
	//}

	uniqueGameIDStartTimeUTCMap := map[string]time.Time{}
	for _, g := range todaysScoreboard.Scoreboard.Games {
		uniqueGameIDStartTimeUTCMap[g.GameID] = g.GameTimeUTC
	}
	//for _, g := range todaysScoreboardOld.Games {
	//	uniqueGameIDStartTimeUTCMap[g.ID] = g.StartTimeUTC
	//}

	for gameID, startTimeUTC := range uniqueGameIDStartTimeUTCMap {
		// check if game is already saved and ended
		g, err := s.gameService.GetGameWithNBAID(ctx, gameID)
		if err == nil && game.GameStatus(g.Status) == game.GameStatusCompleted {
			// already have game saved
			continue
		}

		tag := gameID
		jobs, err := s.scheduler.FindJobsByTag(tag)
		if err == nil && len(jobs) == 1 {
			// job already exists; just update the start time in case it has changed
			job := jobs[0]
			s.scheduler.Job(job).StartAt(startTimeUTC).Update()
			continue
		}

		if len(jobs) > 0 {
			log.WithError(err).WithFields(log.Fields{"num_jobs": len(jobs)}).Warn("jobs found for gameID but creating as if new")
		}

		log.Infof("creating job for gameID: %s", gameID)

		// job does not exist so create it
		_, err = s.scheduler.Every(30).Seconds().StartAt(startTimeUTC).Tag(tag).Do(s.getGameData, ctx, gameID, todaysGameDate)
		if err != nil {
			log.WithError(err).Error("error scheduling job to get game data")
		}
	}
}

func (s *service) getGameData(ctx context.Context, gameID, gameDate string) {
	jobs := []gocron.Job{}
	for _, job := range s.scheduler.Jobs() {
		jobs = append(jobs, *job)
	}
	log.Println("Getting game data for game: ", gameID, " for gameDate: ", gameDate)

	g, err := s.gameService.UpdateGame(ctx, gameID, gameDate, nba.NBACurrentSeasonStartYear)
	if err != nil {
		log.WithError(err).Warn("failed to update games via getGameData")
	}

	if g.EndTime != nil {
		jobs := []gocron.Job{}
		for _, job := range s.scheduler.Jobs() {
			jobs = append(jobs, *job)
		}
		log.WithField("jobs", jobs).Println("scheduler length before removing: ", s.scheduler.Len())
		log.Println("removing scheduled job with tag: ", gameID)
		log.WithField("jobs", s.scheduler.Jobs()).Println("scheduler length after removing: ", s.scheduler.Len())
		err = s.scheduler.RemoveByTag(gameID)
		if err != nil {
			log.Println("could not remove scheduled job with tag: ", gameID, " error: ", err)
		}
	}
}
