package nba

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	log "github.com/sirupsen/logrus"
)

const PlayerHeadshotURL = "https://cdn.nba.com/headshots/nba/latest/260x190/%d.png"
const seasonPlayersURL = "https://data.nba.net/prod/v1/%d/players.json"

type Players struct {
	LeagueNode struct {
		Players []Player `json:"standard"`
	} `json:"league"`
}

type Player struct {
	ID              string `json:"personId"`
	TeamID          string `json:"teamId"`
	FirstName       string `json:"firstName"`
	LastName        string `json:"lastName"`
	Jersey          string `json:"jersey"`
	CurrentlyInNBA  bool   `json:"isActive"`
	Position        string `json:"pos"`
	HeightFeet      string `json:"heightFeet"`
	HeightInches    string `json:"heightInches"`
	HeightMeters    string `json:"heightMeters"`
	WeightPounds    string `json:"weightPounds"`
	WeightKilograms string `json:"weightKilograms"`
	DateOfBirthUTC  string `json:"dateOfBirthUTC"` // this is in format yyyy-mm-dd (nba.TimeBirthdateFormat)
	NBADebutYear    string `json:"nbaDebutYear"`
	YearsPro        string `json:"yearsPro"`
	CollegeName     string `json:"collegeName"`
	LastAffiliation string `json:"lastAffiliation"`
	Country         string `json:"country"`
}

func GetPlayers(seasonStartYear int) ([]Player, error) {
	url := fmt.Sprintf(seasonPlayersURL, seasonStartYear)
	response, httpErr := http.Get(url)

	defer func() {
		response.Body.Close()
		io.Copy(ioutil.Discard, response.Body)
	}()

	if httpErr != nil {
		log.Println(httpErr)
		return nil, httpErr
	}

	playersResult := Players{}
	decodeErr := json.NewDecoder(response.Body).Decode(&playersResult)
	if decodeErr != nil {
		log.Println(decodeErr)
		return nil, decodeErr
	}

	return playersResult.LeagueNode.Players, nil
}
