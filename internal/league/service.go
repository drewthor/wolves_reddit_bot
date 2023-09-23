package league

import (
	"context"

	"go.opentelemetry.io/otel"
)

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
	ctx, span := otel.Tracer("league").Start(ctx, "league.service.UpdateLeagues")
	defer span.End()

	return s.leagueStore.UpdateLeagues(ctx, leagueUpdates)
}
