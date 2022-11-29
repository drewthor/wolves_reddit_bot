package arena

import (
	"context"
	"database/sql"

	"github.com/drewthor/wolves_reddit_bot/api"
)

type ArenaUpdate struct {
	NBAArenaID int
	Name       string
	City       sql.NullString
	State      sql.NullString
	Country    string
}

type Store interface {
	UpdateArenas(ctx context.Context, arenas []ArenaUpdate) ([]api.Arena, error)
}
