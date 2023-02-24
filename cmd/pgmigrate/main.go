package main

import (
	"fmt"
	"os"
	"time"

	"github.com/getsentry/sentry-go"
	log "github.com/sirupsen/logrus"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Debug("Error loading .env file")
	}

	err = sentry.Init(sentry.ClientOptions{
		Dsn: os.Getenv("SENTRY_DSN"),
	})
	if err != nil {
		log.Fatalf("error intializing sentry: %s", err)
	}

	// Flush buffered events before the program terminates.
	defer sentry.Flush(2 * time.Second)

	//log.AddHook(sentryHook.NewHook([]log.Level{log.ErrorLevel, log.FatalLevel, log.PanicLevel}))

	m, err := migrate.New(
		fmt.Sprintf("file://%s", os.Getenv("DB_MIGRATIONS_DIR")),
		os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}
	oldVersion, _, err := m.Version()
	if err != nil {
		log.Fatal(fmt.Errorf("failed to get old schema_migrations version before performing db migrations: %w", err))
	}
	if err := m.Up(); err != nil {
		log.Fatal(err)
	}
	newVersion, _, err := m.Version()
	if err != nil {
		log.Fatal(fmt.Errorf("failed to get new schema_migrations version after performing db migrations: %w", err))
	}
	log.Infof("successfully migrated from version %d to %d", oldVersion, newVersion)
}
