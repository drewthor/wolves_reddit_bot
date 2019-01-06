package nba

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type TeamSchedule struct {
	LeagueNode struct {
		ScheduledGames []ScheduledGame `json:"standard"`
	} `json:"league"`
}

type ScheduledGame struct {
	GameID           string `json:"gameId"`
	StartDateEastern string `json:"startDateEastern"`
	StartTimeUTC     string `json:"startTimeUTC"`
}

func GetScheduledGames(teamAPIPath, teamID string) map[string]ScheduledGame {
	templateURI := MakeURIFormattable(NBAAPIBaseURI + teamAPIPath)
	url := fmt.Sprintf(templateURI, teamID)
	response, httpErr := http.Get(url)
	if httpErr != nil {
		log.Fatal(httpErr)
	}
	defer response.Body.Close()

	teamScheduleResult := TeamSchedule{}
	decodeErr := json.NewDecoder(response.Body).Decode(&teamScheduleResult)
	if decodeErr != nil {
		log.Fatal(decodeErr)
	}
	scheduledGameMap := map[string]ScheduledGame{}
	for _, scheduledGame := range teamScheduleResult.LeagueNode.ScheduledGames {
		if scheduledGame.StartDateEastern != "" && scheduledGame.StartTimeUTC != "" && scheduledGame.GameID != "" {
			scheduledGameMap[scheduledGame.StartDateEastern] = scheduledGame
		}
	}
	return scheduledGameMap
}