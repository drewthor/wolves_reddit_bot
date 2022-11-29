package nba

import (
	"fmt"
	"net/http"
)

const TeamsURL = "https://data.nba.net/prod/v2/%d/teams.json"

const TeamLogoUrl = "https://cdn.nba.com/logos/nba/%d/primary/L/logo.svg"

type TeamsResult struct {
	LeagueNode struct {
		NBA        []Team `json:"standard"`
		Vegas      []Team `json:"vegas,omitempty"`
		Sacramento []Team `json:"sacramento,omitempty"`
		Utah       []Team `json:"utah,omitempty"`
	} `json:"league"`
}

type Team struct {
	IsNBAFranchise bool    `json:"isNBAFranchise"`
	ID             string  `json:"teamId"`
	TriCode        TriCode `json:"tricode"`
	FullName       string  `json:"fullName"`
	ShortName      string  `json:"teamShortName"`
	Nickname       string  `json:"nickname"`
	City           string  `json:"city"`
	AlternateCity  string  `json:"altCityName"`
	UrlName        string  `json:"urlName"`
	Conference     string  `json:"confName"`
	Division       string  `json:"divName"`
	AllStar        bool    `json:"isAllStar"`
}

func GetTeamsForSeason(seasonStartYear int) ([]Team, error) {
	url := fmt.Sprintf(TeamsURL, seasonStartYear)
	response, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to get current teams from nba from url %s", url)
	}
	defer response.Body.Close()

	teamsResult, err := unmarshalNBAHttpResponseToJSON[TeamsResult](response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to get current teams from nba from url %s", url)
	}
	teams := []Team{}
	for _, team := range teamsResult.LeagueNode.NBA {
		teams = append(teams, team)
	}

	return teams, nil
}

func GetTeams(teamsAPIPath string) ([]Team, error) {
	url := nbaAPIBaseURI + teamsAPIPath
	response, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to get current teams from nba from url %s", url)
	}
	defer response.Body.Close()

	teamsResult, err := unmarshalNBAHttpResponseToJSON[TeamsResult](response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to get current teams from nba from url %s", url)
	}
	teams := []Team{}
	for _, team := range teamsResult.LeagueNode.NBA {
		teams = append(teams, team)
	}

	return teams, nil
}
