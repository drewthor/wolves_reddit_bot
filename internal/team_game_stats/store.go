package team_game_stats

import (
	"context"
	"database/sql"
	"time"
)

type Store interface {
	UpdateTeamGameStatsTotals(ctx context.Context, teamGameStatsTotalsUpdates []TeamGameStatsTotalUpdate) ([]TeamGameStatsTotal, error)
	UpdateTeamGameStatsTotalsOld(ctx context.Context, teamGameStatsTotalsUpdates []TeamGameStatsTotalUpdateOld) ([]TeamGameStatsTotal, error)
}

type TeamGameStatsTotalUpdate struct {
	NBAGameID                    string
	NBATeamID                    int
	GameTimePlayedSeconds        int
	TotalPlayerTimePlayedSeconds int
	Points                       int
	PointsAgainst                int
	Assists                      int
	PersonalTurnovers            int
	TeamTurnovers                int
	TotalTurnovers               int
	Steals                       int
	ThreePointersAttempted       int
	ThreePointersMade            int
	FieldGoalsAttempted          int
	FieldGoalsMade               int
	EffectiveAdjustedFieldGoals  float64
	FreeThrowsAttempted          int
	FreeThrowsMade               int
	Blocks                       int
	TimesBlocked                 int
	PersonalOffensiveRebounds    int
	PersonalDefensiveRebounds    int
	TotalPersonalRebounds        int
	TeamRebounds                 int
	TeamOffensiveRebounds        int
	TeamDefensiveRebounds        int
	TotalOffensiveRebounds       int
	TotalDefensiveRebounds       int
	TotalRebounds                int
	PersonalFouls                int
	OffensiveFouls               int
	FoulsDrawn                   int
	TeamFouls                    int
	PersonalTechnicalFouls       int
	TeamTechnicalFouls           int
	FullTimeoutsRemaining        int
	ShortTimeoutsRemaining       int
	TotalTimeoutsRemaining       int
	FastBreakPoints              int
	FastBreakPointsAttempted     int
	FastBreakPointsMade          int
	PointsInPaint                int
	PointsInPaintAttempted       int
	PointsInPaintMade            int
	SecondChancePoints           int
	SecondChancePointsAttempted  int
	SecondChancePointsMade       int
	PointsOffTurnovers           int
	BiggestLead                  int
	BiggestLeadScore             string
	BiggestScoringRun            int
	BiggestScoringRunScore       string
	TimeLeadingTenthSeconds      int
	LeadChanges                  int
	TimesTied                    int
	TrueShootingAttempts         float64
	TrueShootingPercentage       float64
	BenchPoints                  int
}

type TeamGameStatsTotalUpdateOld struct {
	NBAGameID                    string
	NBATeamID                    int
	GameTimePlayedSeconds        int
	TotalPlayerTimePlayedSeconds int
	Points                       int
	PointsAgainst                int
	Assists                      int
	TotalTurnovers               int
	Steals                       int
	ThreePointersAttempted       int
	ThreePointersMade            int
	FieldGoalsAttempted          int
	FieldGoalsMade               int
	FreeThrowsAttempted          int
	FreeThrowsMade               int
	Blocks                       int
	TimesBlocked                 int
	TotalOffensiveRebounds       int
	TotalDefensiveRebounds       int
	TotalRebounds                int
	PersonalFouls                int
	TeamFouls                    int
	TotalTimeoutsRemaining       int
	FastBreakPoints              int
	PointsInPaint                int
	SecondChancePoints           int
	PointsOffTurnovers           int
	BiggestLead                  int
	LeadChanges                  int
	TimesTied                    int
}

type TeamGameStatsTotal struct {
	ID                           string
	GameID                       string
	TeamID                       string
	CreatedAt                    time.Time
	UpdatedAt                    sql.NullTime
	GameTimePlayedSeconds        int
	TotalPlayerTimePlayedSeconds int // total time played on court by players. e.g. 12 min in quarter x 5 players -> 60 min
	Points                       int
	PointsAgainst                int
	Assists                      int
	PersonalTurnovers            int
	TeamTurnovers                int
	TotalTurnovers               int
	Steals                       int
	ThreePointersAttempted       int
	ThreePointersMade            int
	FieldGoalsAttempted          int
	FieldGoalsMade               int
	EffectiveAdjustedFieldGoals  float64
	FreeThrowsAttempted          int
	FreeThrowsMade               int
	Blocks                       int
	TimesBlocked                 int
	PersonalOffensiveRebounds    int
	PersonalDefensiveRebounds    int
	TotalPersonalRebounds        int
	TeamRebounds                 int
	TeamOffensiveRebounds        int
	TeamDefensiveRebounds        int
	TotalOffensiveRebounds       int
	TotalDefensiveRebounds       int
	TotalRebounds                int
	PersonalFouls                int
	OffensiveFouls               int
	FoulsDrawn                   int
	TeamFouls                    int
	PersonalTechnicalFouls       int
	TeamTechnicalFouls           int
	FullTimeoutsRemaining        int
	ShortTimeoutsRemaining       int
	TotalTimeoutsRemaining       int
	FastBreakPoints              int
	FastBreakPointsAttempted     int
	FastBreakPointsMade          int
	PointsInPaint                int
	PointsInPaintAttempted       int
	PointsInPaintMade            int
	SecondChancePoints           int
	SecondChancePointsAttempted  int
	SecondChancePointsMade       int
	PointsOffTurnovers           int
	BiggestLead                  int
	BiggestLeadScore             string
	BiggestScoringRun            int
	BiggestScoringRunScore       string
	TimeLeadingTenthSeconds      int
	LeadChanges                  int
	TimesTied                    int
	TrueShootingAttempts         float64
	TrueShootingPercentage       float64
	BenchPoints                  int
}
