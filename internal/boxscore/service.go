package boxscore

import (
	"github.com/drewthor/wolves_reddit_bot/api"
	"github.com/drewthor/wolves_reddit_bot/apis/nba"
)

type Service interface {
	Get(gameID, gameDate string) (api.Boxscore, error)
}

func NewService() Service {
	return &service{}
}

type service struct{}

func (s *service) Get(gameID, gameDate string) (api.Boxscore, error) {
	boxscore, err := nba.GetCurrentSeasonBoxscore(gameID, gameDate)
	return api.Boxscore{Boxscore: boxscore}, err
}
