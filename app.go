package main

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/drewthor/wolves_reddit_bot/internal/arena"
	"github.com/drewthor/wolves_reddit_bot/internal/game"
	"github.com/drewthor/wolves_reddit_bot/internal/game_referee"
	"github.com/drewthor/wolves_reddit_bot/internal/player"
	"github.com/drewthor/wolves_reddit_bot/internal/referee"
	"github.com/drewthor/wolves_reddit_bot/internal/team"
	"github.com/drewthor/wolves_reddit_bot/internal/team_season"
	"github.com/drewthor/wolves_reddit_bot/schedule"
	"github.com/getsentry/sentry-go"
	sentryhttp "github.com/getsentry/sentry-go/http"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/log/logrusadapter"
	log "github.com/sirupsen/logrus"

	"github.com/jackc/pgx/v4/pgxpool"

	sentryHook "github.com/drewthor/wolves_reddit_bot/pkg/sentry"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

func main() {
	err := sentry.Init(sentry.ClientOptions{
		Dsn: os.Getenv("SENTRY_DSN"),
	})
	if err != nil {
		log.Fatalf("error intializing sentry: %s", err)
	}

	// Flush buffered events before the program terminates.
	defer sentry.Flush(2 * time.Second)

	log.AddHook(sentryHook.NewHook([]log.Level{log.ErrorLevel, log.FatalLevel, log.PanicLevel}))

	s := schedule.NewScheduler()

	s.StartAsync()

	sentryMiddleware := sentryhttp.New(sentryhttp.Options{
		Repanic: true,
	})

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(sentryMiddleware.Handle)

	dbConfig, err := pgxpool.ParseConfig(os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Println("could not create db config")
		os.Exit(1)
	}
	dbConfig.ConnConfig.Logger = logrusadapter.NewLogger(log.StandardLogger())
	dbConfig.ConnConfig.LogLevel = pgx.LogLevelWarn

	dbpool, err := pgxpool.ConnectConfig(context.Background(), dbConfig)
	if err != nil {
		log.Println(os.Stderr, "unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer dbpool.Close()

	arenaService := arena.NewService(arena.NewStore(dbpool))
	gameRefereeService := game_referee.NewService(game_referee.NewStore(dbpool))
	refereeService := referee.NewService(referee.NewStore(dbpool))
	teamSeasonService := team_season.NewService(team_season.NewStore(dbpool))
	teamService := team.NewService(team.NewStore(dbpool), teamSeasonService)
	gameService := game.NewService(
		game.NewStore(dbpool),
		arenaService,
		gameRefereeService,
		refereeService,
		teamService,
	)
	playerService := player.NewService(player.NewStore(dbpool))

	r.Mount("/games", game.NewHandler(gameService).Routes())
	r.Mount("/players", player.NewHandler(playerService).Routes())
	r.Mount("/teams", team.NewHandler(teamService).Routes())

	http.ListenAndServe(":3333", r)
}
