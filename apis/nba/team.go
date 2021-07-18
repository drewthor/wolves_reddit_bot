package nba

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
)

type TeamsResult struct {
	LeagueNode struct {
		Teams []Team `json:"standard"`
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
	for _, team := range teamsResult.LeagueNode.Teams {
		teams = append(teams, team)
	}

	return teams
}
