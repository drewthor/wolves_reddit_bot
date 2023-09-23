package postgres

import (
	"context"
	"fmt"

	"github.com/drewthor/wolves_reddit_bot/api"
	"github.com/jackc/pgx/v5"
	"go.opentelemetry.io/otel"
)

func (d DB) GetPlayerWithID(ctx context.Context, playerID string) (api.Player, error) {
	ctx, span := otel.Tracer("postgres").Start(ctx, "postgres.DB.GetPlayerWithID")
	defer span.End()

	player := api.Player{}

	query := `
		SELECT id, first_name, last_name, birthdate, height_feet, height_inches, height_meters, weight_pounds, weight_kilograms, jersey_number, positions.pos_array, active, years_pro, nba_debut_year, nba_player_id, country, created_at, updated_at 
		FROM nba.player p, LATERAL (
		        SELECT ARRAY (
		            SELECT pos.name 
					FROM nba.position pos
		            JOIN nba.player_position pp ON pos.id = pp.position_id
		            WHERE pp.player_id = p.id
		            ORDER BY pp.priority asc
                ) as pos_array
        ) positions
		WHERE p.id = $1`

	err := d.pgxPool.QueryRow(ctx, query, playerID).Scan(
		&player.ID,
		&player.FirstName,
		&player.LastName,
		&player.Birthdate,
		&player.HeightFeet,
		&player.HeightInches,
		&player.HeightMeters,
		&player.WeightPounds,
		&player.WeightKilograms,
		&player.JerseyNumber,
		&player.Positions,
		&player.Active,
		&player.YearsPro,
		&player.NBADebutYear,
		&player.NBAPlayerID,
		&player.Country,
		&player.CreatedAt,
		&player.UpdatedAt)
	if err != nil {
		return api.Player{}, err
	}

	return player, nil
}

func (d DB) ListPlayers(ctx context.Context) ([]api.Player, error) {
	ctx, span := otel.Tracer("nba").Start(ctx, "postgres.DB.ListPlayers")
	defer span.End()

	query := `
		SELECT id, first_name, last_name, birthdate, height_feet, height_inches, height_meters, weight_pounds, weight_kilograms, jersey_number, positions.pos_array, active, years_pro, nba_debut_year, nba_player_id, country, created_at, updated_at 
		FROM nba.player p, LATERAL (
		        SELECT ARRAY (
		            SELECT pos.name 
					FROM nba.position pos
		            JOIN nba.player_position pp ON pos.id = pp.position_id
		            WHERE pp.player_id = p.id
		            ORDER BY pp.priority asc
                ) as pos_array
        ) positions`

	rows, err := d.pgxPool.Query(ctx, query)
	if err != nil {
		return nil, err
	}

	players := []api.Player{}

	for rows.Next() {
		player := api.Player{}
		err = rows.Scan(
			&player.ID,
			&player.FirstName,
			&player.LastName,
			&player.Birthdate,
			&player.HeightFeet,
			&player.HeightInches,
			&player.HeightMeters,
			&player.WeightPounds,
			&player.WeightKilograms,
			&player.JerseyNumber,
			&player.Positions,
			&player.Active,
			&player.YearsPro,
			&player.NBADebutYear,
			&player.NBAPlayerID,
			&player.Country,
			&player.CreatedAt,
			&player.UpdatedAt)
		if err != nil {
			return nil, err
		}
		players = append(players, player)
	}

	return players, nil
}

func (d DB) GetPlayersWithIDs(ctx context.Context, ids []string) ([]api.Player, error) {
	ctx, span := otel.Tracer("nba").Start(ctx, "postgres.DB.GetPlayersWithIDs")
	defer span.End()

	query := `
		SELECT id, first_name, last_name, birthdate, height_feet, height_inches, height_meters, weight_pounds, weight_kilograms, jersey_number, positions.pos_array, active, years_pro, nba_debut_year, nba_player_id, country, created_at, updated_at 
		FROM nba.player p, LATERAL (
		        SELECT ARRAY (
		            SELECT pos.name 
					FROM nba.position pos
		            JOIN nba.player_position pp ON pos.id = pp.position_id
		            WHERE pp.player_id = p.id
		            ORDER BY pp.priority asc
                ) as pos_array
        ) positions
		WHERE id = ANY($1)`

	rows, err := d.pgxPool.Query(ctx, query, ids)
	if err != nil {
		return nil, err
	}

	players := []api.Player{}

	for rows.Next() {
		player := api.Player{}
		err = rows.Scan(
			&player.ID,
			&player.FirstName,
			&player.LastName,
			&player.Birthdate,
			&player.HeightFeet,
			&player.HeightInches,
			&player.HeightMeters,
			&player.WeightPounds,
			&player.WeightKilograms,
			&player.JerseyNumber,
			&player.Positions,
			&player.Active,
			&player.YearsPro,
			&player.NBADebutYear,
			&player.NBAPlayerID,
			&player.Country,
			&player.CreatedAt,
			&player.UpdatedAt)
		if err != nil {
			return nil, err
		}
		players = append(players, player)
	}

	return players, nil
}

func (d DB) UpdatePlayers(ctx context.Context, players []api.Player) ([]api.Player, error) {
	ctx, span := otel.Tracer("nba").Start(ctx, "postgres.DB.UpdatePlayers")
	defer span.End()

	tx, err := d.pgxPool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not start db transaction with error: %w", err)
	}
	defer tx.Rollback(ctx)

	insertPlayer := `
						INSERT INTO nba.player
							as p(first_name, last_name, birthdate, height_feet, height_inches, height_meters, weight_pounds, weight_kilograms, jersey_number, active, years_pro, nba_debut_year, nba_player_id, country)
						VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
						ON CONFLICT (nba_player_id) DO UPDATE
						SET 
							first_name = excluded.first_name,
							last_name = excluded.last_name,
							birthdate = coalesce(excluded.birthdate, p.birthdate),
							height_feet = coalesce(excluded.height_feet, p.height_feet),
							height_inches = coalesce(excluded.height_inches, p.height_inches),
							height_meters = coalesce(excluded.height_meters, p.height_meters),
							weight_pounds = coalesce(excluded.weight_pounds, p.weight_pounds),
							weight_kilograms = coalesce(excluded.weight_kilograms, p.weight_kilograms),
							jersey_number = coalesce(excluded.jersey_number, p.jersey_number),
							active = excluded.active,
							years_pro = excluded.years_pro,
							nba_debut_year = coalesce(excluded.nba_debut_year, p.nba_debut_year),
							nba_player_id = excluded.nba_player_id,
							country = coalesce(excluded.country, p.country)
						RETURNING p.id`

	removeExistingPlayerPositions := `
	DELETE FROM nba.player_position
	WHERE player_id = $1`

	insertPlayerPositions := `
		INSERT INTO nba.player_position
			as pp(player_id, position_id, priority)
		VALUES ($1, (SELECT id FROM nba.position WHERE name = $2), $3)`

	bp := &pgx.Batch{}

	for _, player := range players {
		bp.Queue(insertPlayer,
			player.FirstName,
			player.LastName,
			player.Birthdate,
			player.HeightFeet,
			player.HeightInches,
			player.HeightMeters,
			player.WeightPounds,
			player.WeightKilograms,
			player.JerseyNumber,
			player.Active,
			player.YearsPro,
			player.NBADebutYear,
			player.NBAPlayerID,
			player.Country)
	}

	bpp := &pgx.Batch{}

	batchResults := tx.SendBatch(ctx, bp)

	insertedPlayerIDs := []string{}
	numPlayerPositions := 0

	for _, player := range players {
		id := ""
		err := batchResults.QueryRow().Scan(&id)
		if err != nil {
			batchResults.Close()
			return nil, err
		}

		insertedPlayerIDs = append(insertedPlayerIDs, id)

		bpp.Queue(removeExistingPlayerPositions, id)
		for j := range player.Positions {
			bpp.Queue(insertPlayerPositions, id, player.Positions[j], j)
			numPlayerPositions++
		}
	}

	err = batchResults.Close()
	if err != nil {
		return nil, err
	}

	batchResults = tx.SendBatch(ctx, bpp)

	for i := 0; i < numPlayerPositions; i++ {
		_, err = batchResults.Exec()
		if err != nil {
			batchResults.Close()
			return nil, err
		}
	}

	err = batchResults.Close()
	if err != nil {
		return nil, err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, err
	}

	return d.GetPlayersWithIDs(ctx, insertedPlayerIDs)
}
