package nba

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"sort"
)

type TeamSchedule struct {
	LeagueNode struct {
		ScheduledGames []ScheduledGame `json:"standard"`
	} `json:"league"`
}

type ScheduledGame struct {
	GameID           string            `json:"gameId"`
	SeasonStage      seasonStage       `json:"seasonStageId"`
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

// map from StartDateEastern to the ScheduledGame
type ScheduledGames map[string]ScheduledGame

func GetScheduledGames(teamAPIPath, teamID string) ScheduledGames {
	templateURI := makeURIFormattable(nbaAPIBaseURI + teamAPIPath)
	url := fmt.Sprintf(templateURI, teamID)
	log.Println(url)
	response, httpErr := http.Get(url)

	defer func() {
		response.Body.Close()
		io.Copy(ioutil.Discard, response.Body)
	}()

	if httpErr != nil {
		log.Fatal(httpErr)
	}

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

type byStartDate []ScheduledGame

func (b byStartDate) Len() int {
	return len(b)
}

func (b byStartDate) Swap(i, j int) {
	b[i], b[j] = b[j], b[i]
}

func (b byStartDate) Less(i, j int) bool {
	firstTime := makeGoTimeFromAPIData("1200" /*startTimeEastern*/, b[i].StartDateEastern)
	secondTime := makeGoTimeFromAPIData("1200" /*startTimeEastern*/, b[j].StartDateEastern)
	return firstTime.Before(secondTime)
}

func (s *ScheduledGames) CurrentGameNumber(gameID string, stage seasonStage) (int, bool) {
	var preSeasonGames []ScheduledGame
	var regularSeasonGames []ScheduledGame
	var postSeasonGames []ScheduledGame

	for _, scheduledGame := range *s {
		switch scheduledGame.SeasonStage {
		case preSeason:
			preSeasonGames = append(preSeasonGames, scheduledGame)
			break
		case regularSeason:
			regularSeasonGames = append(regularSeasonGames, scheduledGame)
			break
		case postSeason:
			postSeasonGames = append(postSeasonGames, scheduledGame)
			break
		}
	}
	switch stage {
	case preSeason:
		sort.Sort(byStartDate(preSeasonGames))
		for i, game := range preSeasonGames {
			if game.GameID == gameID {
				return i + 1, true
			}
		}
		break
	case regularSeason:
		sort.Sort(byStartDate(regularSeasonGames))
		for i, game := range regularSeasonGames {
			if game.GameID == gameID {
				return i + 1, true
			}
		}
		break
	case postSeason:
		sort.Sort(byStartDate(postSeasonGames))
		for i, game := range postSeasonGames {
			if game.GameID == gameID {
				return i + 1, true
			}
		}
		break
	}
	// game not found
	return -1, false
}
