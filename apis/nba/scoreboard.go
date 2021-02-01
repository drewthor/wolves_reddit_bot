package nba

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
)

type Scoreboard struct {
	Games []GameScoreboard `json:"games"`
}

type GameScoreboard struct {
	Active       bool             `json:"isGameActivated"`
	GameDuration GameDuration     `json:"gameDuration"`
	ID           string           `json:"gameId"`
	Period       Period           `json:"period"`
	StartTimeUTC string           `json:"startTimeUTC"`
	EndTimeUTC   string           `json:"endTimeUTC,omitempty"`
	HomeTeamInfo TeamBoxscoreInfo `json:"hTeam"`
	AwayTeamInfo TeamBoxscoreInfo `json:"vTeam"`
}

type GameDuration struct {
	Hours   string `json:"hours"`
	Minutes string `json:"minutes"`
}

type Period struct {
	Current int `json:"current"`
}

func GetGameScoreboard(scoreboardAPIPath, todaysDate string, gameID string) GameScoreboard {
	templateURI := makeURIFormattable(nbaAPIBaseURI + scoreboardAPIPath)
	url := fmt.Sprintf(templateURI, todaysDate)
	response, httpErr := http.Get(url)

	defer func() {
		response.Body.Close()
		io.Copy(ioutil.Discard, response.Body)
	}()

	if httpErr != nil {
		log.Fatal(httpErr)
	}

	scoreboardResult := Scoreboard{}
	decodeErr := json.NewDecoder(response.Body).Decode(&scoreboardResult)
	if decodeErr != nil {
		log.Fatal(decodeErr)
	}
	for _, game := range scoreboardResult.Games {
		if game.ID == gameID {
			return game
		}
	}
	log.Fatal("Game not found")
	return GameScoreboard{}
}

func GetGameScoreboards(gameDate string) Scoreboard {
	gameScoreboardAPIPath := GetDailyAPIPaths().Scoreboard
	templateURI := makeURIFormattable(nbaAPIBaseURI + gameScoreboardAPIPath)
	url := fmt.Sprintf(templateURI, gameDate)
	log.Println(url)
	response, httpErr := http.Get(url)

	defer func() {
		response.Body.Close()
		io.Copy(ioutil.Discard, response.Body)
	}()

	if httpErr != nil {
		log.Fatal(httpErr)
	}
	scoreboardResult := Scoreboard{}
	decodeErr := json.NewDecoder(response.Body).Decode(&scoreboardResult)
	if decodeErr != nil {
		log.Fatal(decodeErr)
	}
	return scoreboardResult
}
