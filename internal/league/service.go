package league

import "context"

type Service interface {
	UpdateLeagues(ctx context.Context, leagueUpdates []LeagueUpdate) ([]League, error)
}

func NewService(leagueStore Store) Service {
	return &service{leagueStore: leagueStore}
}

type service struct {
	leagueStore Store
}

func (s *service) UpdateLeagues(ctx context.Context, leagueUpdates []LeagueUpdate) ([]League, error) {
	return s.leagueStore.UpdateLeagues(ctx, leagueUpdates)
}
