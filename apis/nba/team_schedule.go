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
	GameID           string            `json:"gameId"`
	StartDateEastern string            `json:"startDateEastern"`
	StartTimeUTC     string            `json:"startTimeUTC"`
	PlayoffsNode     *PlayoffsGameInfo `json:"playoffs,omitempty"`
}

func (s ScheduledGame) IsPlayoffGame() bool {
	return s.PlayoffsNode != nil
}

type PlayoffsGameInfo struct {
	Round           string               `json:"roundNum"`
	Conference      string               `json:"confName"`
	SeriesID        string               `json:"seriesId"`
	SeriesCompleted bool                 `json:"isSeriesCompleted"`
	GameInSeries    string               `json:"gameNumInSeries"`
	IsIfNecessary   bool                 `json:"isIfNecessary"`
	HomeTeamInfo    PlayoffsGameTeamInfo `json:"hTeam"`
	AwayTeamInfo    PlayoffsGameTeamInfo `json:"vTeam"`
}

type PlayoffsGameTeamInfo struct {
	Seed       string `json:"seedNum"`
	SeriesWins string `json:"seriesWin"`
	WonSeries  bool   `json:"isSeriesWinner"`
}

type ScheduledGames map[string]ScheduledGame

func GetScheduledGames(teamAPIPath, teamID string) ScheduledGames {
	templateURI := makeURIFormattable(nbaAPIBaseURI + teamAPIPath)
	url := fmt.Sprintf(templateURI, teamID)
	log.Println(url)
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

func (s *ScheduledGames) HaveAnotherMatchup(opposingTeam TriCode, todaysDate string) bool {
	for _, scheduledGame := range *s {
		isFutureGame := scheduledGame.StartDateEastern > todaysDate
		if scheduledGame.IsPlayoffGame() {
			if !scheduledGame.PlayoffsNode.SeriesCompleted && !scheduledGame.PlayoffsNode.IsIfNecessary {
				return true
			}
		} else if isFutureGame {
			return true
		}
	}
	return false
}
