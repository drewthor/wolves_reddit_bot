package service

import (
	"github.com/drewthor/wolves_reddit_bot/api"
	"github.com/drewthor/wolves_reddit_bot/dao"
)

type RefereeService struct {
	RefereeDAO *dao.RefereeDAO
}

func (rs RefereeService) UpdateReferees(refereeUpdates []dao.RefereeUpdate) ([]api.Referee, error) {
	return rs.RefereeDAO.UpdateReferees(refereeUpdates)
}
