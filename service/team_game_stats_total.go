package service

import "github.com/drewthor/wolves_reddit_bot/dao"

type TeamGameStatsTotalService struct {
	TeamGameStatsTotalDAO *dao.TeamGameStatsTotalDAO
}

func (tgsts TeamGameStatsTotalService) UpdateTeamGameStatsTotals(teamGameStatsTotalsUpdates []dao.TeamGameStatsTotalUpdate) ([]dao.TeamGameStatsTotal, error) {
	return tgsts.TeamGameStatsTotalDAO.UpdateTeamGameStatsTotals(teamGameStatsTotalsUpdates)
}
