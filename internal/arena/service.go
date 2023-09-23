package arena

import (
	"context"

	"github.com/drewthor/wolves_reddit_bot/api"
	"go.opentelemetry.io/otel"
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
	ctx, span := otel.Tracer("arena").Start(ctx, "arena.service.UpdateArenas")
	defer span.End()

	return s.ArenaStore.UpdateArenas(ctx, arenas)
}
