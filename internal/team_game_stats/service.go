package team_game_stats

type Service interface {
	UpdateTeamGameStatsTotals(teamGameStatsTotalsUpdates []TeamGameStatsTotalUpdate) ([]TeamGameStatsTotal, error)
}

func NewService(teamGameStatsStore Store) Service {
	return &service{TeamGameStatsStore: teamGameStatsStore}
}

type service struct {
	TeamGameStatsStore Store
}

func (s service) UpdateTeamGameStatsTotals(teamGameStatsTotalsUpdates []TeamGameStatsTotalUpdate) ([]TeamGameStatsTotal, error) {
	return s.TeamGameStatsStore.UpdateTeamGameStatsTotals(teamGameStatsTotalsUpdates)
}
