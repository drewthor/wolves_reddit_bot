package nba

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type Scoreboard struct {
	Games []GameScoreboard `json:"games"`
}

type GameScoreboard struct {
	Active       bool         `json:"isGameActivated"`
	GameDuration GameDuration `json:"gameDuration"`
	ID           string       `json:"gameId"`
	Period       Period       `json:"period"`
	StartTimeUTC string       `json:"startTimeUTC"`
	EndTimeUTC   string       `json:"endTimeUTC,omitempty"`
}

type GameDuration struct {
	Hours   string `json:"hours"`
	Minutes string `json:"minutes"`
}

type Period struct {
	Current int `json:"current"`
}

func GetGameScoreboard(scoreboardAPIPath, todaysDate string, gameID string) GameScoreboard {
	templateURI := MakeURIFormattable(NBAAPIBaseURI + scoreboardAPIPath)
	url := fmt.Sprintf(templateURI, todaysDate)
	response, httpErr := http.Get(url)
	if httpErr != nil {
		log.Fatal(httpErr)
	}
	defer response.Body.Close()

	scoreboardAPIResult := Scoreboard{}
	decodeErr := json.NewDecoder(response.Body).Decode(&scoreboardAPIResult)
	if decodeErr != nil {
		log.Fatal(decodeErr)
	}
	for _, game := range scoreboardAPIResult.Games {
		if game.ID == gameID {
			return game
		}
	}
	log.Fatal("Game not found")
	return GameScoreboard{}
}
