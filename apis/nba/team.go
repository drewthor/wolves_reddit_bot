package nba

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
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

	if response != nil {
		defer response.Body.Close()
	}

	if err != nil {
		return nil, err
	}

	teamsResult := TeamsResult{}
	err = json.NewDecoder(response.Body).Decode(&teamsResult)
	if err != nil {
		return nil, err
	}
	teams := []Team{}
	for _, team := range teamsResult.LeagueNode.NBA {
		teams = append(teams, team)
	}

	return teams, nil
}

func GetTeams(teamsAPIPath string) []Team {
	url := nbaAPIBaseURI + teamsAPIPath
	response, httpErr := http.Get(url)

	defer func() {
		response.Body.Close()
		io.Copy(ioutil.Discard, response.Body)
	}()

	if httpErr != nil {
		log.Fatal(httpErr)
	}

	teamsResult := TeamsResult{}
	decodeErr := json.NewDecoder(response.Body).Decode(&teamsResult)
	if decodeErr != nil {
		log.Fatal(decodeErr)
	}
	teams := []Team{}
	for _, team := range teamsResult.LeagueNode.NBA {
		teams = append(teams, team)
	}

	return teams
}
