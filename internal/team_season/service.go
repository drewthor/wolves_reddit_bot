package team_season

import (
	"context"
	"fmt"

	"github.com/drewthor/wolves_reddit_bot/api"
	"github.com/drewthor/wolves_reddit_bot/apis/nba"
	"go.opentelemetry.io/otel"
)

type Service interface {
	UpdateFranchiseTeamSeasons(ctx context.Context, nbaLeagueID string, nbaFranchises []nba.Franchise) ([]api.TeamSeason, error)
}

func NewService(teamSeasonStore Store, nbaClient nba.Client) Service {
	return &service{teamSeasonStore: teamSeasonStore, nbaClient: nbaClient}
}

type service struct {
	teamSeasonStore Store

	nbaClient nba.Client
}

func (s *service) UpdateFranchiseTeamSeasons(ctx context.Context, nbaLeagueID string, nbaFranchises []nba.Franchise) ([]api.TeamSeason, error) {
	ctx, span := otel.Tracer("team_season").Start(ctx, "team_season.service.UpdateFranchiseTeamSeasons")
	defer span.End()

	// update team seasons
	minSeasonStartYear := 0
	maxSeasonStartYear := 0
	for _, updatedFranchise := range nbaFranchises {
		minSeasonStartYear = min(minSeasonStartYear, updatedFranchise.StartYear)
		maxSeasonStartYear = max(maxSeasonStartYear, updatedFranchise.EndYear)
	}

	seasonFetched := make(map[int]bool)
	teamIDSeasonStartYearStanding := make(map[int]map[int]nba.TeamStanding)

	teamSeasonUpdates := []TeamSeasonUpdate{}
	for _, nbaFranchise := range nbaFranchises {
		for _, teamSeason := range nbaFranchise.TeamSeasons {
			if !seasonFetched[teamSeason.Year] {
				teamStandings, err := s.nbaClient.TeamStandings(ctx, nbaLeagueID, teamSeason.Year, nba.SeasonTypeRegular)
				if err != nil {
					return nil, fmt.Errorf("failed to get team standings updating franchise team seasons: %w", err)
				}

				for _, teamStanding := range teamStandings {
					if _, ok := teamIDSeasonStartYearStanding[teamStanding.TeamID]; !ok {
						teamIDSeasonStartYearStanding[teamStanding.TeamID] = make(map[int]nba.TeamStanding)
					}

					teamIDSeasonStartYearStanding[teamStanding.TeamID][teamStanding.SeasonStartYear] = teamStanding
				}
			}

			teamStanding := teamIDSeasonStartYearStanding[nbaFranchise.TeamID][teamSeason.Year]

			teamSeasonUpdates = append(teamSeasonUpdates, TeamSeasonUpdate{
				NBALeagueID:     nbaLeagueID,
				NBATeamID:       nbaFranchise.TeamID,
				SeasonStartYear: teamSeason.Year,
				ConferenceName:  teamStanding.Conference,
				DivisionName:    teamStanding.Division,
				Name:            teamSeason.Name,
				City:            teamSeason.City,
			})
		}
	}

	updatedTeamSeasons, err := s.teamSeasonStore.UpdateTeamSeasons(ctx, teamSeasonUpdates)
	if err != nil {
		return nil, err
	}

	return updatedTeamSeasons, nil
}
