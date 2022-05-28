-- name: InsertArena :one
INSERT INTO arena
    as a(name, city, state, country, nba_arena_id)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (nba_arena_id) DO UPDATE
SET
    name = coalesce(excluded.name, a.name),
    city = coalesce(excluded.city, a.city),
    state = coalesce(excluded.state, a.state),
    country = coalesce(excluded.country, a.country),
    nba_arena_id = coalesce(excluded.nba_arena_id, a.nba_arena_id)
RETURNING *;
