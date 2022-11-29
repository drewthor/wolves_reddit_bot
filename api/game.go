package api

import (
	"time"
)

type Game struct {
	ID                  string     `json:"id"`
	HomeTeamID          *string    `json:"home_team_id"`
	AwayTeamID          *string    `json:"away_team_id"`
	HomeTeamPoints      *int       `json:"home_team_points"`
	AwayTeamPoints      *int       `json:"away_team_points"`
	Status              string     `json:"status"`
	ArenaID             *string    `json:"arena_id"`
	Attendance          int        `json:"attendance"`
	Season              string     `json:"season"`
	SeasonStage         string     `json:"season_stage"`
	Period              int        `json:"period"`
	PeriodTimeRemaining int        `json:"period_time_remaining"`
	Duration            *int       `json:"duration"`
	StartTime           time.Time  `json:"start_time"`
	EndTime             *time.Time `json:"end_time"`
	NBAGameID           string     `json:"nba_game_id"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           *time.Time `json:"updated_at"`
}
