package team_season

import (
	"context"
	"strings"

	"github.com/drewthor/wolves_reddit_bot/api"
	"github.com/drewthor/wolves_reddit_bot/apis/nba"
)

type Service interface {
	UpdateTeamSeasons(ctx context.Context, teamIDs map[string]nba.Team, seasonStartYear int) ([]api.TeamSeason, error)
}

func NewService(teamSeasonStore Store) Service {
	return &service{TeamSeasonStore: teamSeasonStore}
}

type service struct {
	TeamSeasonStore Store
}

func (s *service) UpdateTeamSeasons(ctx context.Context, teamIDs map[string]nba.Team, seasonStartYear int) ([]api.TeamSeason, error) {
	teamSeasonUpdates := []TeamSeasonUpdate{}
	for teamID, nbaTeam := range teamIDs {
		league := "nba"
		if !nbaTeam.IsNBAFranchise {
			league = "international"
		}

		teamSeasonUpdates = append(teamSeasonUpdates, TeamSeasonUpdate{
			TeamID:          teamID,
			LeagueName:      league,
			SeasonStartYear: seasonStartYear,
			ConferenceName:  strings.ToLower(nbaTeam.Conference),
			DivisionName:    strings.ToLower(nbaTeam.Division),
		})
	}

	updatedTeamSeasons, err := s.TeamSeasonStore.UpdateTeamSeasons(ctx, teamSeasonUpdates)
	if err != nil {
		return nil, err
	}

	return updatedTeamSeasons, nil
}
