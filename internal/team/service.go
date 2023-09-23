package team

import (
	"context"
	"fmt"
	"strconv"

	"github.com/drewthor/wolves_reddit_bot/api"
	"github.com/drewthor/wolves_reddit_bot/apis/nba"
	"github.com/drewthor/wolves_reddit_bot/internal/team_season"
	"go.opentelemetry.io/otel"
)

type Service interface {
	Get(ctx context.Context, teamID string) (api.Team, error)
	ListTeams(ctx context.Context) ([]api.Team, error)
	UpdateTeams(ctx context.Context, seasonStartYear int) ([]api.Team, error)
	EnsureTeamsExistForLeague(ctx context.Context, nbaLeagueID string, nbaTeamIDs []int) error
}

func NewService(teamStore Store, teamSeasonService team_season.Service, nbaClient nba.Client) Service {
	return &service{
		teamStore:         teamStore,
		teamSeasonService: teamSeasonService,
		nbaClient:         nbaClient,
	}
}

type service struct {
	teamStore         Store
	teamSeasonService team_season.Service

	nbaClient nba.Client
}

func (s service) Get(ctx context.Context, teamID string) (api.Team, error) {
	team, err := s.teamStore.GetTeamWithID(ctx, teamID)
	if err != nil {
		return api.Team{}, err
	}

	return team, err
}

func (s service) ListTeams(ctx context.Context) ([]api.Team, error) {
	teams, err := s.teamStore.ListTeams(ctx)
	if err != nil {
		return nil, err
	}

	return teams, err
}

func (s service) UpdateTeams(ctx context.Context, seasonStartYear int) ([]api.Team, error) {
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
			AlternateCity: &nbaTeam.AlternateCity,
			NBAURLName:    nbaTeam.UrlName,
			NBATeamID:     teamID,
			NBAShortName:  nbaTeam.ShortName,
		})

		nbaTeamIDTeamMappings[teamID] = nbaTeam
	}

	teamIDNBATeamMappings := map[string]nba.Team{}

	updatedTeams, err := s.teamStore.UpdateTeams(ctx, teamUpdates)
	if err != nil {
		return nil, err
	}

	for _, updatedTeam := range updatedTeams {
		teamIDNBATeamMappings[updatedTeam.ID] = nbaTeamIDTeamMappings[updatedTeam.NBATeamID]
	}

	_, err = s.teamSeasonService.UpdateTeamSeasons(ctx, teamIDNBATeamMappings, seasonStartYear)
	if err != nil {
		return nil, err
	}

	return updatedTeams, nil
}

func (s service) UpdateTeamsForLeague(ctx context.Context, nbaLeagueID string, teamIDs []int) ([]api.Team, error) {
	var teamUpdates []TeamUpdate
	for _, teamID := range teamIDs {
		teamInfo, err := s.nbaClient.GetCommonTeamInfo(ctx, nbaLeagueID, teamID)
		if err != nil {
			return nil, fmt.Errorf("failed to update teams for season: %w", err)
		}

		teamUpdates = append(teamUpdates, TeamUpdate{
			Name:       teamInfo.Name,
			Nickname:   teamInfo.Name,
			City:       teamInfo.City,
			NBAURLName: teamInfo.Slug,
			NBATeamID:  teamID,
		})
	}

	updatedTeams, err := s.teamStore.UpdateTeams(ctx, teamUpdates)
	if err != nil {
		return nil, fmt.Errorf("failed to update teams in db: %w", err)
	}

	return updatedTeams, nil
}

func (s service) EnsureTeamsExistForLeague(ctx context.Context, nbaLeagueID string, nbaTeamIDs []int) error {
	ctx, span := otel.Tracer("team").Start(ctx, "team.service.EnsureTeamsExistForLeague")
	defer span.End()

	existingTeams, err := s.teamStore.GetTeamsWithNBAIDs(ctx, nbaTeamIDs)
	if err != nil {
		return fmt.Errorf("failed to ensure teams exist: %w", err)
	}

	existingTeamIDsMap := make(map[int]bool, len(existingTeams))
	for _, existingTeam := range existingTeams {
		existingTeamIDsMap[existingTeam.NBATeamID] = true
	}

	var missingTeamIDs []int
	for _, teamID := range nbaTeamIDs {
		if !existingTeamIDsMap[teamID] {
			missingTeamIDs = append(missingTeamIDs, teamID)
		}
	}

	if _, err = s.UpdateTeamsForLeague(ctx, nbaLeagueID, missingTeamIDs); err != nil {
		return fmt.Errorf("failed to ensure teams exist for league: %w", err)
	}

	return nil
}

// get a mapping from nba team id -> db team id
func (s service) NBATeamIDMappings(ctx context.Context) (map[string]string, error) {
	ctx, span := otel.Tracer("team").Start(ctx, "team.service.NBATeamIDMappings")
	defer span.End()

	nbaTeamIDMappings, err := s.teamStore.NBATeamIDMappings(ctx)
	if err != nil {
		return nil, err
	}

	return nbaTeamIDMappings, err
}
