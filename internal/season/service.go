package season

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/drewthor/wolves_reddit_bot/apis/cloudflare"
	"github.com/drewthor/wolves_reddit_bot/apis/nba"
	"github.com/drewthor/wolves_reddit_bot/util"
	"go.opentelemetry.io/otel"
)

type Service interface {
	GetCurrentSeasonStartYear(ctx context.Context) (int, error)
	UpdateSeasonForLeague(ctx context.Context, leagueID string, seasonStartYear int) (string, error)
	UpdateSeasonWeeks(ctx context.Context) ([]SeasonWeek, error)
}

func NewService(seasonStore Store, r2Client cloudflare.Client) Service {
	return &service{seasonStore: seasonStore, r2Client: r2Client}
}

type service struct {
	seasonStore Store

	r2Client cloudflare.Client
}

func (s service) GetCurrentSeasonStartYear(ctx context.Context) (int, error) {
	ctx, span := otel.Tracer("season").Start(ctx, "season.service.GetCurrentSeasonStartYear")
	defer span.End()

	//leagueSchedule, err := nba.GetLeagueSchedule(ctx, s.r2Client, util.NBAR2Bucket)
	//if err != nil {
	//	return 0, fmt.Errorf("failed to get current season start year: %w", err)
	//}
	//
	//seasonYear, err := strconv.Atoi(strings.Split(leagueSchedule.LeagueSchedule.SeasonYear, "-")[0])
	//if err != nil {
	//	return 0, fmt.Errorf("failed to convert season start year to int when getting current season start year: %w", err)
	//}
	// TODO: don't make this static
	now := time.Now().UTC()
	seasonStartYear := now.Year()
	if now.Month() <= time.September {
		seasonStartYear--
	}

	return seasonStartYear, nil
}

func (s service) UpdateSeasonForLeague(ctx context.Context, leagueID string, seasonStartYear int) (string, error) {
	ctx, span := otel.Tracer("season").Start(ctx, "season.service.UpdateSeasonForLeague")
	defer span.End()

	return "", nil
}

func (s service) UpdateSeasonWeeks(ctx context.Context) ([]SeasonWeek, error) {
	ctx, span := otel.Tracer("season").Start(ctx, "season.service.UpdateSeasonWeeks")
	defer span.End()

	// TODO don't make this static
	seasonStartYear, err := s.GetCurrentSeasonStartYear(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get current season start year when updating season weeks: %w", err)
	}
	leagueSchedule, err := nba.GetLeagueSchedule(ctx, s.r2Client, util.NBAR2Bucket, seasonStartYear)
	if err != nil {
		return nil, fmt.Errorf("failed to update season weeks: %w", err)
	}

	seasonYear, err := strconv.Atoi(strings.Split(leagueSchedule.LeagueSchedule.SeasonYear, "-")[0])
	if err != nil {
		return nil, fmt.Errorf("failed to convert season year to int when updating season weeks: %w", err)
	}

	var seasonWeekUpdates []SeasonWeekUpdate
	for _, week := range leagueSchedule.LeagueSchedule.Weeks {
		// add a day to the end date as nba uses non-overlapping dates e.g. 10-14 -> 10:20, 10-21 -> 10-27
		seasonWeekUpdates = append(seasonWeekUpdates, SeasonWeekUpdate{SeasonStartYear: seasonYear, StartDate: week.StartDate, EndDate: week.EndDate.AddDate(0, 0, 1)})
	}

	return s.seasonStore.UpdateSeasonWeeks(ctx, seasonWeekUpdates)
}
