package team_game_stats

import "context"

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
	return s.TeamGameStatsStore.UpdateTeamGameStatsTotals(ctx, teamGameStatsTotalsUpdates)
}

func (s service) UpdateTeamGameStatsTotalsOld(ctx context.Context, teamGameStatsTotalsUpdates []TeamGameStatsTotalUpdateOld) ([]TeamGameStatsTotal, error) {
	return s.TeamGameStatsStore.UpdateTeamGameStatsTotalsOld(ctx, teamGameStatsTotalsUpdates)
}
