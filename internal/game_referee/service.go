package game_referee

import (
	"context"

	"go.opentelemetry.io/otel"
)

type Service interface {
	UpdateGameReferees(ctx context.Context, gameRefereeUpdates []GameRefereeUpdate) error
}

func NewService(gameRefereeStore Store) Service {
	return &service{GameRefereeStore: gameRefereeStore}
}

type service struct {
	GameRefereeStore Store
}

func (s *service) UpdateGameReferees(ctx context.Context, gameRefereeUpdates []GameRefereeUpdate) error {
	ctx, span := otel.Tracer("game_referee").Start(ctx, "boxscore.service.UpdateGameReferees")
	defer span.End()

	_, err := s.GameRefereeStore.UpdateGameReferees(ctx, gameRefereeUpdates)
	return err
}
