package nba

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

type TeamSchedule struct {
	LeagueNode struct {
		Games []Game `json:"standard"`
	} `json:"league"`
}

func GetCurrentTeamSchedule(teamAPIPath, teamID string) (GamesByStartDate, error) {
	templateURI := makeURIFormattable(nbaAPIBaseURI + teamAPIPath)
	url := fmt.Sprintf(templateURI, teamID)
	log.Println(url)
	response, err := http.Get(url)

	defer func() {
		response.Body.Close()
		io.Copy(ioutil.Discard, response.Body)
	}()

	if err != nil {
		log.Error(err)
		return nil, err
	}

	teamScheduleResult := TeamSchedule{}
	err = json.NewDecoder(response.Body).Decode(&teamScheduleResult)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	scheduledGameMap := map[string]Game{}
	for _, scheduledGame := range teamScheduleResult.LeagueNode.Games {
		if scheduledGame.StartDateEastern != "" && scheduledGame.StartTimeUTC.Time.Equal(time.Time{}) && scheduledGame.GameID != "" {
			scheduledGameMap[scheduledGame.StartDateEastern] = scheduledGame
		}
	}
	return scheduledGameMap, nil
}
