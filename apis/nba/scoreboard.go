package nba

import (
	"fmt"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

const todaysScoreboardURL = "https://cdn.nba.com/static/json/liveData/scoreboard/todaysScoreboard_00.json"

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

type TodaysScoreboard struct {
	Scoreboard ScoreboardDetailed `json:"scoreboard"`
}

type ScoreboardDetailed struct {
	GameDate   string `json:"gameDate"`   // ex. 2021-11-08
	LeagueID   string `json:"leagueId"`   // ex. 00 for NBA
	LeagueName string `json:"leagueName"` // ex. National Basketball Association
	Games      []struct {
		GameID            string                        `json:"gameId"`         // ex. 20211108
		GameCode          string                        `json:"gameCode"`       // ex. 20211108/NYKPHI
		GameStatus        int                           `json:"gameStatus"`     // ex. 1
		GameStatusText    string                        `json:"gameStatusText"` // ex. 7:00 pm ET
		Period            int                           `json:"period"`
		GameClock         BoxscoreGameClockTenthSeconds `json:"gameClock"`
		GameTimeUTC       time.Time                     `json:"gameTimeUTC"`
		GameTimeET        time.Time                     `json:"gameTimeET"`
		RegulationPeriods int                           `json:"regulationPeriods"`
		IfNecessary       bool                          `json:"ifNecessary"`
		SeriesGameNumber  string                        `json:"seriesGameNumber"`
		SeriesText        string                        `json:"seriesText"`
		HomeTeam          TeamScoreboard                `json:"homeTeam"`
		AwayTeam          TeamScoreboard                `json:"awayTeam"`
	} `json:"games"`
	GameLeaders struct {
		HomeTeamLeaders ScoreboardTeamLeaders `json:"homeLeaders"`
		AwayTeamLeaders ScoreboardTeamLeaders `json:"awayLeaders"`
	} `json:"gameLeaders"`
}

type TeamScoreboard struct {
	TeamID            int     `json:"teamId"`
	TeamName          string  `json:"teamName"` // ex. 76ers
	TeamCity          string  `json:"teamCity"`
	TeamTricode       string  `json:"tricode"`
	Wins              int     `json:"wins"`
	Losses            int     `json:"losses"`
	Score             int     `json:"score"`
	Seed              *int    `json:"seed"`    // ex. null
	InBonus           *string `json:"inBonus"` // ex. null
	TimeoutsRemaining int     `json:"timeoutsRemaining"`
	Periods           []struct {
		Period     int    `json:"period"`
		PeriodType string `json:"periodType"`
		Score      int    `json:"score"`
	} `json:"periods"`
}

type ScoreboardTeamLeaders struct {
	PersonID     int     `json:"personId"`
	Name         string  `json:"name"`
	JerseyNumber string  `json:"jerseyNum"`
	Position     string  `json:"position"`
	TeamTricode  string  `json:"teamTricode"`
	PlayerSlug   *string `json:"playerSlug"`
	Points       int     `json:"points"`
	Rebounds     int     `json:"rebounds"`
	Assists      int     `json:"assists"`
}

func GetTodaysScoreboard() (TodaysScoreboard, error) {
	response, err := http.Get(todaysScoreboardURL)
	if err != nil {
		return TodaysScoreboard{}, err
	}
	defer response.Body.Close()

	todaysScoreboard, err := unmarshalNBAHttpResponseToJSON[TodaysScoreboard](response.Body)
	if err != nil {
		return TodaysScoreboard{}, fmt.Errorf("failed to decode todays scoreboard from nba %s %w", time.Now().UTC().Format(time.RFC3339), err)
	}

	return todaysScoreboard, nil
}

func GetGameScoreboard(scoreboardAPIPath, todaysDate string, gameID string) (GameScoreboard, error) {
	templateURI := makeURIFormattable(nbaAPIBaseURI + scoreboardAPIPath)
	url := fmt.Sprintf(templateURI, todaysDate)
	response, err := http.Get(url)
	if err != nil {
		return GameScoreboard{}, fmt.Errorf("failed to get game scoreboard from nba for date %s and gameID %d from url %s %w", todaysDate, gameID, url, err)
	}
	defer response.Body.Close()

	scoreboardResult, err := unmarshalNBAHttpResponseToJSON[Scoreboard](response.Body)
	if err != nil {
		return GameScoreboard{}, fmt.Errorf("failed to get game scoreboard from nba for date %s and gameID %d from url %s %w", todaysDate, gameID, url, err)
	}
	for _, game := range scoreboardResult.Games {
		if game.ID == gameID {
			return game, nil
		}
	}
	return GameScoreboard{}, fmt.Errorf("could not find game with ID: %s on date: %s", gameID, todaysDate)
}

func GetGameScoreboards(gameDate string) (Scoreboard, error) {
	dailyAPIPaths, err := GetDailyAPIPaths()
	if err != nil {
		return Scoreboard{}, err
	}
	gameScoreboardAPIPath := dailyAPIPaths.APIPaths.Scoreboard
	templateURI := makeURIFormattable(nbaAPIBaseURI + gameScoreboardAPIPath)
	url := fmt.Sprintf(templateURI, gameDate)
	log.Println(url)
	response, err := http.Get(url)
	if err != nil {
		return Scoreboard{}, fmt.Errorf("failed to get scoreboard from nba from url %s %w", url, err)
	}
	defer response.Body.Close()

	scoreboardResult, err := unmarshalNBAHttpResponseToJSON[Scoreboard](response.Body)
	if err != nil {
		return Scoreboard{}, fmt.Errorf("failed to get scoreboard from nba from url %s %w", url, err)
	}
	return scoreboardResult, nil
}
