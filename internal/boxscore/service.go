package boxscore

import (
	"context"

	"github.com/drewthor/wolves_reddit_bot/api"
	"go.opentelemetry.io/otel"
)

type Service interface {
	Get(ctx context.Context, gameID, gameDate string) (api.Boxscore, error)
}

func NewService() Service {
	return &service{}
}

type service struct{}

func (s *service) Get(ctx context.Context, gameID, gameDate string) (api.Boxscore, error) {
	ctx, span := otel.Tracer("boxscore").Start(ctx, "boxscore.service.Get")
	defer span.End()

	//TODO: actually implement this
	return api.Boxscore{}, nil
}
