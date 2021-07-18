package service

import (
	"github.com/drewthor/wolves_reddit_bot/api"
	"github.com/drewthor/wolves_reddit_bot/dao"
)

type ArenaService struct {
	ArenaDAO *dao.ArenaDAO
}

func (as *ArenaService) UpdateArenas(arenas []dao.ArenaUpdate) ([]api.Arena, error) {
	return as.ArenaDAO.UpdateArenas(arenas)
}
