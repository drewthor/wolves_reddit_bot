package nba

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"slices"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
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

type SeriesText string

const (
	SeriesTextEmpty               SeriesText = "" // "" - this is probably just always regular season
	SeriesTextISTChampionship     SeriesText = "Championship"
	SeriesTextISTEastGroupA       SeriesText = "East Group A"         // In season tournament
	SeriesTextISTEastGroupB       SeriesText = "East Group B"         // In season tournament
	SeriesTextISTEastGroupC       SeriesText = "East Group C"         // In season tournament
	SeriesTextISTEastQuarterFinal SeriesText = "East Quarterfinal"    // In season tournament
	SeriesTextISTEastSemiFinal    SeriesText = "East Semifinal"       // In season tournament
	SeriesTextMexicoCityGame      SeriesText = "NBA Mexico City Game" // Regular season
	SeriesTextParisGame           SeriesText = "NBA Paris Game"       // Regular season
	SeriesTextPreseason           SeriesText = "Preseason"
	SeriesTextISTWestGroupA       SeriesText = "West Group A"      // In season tournament
	SeriesTextISTWestGroupB       SeriesText = "West Group B"      // In season tournament
	SeriesTextISTWestGroupC       SeriesText = "West Group C"      // In season tournament
	SeriesTextISTWestQuarterFinal SeriesText = "West Quarterfinal" // In season tournament
	SeriesTextISTWestSemiFinal    SeriesText = "West Semifinal"    // In season tournament
)

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
		SeriesText        SeriesText     `json:"seriesText"`
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

func (c Client) GetTodaysScoreboard(ctx context.Context, objectKey string) (TodaysScoreboard, error) {
	ctx, span := otel.Tracer("nba").Start(ctx, "nba.Client.GetTodaysScoreboard")
	defer span.End()

	req, err := retryablehttp.NewRequest(http.MethodGet, todaysScoreboardURL, nil)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return TodaysScoreboard{}, fmt.Errorf("failed to create request to get todays scoreboard: %w", err)
	}

	response, err := c.client.Do(req)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return TodaysScoreboard{}, fmt.Errorf("failed to get TodaysScoreboard object: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		err = fmt.Errorf("failed to successfully get TodaysScoreboard object")
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		if slices.Contains([]int{http.StatusNotFound, http.StatusForbidden}, response.StatusCode) {
			return TodaysScoreboard{}, ErrNotFound
		}
		return TodaysScoreboard{}, err
	}

	var respBody []byte
	respBody, err = io.ReadAll(response.Body)
	if err != nil {
		span.RecordError(err)
	}

	if c.Cache != nil {
		if err := c.Cache.PutObject(ctx, objectKey, bytes.NewReader(respBody)); err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return TodaysScoreboard{}, fmt.Errorf("failed to cache todays scoreboard object: %w", err)
		}
	}

	var scoreboard TodaysScoreboard
	if err := json.Unmarshal(respBody, &scoreboard); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return TodaysScoreboard{}, fmt.Errorf("failed to unmarshal league schedule json: %w", err)
	}

	return scoreboard, nil
}
