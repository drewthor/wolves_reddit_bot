package game

import (
	"context"
	"database/sql"
	"time"

	"github.com/drewthor/wolves_reddit_bot/api"
)

type Store interface {
	GetGameWithID(ctx context.Context, id string) (api.Game, error)
	GetGamesWithIDs(ctx context.Context, ids []string) ([]api.Game, error)
	GetGameWithNBAID(ctx context.Context, id string) (api.Game, error)
	UpdateGamesOld(ctx context.Context, gameUpdates []GameUpdateOld) ([]api.Game, error)
	UpdateGames(ctx context.Context, gameUpdates []GameUpdate) ([]api.Game, error)
}

type GameUpdateOld struct {
	NBAGameID                       string
	NBAHomeTeamID                   int
	NBAAwayTeamID                   int
	HomeTeamPoints                  sql.NullInt64
	AwayTeamPoints                  sql.NullInt64
	GameStatusName                  string
	Attendance                      int
	SeasonStartYear                 string
	SeasonStageName                 string
	Period                          int
	PeriodTimeRemainingTenthSeconds int
	DurationSeconds                 sql.NullInt64
	StartTime                       time.Time
	EndTime                         sql.NullTime
}

type GameUpdate struct {
	NBAGameID                       string
	NBAHomeTeamID                   int
	NBAAwayTeamID                   int
	HomeTeamPoints                  sql.NullInt64
	AwayTeamPoints                  sql.NullInt64
	GameStatusName                  string
	NBAArenaID                      int
	Attendance                      int
	SeasonStartYear                 int
	SeasonStageName                 string
	Sellout                         bool
	Period                          int
	PeriodTimeRemainingTenthSeconds int
	DurationSeconds                 int
	RegulationPeriods               int
	StartTime                       time.Time
	EndTime                         sql.NullTime
}
