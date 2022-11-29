package postgres

import (
	"context"

	"github.com/drewthor/wolves_reddit_bot/internal/season"
)

func (d DB) UpdateSeasons(ctx context.Context, seasonUpdates []season.SeasonUpdate) ([]season.Season, error) {
	return nil, nil
}
