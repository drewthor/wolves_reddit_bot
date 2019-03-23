package nba

import (
	"encoding/json"
	"log"
	"net/http"
)

type TeamsResult struct {
	LeagueNode struct {
		Teams []Team `json:"standard"`
	} `json:"league"`
}

type Team struct {
	IsNBAFranchise  bool   `json:"isNBAFranchise"`
	ID              string `json:"teamId"`
	AbbreviatedName string `json:"tricode"`
	FullName        string `json:"fullName"`
	Nickname        string `json:"nickname"`
}

func GetTeams(teamsAPIPath string) map[string]Team {
	url := nbaAPIBaseURI + teamsAPIPath
	response, httpErr := http.Get(url)
	if httpErr != nil {
		log.Fatal(httpErr)
	}
	defer response.Body.Close()

	teamsResult := TeamsResult{}
	decodeErr := json.NewDecoder(response.Body).Decode(&teamsResult)
	if decodeErr != nil {
		log.Fatal(decodeErr)
	}
	teamMap := map[string]Team{}
	for _, team := range teamsResult.LeagueNode.Teams {
		if team.IsNBAFranchise {
			teamMap[team.AbbreviatedName] = team
		}
	}
	return teamMap
}
