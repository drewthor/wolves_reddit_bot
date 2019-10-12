package nba

import (
	"encoding/json"
	"io"
	"io/ioutil"
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

	defer func() {
		response.Body.Close()
		io.Copy(ioutil.Discard, response.Body)
	}()

	if httpErr != nil {
		log.Fatal(httpErr)
	}

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
