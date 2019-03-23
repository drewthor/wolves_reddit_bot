package nba

import (
	"encoding/json"
	"log"
	"net/http"
)

type Players struct {
	LeagueNode struct {
		Players []Player `json:"standard"`
	} `json:"league"`
}

type Player struct {
	ID        string `json:"personId"`
	TeamID    string `json:"teamId"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

func GetPlayers(playersAPIPath string) map[string]Player {
	url := makeURIFormattable(nbaAPIBaseURI + playersAPIPath)
	response, httpErr := http.Get(url)
	if httpErr != nil {
		log.Fatal(httpErr)
	}
	defer response.Body.Close()

	playersResult := Players{}
	decodeErr := json.NewDecoder(response.Body).Decode(&playersResult)
	if decodeErr != nil {
		log.Fatal(decodeErr)
	}
	playerMap := map[string]Player{}
	for _, player := range playersResult.LeagueNode.Players {
		playerMap[player.ID] = player
	}
	return playerMap
}
