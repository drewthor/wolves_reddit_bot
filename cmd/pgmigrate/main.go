package main

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/getsentry/sentry-go"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		slog.Debug("Error loading .env file")
	}

	err = sentry.Init(sentry.ClientOptions{
		Dsn: os.Getenv("SENTRY_DSN"),
	})
	if err != nil {
		slog.Error("error intializing sentrymiddleware", slog.Any("error", err))
	}

	// Flush buffered events before the program terminates.
	defer sentry.Flush(2 * time.Second)

	m, err := migrate.New(
		fmt.Sprintf("file://%s", os.Getenv("DB_MIGRATIONS_DIR")),
		os.Getenv("DATABASE_URL"))
	if err != nil {
		slog.Error("failed to create db migration instance", slog.Any("error", err))
	}
	oldVersion, _, err := m.Version()
	if err != nil {
		slog.Error("failed to get old schema_migrations version before performing db migrations", slog.Any("error", err))
	}
	if err := m.Up(); err != nil {
		slog.Error("failed to run up migrations", slog.Any("error", err))
	}
	newVersion, _, err := m.Version()
	if err != nil {
		slog.Error("failed to get new schema_migrations version after performing db migrations", slog.Any("error", err))
	}
	slog.Info(fmt.Sprintf("successfully migrated from version %d to %d", oldVersion, newVersion))
}
