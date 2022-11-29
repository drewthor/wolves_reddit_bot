package arena

import (
	"context"

	"github.com/drewthor/wolves_reddit_bot/api"
)

type Service interface {
	UpdateArenas(ctx context.Context, arenas []ArenaUpdate) ([]api.Arena, error)
}

func NewService(arenaStore Store) Service {
	return &service{ArenaStore: arenaStore}
}

type service struct {
	ArenaStore Store
}

func (s *service) UpdateArenas(ctx context.Context, arenas []ArenaUpdate) ([]api.Arena, error) {
	return s.ArenaStore.UpdateArenas(ctx, arenas)
}
