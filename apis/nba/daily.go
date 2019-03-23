package nba

import (
	"encoding/json"
	"log"
	"net/http"
)

const NBADailyAPIPath = "/prod/v1/today.json"

type DailyAPI struct {
	APIPaths DailyAPIPaths `json:"links"`
}

type DailyAPIPaths struct {
	Boxscore     string `json:"boxscore"`
	CurrentDate  string `json:"currentDate"`
	Players      string `json:"leagueRosterPlayers"`
	Scoreboard   string `json:"scoreboard"`
	Teams        string `json:"teams"`
	TeamSchedule string `json:"teamSchedule"`
}

func GetDailyAPIPaths() DailyAPIPaths {
	url := nbaAPIBaseURI + NBADailyAPIPath
	response, httpErr := http.Get(url)
	if httpErr != nil {
		log.Fatal(httpErr)
	}
	defer response.Body.Close()

	dailyAPIResult := DailyAPI{}
	decodeErr := json.NewDecoder(response.Body).Decode(&dailyAPIResult)
	if decodeErr != nil {
		log.Fatal(decodeErr)
	}
	if dailyAPIResult.APIPaths.CurrentDate == "" || dailyAPIResult.APIPaths.Teams == "" || dailyAPIResult.APIPaths.TeamSchedule == "" || dailyAPIResult.APIPaths.Scoreboard == "" {
		log.Fatal("Could not get daily API paths")
	}
	return dailyAPIResult.APIPaths
}
