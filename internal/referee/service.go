package referee

import (
	"github.com/drewthor/wolves_reddit_bot/api"
)

type Service interface {
	UpdateReferees(refereeUpdates []RefereeUpdate) ([]api.Referee, error)
}

func NewService(refereeStore Store) Service {
	return &service{RefereeStore: refereeStore}
}

type service struct {
	RefereeStore Store
}

func (s service) UpdateReferees(refereeUpdates []RefereeUpdate) ([]api.Referee, error) {
	return s.RefereeStore.UpdateReferees(refereeUpdates)
}
