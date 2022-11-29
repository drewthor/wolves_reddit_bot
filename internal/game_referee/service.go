package game_referee

import "context"

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
	_, err := s.GameRefereeStore.UpdateGameReferees(ctx, gameRefereeUpdates)
	return err
}
