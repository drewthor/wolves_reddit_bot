package player

import (
	"context"

	"github.com/drewthor/wolves_reddit_bot/api"
)

type Store interface {
	GetPlayerWithID(ctx context.Context, playerID string) (api.Player, error)
	ListPlayers(ctx context.Context) ([]api.Player, error)
	GetPlayersWithIDs(ctx context.Context, ids []string) ([]api.Player, error)
	UpdatePlayers(ctx context.Context, players []api.Player) ([]api.Player, error)
}
