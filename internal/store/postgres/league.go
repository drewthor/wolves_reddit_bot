package postgres

import (
	"context"

	"github.com/drewthor/wolves_reddit_bot/internal/league"
)

func (d DB) UpdateLeagues(ctx context.Context, leagueUpdates []league.LeagueUpdate) ([]league.League, error) {
	return nil, nil
}
