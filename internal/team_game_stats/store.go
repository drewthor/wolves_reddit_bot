package team_game_stats

import (
	"context"

	log "github.com/sirupsen/logrus"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type Store interface {
	UpdateTeamGameStatsTotals(teamGameStatsTotalsUpdates []TeamGameStatsTotalUpdate) ([]TeamGameStatsTotal, error)
}

func NewStore(db *pgxpool.Pool) Store {
	return &store{DB: db}
}

type store struct {
	DB *pgxpool.Pool
}

type TeamGameStatsTotalUpdate struct {
	NBAGameID              string
	NBATeamID              string
	TimePlayedSeconds      int
	Points                 int
	Assists                int
	Turnovers              int
	Steals                 int
	ThreePointersAttempted int
	ThreePointersMade      int
	ThreePointPercentage   float64
	FieldGoalsAttempted    int
	FieldGoalsMade         int
	FieldGoalPercentage    float64
	FreeThrowsAttempted    int
	FreeThrowsMade         int
	FreeThrowPercentage    float64
	Blocks                 int
	ReboundsOffensive      int
	ReboundsDefensive      int
	ReboundsTotal          int
	FoulsPersonal          int
	FoulsTeam              int
	TimeoutsFull           int
	TimeoutsShort          int
	PointsFastBreak        int
	// Points
}

type TeamGameStatsTotal struct {
	GameID                 string
	TeamID                 string
	TimePlayedSeconds      int
	Points                 int
	Assists                int
	Turnovers              int
	Steals                 int
	ThreePointersAttempted int
	ThreePointersMade      int
	FieldGoalsAttempted    int
	FieldGoalsMade         int
	FreeThrowsAttempted    int
	FreeThrowsMade         int
	Blocks                 int
	OffensiveRebounds      int
	DefensiveRebounds      int
	TotalRebounds          int
}

func (s *store) UpdateTeamGameStatsTotals(teamGameStatsTotalsUpdates []TeamGameStatsTotalUpdate) ([]TeamGameStatsTotal, error) {
	tx, err := s.DB.Begin(context.Background())
	defer tx.Rollback(context.Background())
	if err != nil {
		log.Printf("could not start db transaction with error: %v", err)
		return nil, err
	}

	// insertTeamGameStatsTotal := `
	_ = `
		INSERT INTO nba.team_game_stats_total
			as tgst(game_id, team_id, time_played_seconds, points, assists, turnovers, steals, three_pointers_attempted, three_pointers_made, three_point_percentage, , city, state, country)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (name) DO UPDATE
		SET
			name = coalesce(excluded.name, tgst.name),
			city = coalesce(excluded.city, tgst.city),
			state = coalesce(excluded.state, tgst.state),
			country = coalesce(excluded.country, tgst.country)
		RETURNING tgst.id, a.name, a.city, a.state, a.country, a.created_at, a.updated_at`

	bp := &pgx.Batch{}

	// for _, teamGameStatsTotalUpdate := range teamGameStatsTotalsUpdates {
	// 	bp.Queue(insertTeamGameStatsTotal,
	// 		arena.Name,
	// 		arena.City,
	// 		arena.State,
	// 		arena.Country)
	// }

	batchResults := tx.SendBatch(context.Background(), bp)

	insertedTeamGameStatsTotals := []TeamGameStatsTotal{}

	for _, _ = range teamGameStatsTotalsUpdates {
		teamGameStatsTotal := TeamGameStatsTotal{}

		// err := batchResults.QueryRow().Scan(&arena.ID,
		// 	&arena.Name,
		// 	&arena.City,
		// 	&arena.State,
		// 	&arena.Country,
		// 	&arena.CreatedAt,
		// 	&arena.UpdatedAt)

		if err != nil {
			return nil, err
		}

		insertedTeamGameStatsTotals = append(insertedTeamGameStatsTotals, teamGameStatsTotal)
	}

	err = batchResults.Close()
	if err != nil {
		return nil, err
	}

	err = tx.Commit(context.Background())
	if err != nil {
		return nil, err
	}

	return insertedTeamGameStatsTotals, nil
}
