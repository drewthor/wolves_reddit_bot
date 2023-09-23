package postgres

import "github.com/jackc/pgx/v5/pgxpool"

//pgxpool"

type DB struct {
	pgxPool *pgxpool.Pool
}

func NewDB(pgxpool *pgxpool.Pool) DB {
	return DB{pgxPool: pgxpool}
}
