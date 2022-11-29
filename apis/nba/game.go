package nba

import (
	"fmt"
	"sort"

	log "github.com/sirupsen/logrus"
)

type Game struct {
	GameID           string            `json:"gameId"`
	SeasonStage      seasonStage       `json:"seasonStageId"`
	Status           int               `json:"statusNum"`
	StartDateEastern string            `json:"startDateEastern"`
	StartTimeUTC     datetime          `json:"startTimeUTC"`
	PlayoffsNode     *PlayoffsGameInfo `json:"playoffs,omitempty"`
	HomeTeam         TeamGameInfo      `json:"hTeam"`
	AwayTeam         TeamGameInfo      `json:"vTeam"`
}

func (s Game) IsPlayoffGame() bool {
	return s.PlayoffsNode != nil
}

type TeamGameInfo struct {
	ID    string `json:"teamId"`
	Score string `json:"score"`
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

// map from StartDateEastern to the Game
type GamesByStartDate map[string]Game

func (s *GamesByStartDate) HaveAnotherMatchup(opposingTeam TriCode, todaysDate string) bool {
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

type startDateAscending []Game

func (b startDateAscending) Len() int {
	return len(b)
}

func (b startDateAscending) Swap(i, j int) {
	b[i], b[j] = b[j], b[i]
}

func (b startDateAscending) Less(i, j int) bool {
	firstTime, err := makeGoTimeFromAPIData("12:00 PM ET" /*startTimeEastern*/, b[i].StartDateEastern)
	if err != nil {
		return false
	}
	secondTime, err := makeGoTimeFromAPIData("12:00 PM ET" /*startTimeEastern*/, b[j].StartDateEastern)
	if err != nil {
		return false
	}
	return firstTime.Before(secondTime)
}

func (s *GamesByStartDate) CurrentGameNumber(gameID string, stage seasonStage) (int, bool) {
	var preSeasonGames []Game
	var regularSeasonGames []Game
	var postSeasonGames []Game

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
		sort.Sort(startDateAscending(preSeasonGames))
		for i, game := range preSeasonGames {
			if game.GameID == gameID {
				return i + 1, true
			}
		}
		break
	case regularSeason:
		sort.Sort(startDateAscending(regularSeasonGames))
		for i, game := range regularSeasonGames {
			if game.GameID == gameID {
				return i + 1, true
			}
		}
		break
	case postSeason:
		sort.Sort(startDateAscending(postSeasonGames))
		for i, game := range postSeasonGames {
			if game.GameID == gameID {
				return i + 1, true
			}
		}
		break
	}
	// game not found
	log.Error(fmt.Printf("failed to find current game number from nba gameID: %s stage: %v ", gameID, stage))
	return -1, false
}
