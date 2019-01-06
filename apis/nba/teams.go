package nba

import (
	"encoding/json"
	"log"
	"net/http"
)

type TeamsResult struct {
	TeamsNode struct {
		Teams []Team `json:"config"`
	} `json:"teams"`
}
type Team struct {
	ID   string `json:"teamId"`
	Abbr string `json:"tricode"`
	Name string `json:"ttsName"`
}

func GetTeams() map[string]Team {
	url := "http://data.nba.net/prod/2018/teams_config.json"
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
	for _, team := range teamsResult.TeamsNode.Teams {
		if team.ID != "" && team.Abbr != "" && team.Name != "" {
			teamMap[team.Abbr] = team
		}
	}
	return teamMap
}
