package playbyplay

import (
	"context"

	"github.com/drewthor/wolves_reddit_bot/api"
)

type PlayByPlayWriter interface {
	UpdatePlayByPlays(ctx context.Context, playByPlayUpdates []PlayByPlayUpdate) ([]api.PlayByPlay, error)
}

type PlayByPlayUpdate struct {
}