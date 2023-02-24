package main

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/drewthor/wolves_reddit_bot/apis/cloudflare"
	"github.com/drewthor/wolves_reddit_bot/apis/nba"
	"github.com/drewthor/wolves_reddit_bot/internal/arena"
	"github.com/drewthor/wolves_reddit_bot/internal/boxscore"
	"github.com/drewthor/wolves_reddit_bot/internal/game"
	"github.com/drewthor/wolves_reddit_bot/internal/game_referee"
	"github.com/drewthor/wolves_reddit_bot/internal/league"
	"github.com/drewthor/wolves_reddit_bot/internal/playbyplay"
	"github.com/drewthor/wolves_reddit_bot/internal/player"
	"github.com/drewthor/wolves_reddit_bot/internal/referee"
	"github.com/drewthor/wolves_reddit_bot/internal/schedule"
	"github.com/drewthor/wolves_reddit_bot/internal/season"
	"github.com/drewthor/wolves_reddit_bot/internal/store/postgres"
	"github.com/drewthor/wolves_reddit_bot/internal/team"
	"github.com/drewthor/wolves_reddit_bot/internal/team_game_stats"
	"github.com/drewthor/wolves_reddit_bot/internal/team_season"
	sentryHook "github.com/drewthor/wolves_reddit_bot/pkg/sentry"
	"github.com/drewthor/wolves_reddit_bot/util"
	"github.com/getsentry/sentry-go"
	sentryhttp "github.com/getsentry/sentry-go/http"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/log/logrusadapter"
	"github.com/joho/godotenv"
	"github.com/riandyrn/otelchi"
	log "github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/trace"

	"github.com/jackc/pgx/v4/pgxpool"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	defer func() { //catch or finally
		if err := recover(); err != nil { //catch
			log.Error("encountered panic: %v", err)
			os.Exit(1)
		}
	}()

	err := godotenv.Load()
	if err != nil {
		log.Debug("Error loading .env file")
	}

	ctx := context.Background()

	// Configure a new exporter using environment variables for sending data to Honeycomb over gRPC.
	exporter, err := otlptracegrpc.New(ctx)
	if err != nil {
		log.Fatalf("failed to initialize exporter: %v", err)
	}

	// Create a new tracer provider with a batch span processor and the otlp exporter.
	tp := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
	)

	// Handle shutdown errors in a sensible manner where possible
	defer func() { _ = tp.Shutdown(ctx) }()

	// Set the Tracer Provider global
	otel.SetTracerProvider(tp)

	// Register the trace context and baggage propagators so data is propagated across services/processes.
	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		),
	)

	err = sentry.Init(sentry.ClientOptions{
		Dsn: os.Getenv("SENTRY_DSN"),
	})
	if err != nil {
		log.Fatalf("error intializing sentry: %s", err)
	}

	// Flush buffered events before the program terminates.
	defer sentry.Flush(2 * time.Second)

	log.AddHook(sentryHook.NewHook([]log.Level{log.ErrorLevel, log.FatalLevel, log.PanicLevel}))

	dbConfig, err := pgxpool.ParseConfig(os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Error("could not create db config")
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

	r2Client := cloudflare.NewClient(os.Getenv("CLOUDFLARE_ACCOUNT_ID"), os.Getenv("CLOUDFLARE_ACCESS_KEY_ID"), os.Getenv("CLOUDFLARE_ACCESS_KEY_SECRET"))
	if err := r2Client.CreateBucket(ctx, util.NBAR2Bucket); err != nil {
		log.WithError(err).WithFields(log.Fields{"bucket": util.NBAR2Bucket}).Fatal("failed to create bucket for file in r2")
	}

	nbaClient := nba.NewClient()

	postgresStore := postgres.NewDB(dbpool)

	arenaService := arena.NewService(postgresStore)
	boxscoreService := boxscore.NewService()
	gameRefereeService := game_referee.NewService(postgresStore)
	leagueService := league.NewService(postgresStore)
	playByPlayService := playbyplay.NewService(nbaClient, r2Client, postgresStore)
	refereeService := referee.NewService(postgresStore)
	seasonService := season.NewService(postgresStore, r2Client)
	teamSeasonService := team_season.NewService(postgresStore)
	teamService := team.NewService(postgresStore, teamSeasonService, nbaClient)
	teamGameStatsService := team_game_stats.NewService(postgresStore)
	gameService := game.NewService(
		postgresStore,
		arenaService,
		gameRefereeService,
		leagueService,
		playByPlayService,
		refereeService,
		seasonService,
		teamService,
		teamGameStatsService,
		nbaClient,
		r2Client,
	)
	playerService := player.NewService(postgresStore)
	schedulerService := schedule.NewService(gameService, seasonService, r2Client)
	schedulerService.Start()
	defer schedulerService.Stop()

	sentryMiddleware := sentryhttp.New(sentryhttp.Options{
		Repanic: true,
	})

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(sentryMiddleware.Handle)
	r.Use(otelchi.Middleware("nba", otelchi.WithChiRoutes(r)))

	r.Mount("/games", game.NewHandler(gameService).Routes())
	r.Mount("/players", player.NewHandler(playerService).Routes())
	r.Mount("/teams", team.NewHandler(teamService).Routes())
	r.Mount("/boxscores", boxscore.NewHandler(boxscoreService).Routes())

	log.Info("starting http server")

	err = http.ListenAndServe(":3333", r)
	if err != nil {
		log.WithError(err).Error("failed to run http server")
	}
	log.Info("shutting down server")
}
