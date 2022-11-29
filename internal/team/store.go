package team

import (
	"context"

	"github.com/drewthor/wolves_reddit_bot/api"
)

type Store interface {
	GetTeamWithID(ctx context.Context, teamID string) (api.Team, error)
	GetTeamsWithIDs(ctx context.Context, ids []string) ([]api.Team, error)
	ListTeams(ctx context.Context) ([]api.Team, error)
	UpdateTeams(ctx context.Context, teams []TeamUpdate) ([]api.Team, error)
	NBATeamIDMappings(ctx context.Context) (map[string]string, error)
}

type TeamUpdate struct {
	Name          string
	Nickname      string
	City          string
	AlternateCity string
	NBAURLName    string
	NBAShortName  string
	NBATeamID     int
}
