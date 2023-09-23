package nba

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/drewthor/wolves_reddit_bot/apis/cloudflare"
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
	StartTimeUTC time.Time        `json:"startTimeUTC"`
	EndTimeUTC   time.Time        `json:"endTimeUTC,omitempty"`
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
	LeagueID   string `json:"leagueId"`   // ex. 00 for NBA, 15 for Las Vegas, 13 for California classical, 16 for Utah, and 14 for Orlando
	LeagueName string `json:"leagueName"` // ex. National Basketball Association
	Games      []struct {
		GameID            string         `json:"gameId"`         // ex. 20211108
		GameCode          string         `json:"gameCode"`       // ex. 20211108/NYKPHI
		GameStatus        int            `json:"gameStatus"`     // ex. 1
		GameStatusText    string         `json:"gameStatusText"` // ex. 7:00 pm ET
		Period            int            `json:"period"`
		GameClock         duration       `json:"gameClock"`
		GameTimeUTC       time.Time      `json:"gameTimeUTC"`
		GameTimeET        time.Time      `json:"gameTimeET"`
		RegulationPeriods int            `json:"regulationPeriods"`
		IfNecessary       bool           `json:"ifNecessary"`
		SeriesGameNumber  string         `json:"seriesGameNumber"`
		SeriesText        string         `json:"seriesText"`
		HomeTeam          TeamScoreboard `json:"homeTeam"`
		AwayTeam          TeamScoreboard `json:"awayTeam"`
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

func GetTodaysScoreboard(ctx context.Context, r2Client cloudflare.Client, bucket string) (TodaysScoreboard, error) {
	t := time.Now().UTC().Round(time.Hour).Format(time.RFC3339)
	filePath := os.Getenv("STORAGE_PATH") + fmt.Sprintf("/scoreboard/%s", t)

	objectKey := fmt.Sprintf("scoreboard/%s_cdn.json", t)

	todaysScoreboard, err := fetchObjectAndSaveToFile[TodaysScoreboard](ctx, r2Client, todaysScoreboardURL, filePath, bucket, objectKey)
	if err != nil {
		return TodaysScoreboard{}, err
	}

	return todaysScoreboard, nil
}
