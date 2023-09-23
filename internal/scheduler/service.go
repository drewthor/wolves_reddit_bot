package scheduler

import (
	"context"
	"log/slog"
	"time"

	"github.com/drewthor/wolves_reddit_bot/apis/cloudflare"
	"github.com/drewthor/wolves_reddit_bot/internal/game"
	"github.com/drewthor/wolves_reddit_bot/internal/season"
	"github.com/drewthor/wolves_reddit_bot/util"
	"go.opentelemetry.io/otel"

	"github.com/drewthor/wolves_reddit_bot/apis/nba"

	"github.com/go-co-op/gocron"
)

type Service interface {
	Start(logger *slog.Logger)
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

func (s *service) Start(logger *slog.Logger) {
	s.scheduler.Every(5).Minutes().Do(s.getTodaysGamesAndAddToJobs, logger)
	// 11am UTC or 3/4 am LA time
	s.scheduler.Every(1).Day().At("11:00").Do(s.updateSeasonWeeks, logger)

	s.scheduler.StartAsync()
}

func (s *service) Stop() {
	s.scheduler.Stop()
}

func (s *service) updateSeasonWeeks(logger *slog.Logger) {
	ctx := context.Background()

	_, err := s.seasonService.UpdateSeasonWeeks(ctx)
	if err != nil {
		logger.Error("failed to update season weeks during scheduled job", slog.Any("error", err))
	}
}

func (s *service) getTodaysGamesAndAddToJobs(logger *slog.Logger) {
	ctx := context.Background()
	ctx, span := otel.Tracer("scheduler").Start(ctx, "scheduler.service.getTodaysGamesAndAddToJobs")
	defer span.End()

	todaysScoreboard, err := nba.GetTodaysScoreboard(ctx, s.r2Client, util.NBAR2Bucket)
	if err != nil {
		logger.Error("failed to get new todays scoreboard to add games to jobs", slog.Any("error", err))
	}

	uniqueGameIDStartTimeUTCMap := map[string]time.Time{}
	for _, g := range todaysScoreboard.Scoreboard.Games {
		uniqueGameIDStartTimeUTCMap[g.GameID] = g.GameTimeUTC
	}

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
			slog.Warn("jobs found for gameID but creating as if new", slog.Any("error", err), slog.Int("num_jobs", len(jobs)))
		}

		slog.Info("creating job", slog.String("gameID", gameID))

		// job does not exist so create it
		_, err = s.scheduler.Every(30).Seconds().StartAt(startTimeUTC).Tag(tag).Do(s.updateGame, logger, gameID)
		if err != nil {
			logger.Error("error scheduling job to get game data", slog.Any("error", err))
		}
	}
}

func (s *service) updateGame(logger *slog.Logger, gameID string) {
	ctx := context.Background()
	ctx, span := otel.Tracer("scheduler").Start(ctx, "scheduler.service.updateGame")
	defer span.End()

	logger = logger.With(slog.String("game_id", gameID))

	jobs := []gocron.Job{}
	for _, job := range s.scheduler.Jobs() {
		jobs = append(jobs, *job)
	}

	seasonStartYear, err := s.seasonService.GetCurrentSeasonStartYear(ctx)
	if err != nil {
		logger.Error("failed to get game current season start year when updating game", slog.Any("error", err))
		return
	}

	g, err := s.gameService.UpdateGame(ctx, logger, gameID, seasonStartYear)
	if err != nil {
		logger.Error("failed to update game via updateGame", slog.Int("season_start_year", seasonStartYear))
		return
	}

	if g.EndTime != nil {
		jobs := []gocron.Job{}
		for _, job := range s.scheduler.Jobs() {
			jobs = append(jobs, *job)
		}
		err = s.scheduler.RemoveByTag(gameID)
		if err != nil {
			slog.Error("could not remove scheduled job", slog.String("tag", gameID), slog.Any("error", err))
		}
	}
}
