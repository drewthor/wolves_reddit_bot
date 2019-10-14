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
	Nickname       string  `json:"nickname"`
}

func GetTeams(teamsAPIPath string) map[TriCode]Team {
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
	teamMap := map[TriCode]Team{}
	for _, team := range teamsResult.LeagueNode.Teams {
		teamMap[team.TriCode] = team
	}

	return teamMap
}
