package service

import (
	"strings"

	"github.com/drewthor/wolves_reddit_bot/api"
	"github.com/drewthor/wolves_reddit_bot/apis/nba"
	"github.com/drewthor/wolves_reddit_bot/dao"
)

type TeamSeasonService struct {
	TeamSeasonDao *dao.TeamSeasonDAO
}

func (tss TeamSeasonService) UpdateTeamSeasons(teamIDs map[string]nba.Team, seasonStartYear int) ([]api.TeamSeason, error) {
	teamSeasonUpdates := []dao.TeamSeasonUpdate{}
	for teamID, nbaTeam := range teamIDs {
		league := "nba"
		if !nbaTeam.IsNBAFranchise {
			league = "international"
		}

		teamSeasonUpdates = append(teamSeasonUpdates, dao.TeamSeasonUpdate{
			TeamID:          teamID,
			LeagueName:      league,
			SeasonStartYear: seasonStartYear,
			ConferenceName:  strings.ToLower(nbaTeam.Conference),
			DivisionName:    strings.ToLower(nbaTeam.Division),
		})
	}

	updatedTeamSeasons, err := tss.TeamSeasonDao.UpdateTeamSeasons(teamSeasonUpdates)
	if err != nil {
		return nil, err
	}

	return updatedTeamSeasons, nil
}
