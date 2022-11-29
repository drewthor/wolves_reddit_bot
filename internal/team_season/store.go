package team_season

import (
	"context"

	"github.com/drewthor/wolves_reddit_bot/api"
)

type Store interface {
	GetTeamSeasonsWithIDs(ctx context.Context, teamSeasonIDs []string) ([]api.TeamSeason, error)
	UpdateTeamSeasons(ctx context.Context, teamSeasonUpdates []TeamSeasonUpdate) ([]api.TeamSeason, error)
}

type TeamSeasonUpdate struct {
	TeamID          string
	LeagueName      string
	SeasonStartYear int
	ConferenceName  string
	DivisionName    string
}
