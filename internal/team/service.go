package team

import (
	"context"
	"strconv"

	"github.com/drewthor/wolves_reddit_bot/api"
	"github.com/drewthor/wolves_reddit_bot/apis/nba"
	"github.com/drewthor/wolves_reddit_bot/internal/team_season"
)

type Service interface {
	Get(ctx context.Context, teamID string) (api.Team, error)
	ListPlayers(ctx context.Context) ([]api.Team, error)
	UpdateTeams(ctx context.Context, seasonStartYear int) ([]api.Team, error)
}

func NewService(teamStore Store, teamSeasonService team_season.Service) Service {
	return &service{
		TeamStore:         teamStore,
		TeamSeasonService: teamSeasonService,
	}
}

type service struct {
	TeamStore         Store
	TeamSeasonService team_season.Service
}

func (s *service) Get(ctx context.Context, teamID string) (api.Team, error) {
	team, err := s.TeamStore.GetTeamWithID(ctx, teamID)
	if err != nil {
		return api.Team{}, err
	}

	return team, err
}

func (s *service) ListPlayers(ctx context.Context) ([]api.Team, error) {
	teams, err := s.TeamStore.ListTeams(ctx)
	if err != nil {
		return nil, err
	}

	return teams, err
}

func (s *service) UpdateTeams(ctx context.Context, seasonStartYear int) ([]api.Team, error) {
	nbaTeams, err := nba.GetTeamsForSeason(seasonStartYear)
	if err != nil {
		return nil, err
	}

	nbaTeamIDTeamMappings := map[int]nba.Team{}

	teamUpdates := []TeamUpdate{}
	for _, nbaTeam := range nbaTeams {
		teamID, err := strconv.Atoi(nbaTeam.ID)
		if err != nil {
			return nil, err
		}

		teamUpdates = append(teamUpdates, TeamUpdate{
			Name:          nbaTeam.FullName,
			Nickname:      nbaTeam.Nickname,
			City:          nbaTeam.City,
			AlternateCity: nbaTeam.AlternateCity,
			NBAURLName:    nbaTeam.UrlName,
			NBATeamID:     teamID,
			NBAShortName:  nbaTeam.ShortName,
		})

		nbaTeamIDTeamMappings[teamID] = nbaTeam
	}

	teamIDNBATeamMappings := map[string]nba.Team{}

	updatedTeams, err := s.TeamStore.UpdateTeams(ctx, teamUpdates)
	if err != nil {
		return nil, err
	}

	for _, updatedTeam := range updatedTeams {
		teamIDNBATeamMappings[updatedTeam.ID] = nbaTeamIDTeamMappings[updatedTeam.NBATeamID]
	}

	_, err = s.TeamSeasonService.UpdateTeamSeasons(ctx, teamIDNBATeamMappings, seasonStartYear)
	if err != nil {
		return nil, err
	}

	return updatedTeams, nil
}

// get a mapping from nba team id -> db team id
func (s *service) NBATeamIDMappings(ctx context.Context) (map[string]string, error) {
	nbaTeamIDMappings, err := s.TeamStore.NBATeamIDMappings(ctx)
	if err != nil {
		return nil, err
	}

	return nbaTeamIDMappings, err
}
