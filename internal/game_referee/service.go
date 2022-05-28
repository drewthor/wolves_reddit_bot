package game_referee

type Service interface {
	UpdateGameReferees(gameRefereeUpdates []GameRefereeUpdate) error
}

func NewService(gameRefereeStore Store) Service {
	return &service{GameRefereeStore: gameRefereeStore}
}

type service struct {
	GameRefereeStore Store
}

func (s *service) UpdateGameReferees(gameRefereeUpdates []GameRefereeUpdate) error {
	_, err := s.GameRefereeStore.UpdateGameReferees(gameRefereeUpdates)
	return err
}
