package team_game_stats

import (
	"context"

	"go.opentelemetry.io/otel"
)

type Service interface {
	UpdateTeamGameStatsTotals(ctx context.Context, teamGameStatsTotalsUpdates []TeamGameStatsTotalUpdate) ([]TeamGameStatsTotal, error)
}

func NewService(teamGameStatsStore Store) Service {
	return &service{TeamGameStatsStore: teamGameStatsStore}
}

type service struct {
	TeamGameStatsStore Store
}

func (s service) UpdateTeamGameStatsTotals(ctx context.Context, teamGameStatsTotalsUpdates []TeamGameStatsTotalUpdate) ([]TeamGameStatsTotal, error) {
	ctx, span := otel.Tracer("team_game_stats").Start(ctx, "team_game_stats.service.UpdateTeamGameStatsTotals")
	defer span.End()

	return s.TeamGameStatsStore.UpdateTeamGameStatsTotals(ctx, teamGameStatsTotalsUpdates)
}

func (s service) UpdateTeamGameStatsTotalsOld(ctx context.Context, teamGameStatsTotalsUpdates []TeamGameStatsTotalUpdateOld) ([]TeamGameStatsTotal, error) {
	ctx, span := otel.Tracer("team_game_stats").Start(ctx, "team_game_stats.service.UpdateTeamGameStatsTotalsOld")
	defer span.End()

	return s.TeamGameStatsStore.UpdateTeamGameStatsTotalsOld(ctx, teamGameStatsTotalsUpdates)
}
