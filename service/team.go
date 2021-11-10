package service

import (
	"strconv"
	"strings"

	"github.com/drewthor/wolves_reddit_bot/api"
	"github.com/drewthor/wolves_reddit_bot/apis/nba"
	"github.com/drewthor/wolves_reddit_bot/dao"
)

type TeamService struct {
	TeamDAO *dao.TeamDAO
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

	teamUpdates := []dao.TeamUpdate{}
	for _, nbaTeam := range nbaTeams {
		league := "nba"
		if !nbaTeam.IsNBAFranchise {
			league = "international"
		}

		teamID, err := strconv.Atoi(nbaTeam.ID)
		if err != nil {
			return nil, err
		}

		teamUpdates = append(teamUpdates, dao.TeamUpdate{
			Name:            nbaTeam.FullName,
			Nickname:        nbaTeam.Nickname,
			City:            nbaTeam.City,
			AlternateCity:   nbaTeam.AlternateCity,
			LeagueName:      league,
			SeasonStartYear: seasonStartYear,
			ConferenceName:  strings.ToLower(nbaTeam.Conference),
			DivisionName:    strings.ToLower(nbaTeam.Division),
			NBAURLName:      nbaTeam.UrlName,
			NBATeamID:       teamID,
			NBAShortName:    nbaTeam.ShortName,
		})
	}

	updatedTeams, err := ts.TeamDAO.UpdateTeams(teamUpdates)
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
