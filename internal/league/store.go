package league

import (
	"context"
)

type LeagueUpdate struct {
	Name string
}

type League struct {
	ID   string
	Name string
}

type Store interface {
	UpdateLeagues(ctx context.Context, leagueUpdates []LeagueUpdate) ([]League, error)
}
