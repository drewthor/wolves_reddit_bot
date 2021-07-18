package service

import "github.com/drewthor/wolves_reddit_bot/dao"

type GameRefereeService struct {
	GameRefereeDAO *dao.GameRefereeDAO
}

func (grs GameRefereeService) UpdateGameReferees(gameRefereeUpdates []dao.GameRefereeUpdate) error {
	_, err := grs.GameRefereeDAO.UpdateGameReferees(gameRefereeUpdates)
	return err
}
