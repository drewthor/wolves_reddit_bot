package service

import (
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

func (ts TeamService) UpdateTeams() ([]api.Team, error) {
	teams, err := ts.getAllTeamsFromNBAApi()
	if err != nil {
		return nil, err
	}

	updatedTeams, err := ts.TeamDAO.UpdateTeams(teams)
	if err != nil {
		return nil, err
	}
	return updatedTeams, nil
}

func (ts TeamService) getAllTeamsFromNBAApi() ([]api.Team, error) {
	nbaTeams := nba.GetTeams(nba.GetDailyAPIPaths().Teams)

	teams := []api.Team{}
	for _, nbaTeam := range nbaTeams {
		league := "nba"
		if !nbaTeam.IsNBAFranchise {
			league = "international"
		}
		teams = append(teams, api.Team{
			Name:          nbaTeam.FullName,
			Nickname:      nbaTeam.Nickname,
			City:          nbaTeam.City,
			AlternateCity: nbaTeam.AlternateCity,
			League:        league,
			Season:        "2020",
			Conference:    strings.ToLower(nbaTeam.Conference),
			Division:      strings.ToLower(nbaTeam.Division),
			NBAURLName:    nbaTeam.UrlName,
			NBATeamID:     nbaTeam.ID,
			NBAShortName:  nbaTeam.ShortName,
		})
	}

	return teams, nil
}
