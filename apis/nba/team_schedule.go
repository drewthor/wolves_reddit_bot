package nba

import (
	"fmt"
	"net/http"
	"time"
)

type TeamSchedule struct {
	LeagueNode struct {
		Games []Game `json:"standard"`
	} `json:"league"`
}

func GetCurrentTeamSchedule(teamAPIPath, teamID string) (GamesByStartDate, error) {
	templateURI := makeURIFormattable(nbaAPIBaseURI + teamAPIPath)
	url := fmt.Sprintf(templateURI, teamID)
	response, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to get current team schedule from nba for teamID %d from url %s %w", teamID, url, err)
	}
	defer response.Body.Close()

	teamScheduleResult, err := unmarshalNBAHttpResponseToJSON[TeamSchedule](response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to get current team schedule from nba for teamID %d from url %s %w", teamID, url, err)
	}
	scheduledGameMap := map[string]Game{}
	for _, scheduledGame := range teamScheduleResult.LeagueNode.Games {
		if scheduledGame.StartDateEastern != "" && scheduledGame.StartTimeUTC.Time.Equal(time.Time{}) && scheduledGame.GameID != "" {
			scheduledGameMap[scheduledGame.StartDateEastern] = scheduledGame
		}
	}
	return scheduledGameMap, nil
}
