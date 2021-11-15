package service

import (
	"strconv"

	"github.com/drewthor/wolves_reddit_bot/api"
	"github.com/drewthor/wolves_reddit_bot/apis/nba"
	"github.com/drewthor/wolves_reddit_bot/dao"
)

type TeamService struct {
	TeamDAO           *dao.TeamDAO
	TeamSeasonService *TeamSeasonService
}

func (ts TeamService) Get(teamID string) (api.Team, error) {
	team, err := ts.TeamDAO.Get(teamID)
	if err != nil {
		return api.Team{}, err
	}

	return team, err
}

func (ts TeamService) GetAll() ([]api.Team, error) {
	teams, err := ts.TeamDAO.GetAll()
	if err != nil {
		return nil, err
	}

	return teams, err
}

func (ts TeamService) UpdateTeams(seasonStartYear int) ([]api.Team, error) {
	nbaTeams, err := nba.GetTeamsForSeason(seasonStartYear)
	if err != nil {
		return nil, err
	}

	nbaTeamIDTeamMappings := map[int]nba.Team{}

	teamUpdates := []dao.TeamUpdate{}
	for _, nbaTeam := range nbaTeams {
		teamID, err := strconv.Atoi(nbaTeam.ID)
		if err != nil {
			return nil, err
		}

		teamUpdates = append(teamUpdates, dao.TeamUpdate{
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

	updatedTeams, err := ts.TeamDAO.UpdateTeams(teamUpdates)
	if err != nil {
		return nil, err
	}

	for _, updatedTeam := range updatedTeams {
		teamIDNBATeamMappings[updatedTeam.ID] = nbaTeamIDTeamMappings[updatedTeam.NBATeamID]
	}

	_, err = ts.TeamSeasonService.UpdateTeamSeasons(teamIDNBATeamMappings, seasonStartYear)
	if err != nil {
		return nil, err
	}

	return updatedTeams, nil
}

// get a mapping from nba team id -> db team id
func (ts TeamService) NBATeamIDMappings() (map[string]string, error) {
	nbaTeamIDMappings, err := ts.TeamDAO.NBATeamIDMappings()
	if err != nil {
		return nil, err
	}

	return nbaTeamIDMappings, err
}
