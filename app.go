package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"
	_ "time/tzdata"

	"github.com/drewthor/wolves_reddit_bot/apis/cloudflare"
	"github.com/drewthor/wolves_reddit_bot/apis/nba"
	"github.com/drewthor/wolves_reddit_bot/internal/arena"
	"github.com/drewthor/wolves_reddit_bot/internal/boxscore"
	"github.com/drewthor/wolves_reddit_bot/internal/franchise"
	"github.com/drewthor/wolves_reddit_bot/internal/game"
	"github.com/drewthor/wolves_reddit_bot/internal/game_referee"
	"github.com/drewthor/wolves_reddit_bot/internal/league"
	"github.com/drewthor/wolves_reddit_bot/internal/playbyplay"
	"github.com/drewthor/wolves_reddit_bot/internal/player"
	"github.com/drewthor/wolves_reddit_bot/internal/r2"
	"github.com/drewthor/wolves_reddit_bot/internal/referee"
	"github.com/drewthor/wolves_reddit_bot/internal/scheduler"
	"github.com/drewthor/wolves_reddit_bot/internal/season"
	"github.com/drewthor/wolves_reddit_bot/internal/store/postgres"
	"github.com/drewthor/wolves_reddit_bot/internal/team"
	"github.com/drewthor/wolves_reddit_bot/internal/team_game_stats"
	"github.com/drewthor/wolves_reddit_bot/internal/team_season"
	"github.com/drewthor/wolves_reddit_bot/pkg/chimiddleware"
	"github.com/drewthor/wolves_reddit_bot/pkg/pgxutil"
	"github.com/drewthor/wolves_reddit_bot/pkg/rlhttp"
	"github.com/drewthor/wolves_reddit_bot/pkg/slogmiddleware"
	"github.com/drewthor/wolves_reddit_bot/util"
	"github.com/getsentry/sentry-go"
	sentryhttp "github.com/getsentry/sentry-go/http"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/tracelog"
	"github.com/joho/godotenv"
	"github.com/riandyrn/otelchi"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/trace"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	_ "github.com/KimMachineGun/automemlimit"
)

func main() {
	ctx := context.Background()

	handlerOpts := slog.HandlerOptions{
		AddSource:   true,
		Level:       slog.LevelInfo,
		ReplaceAttr: nil,
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &handlerOpts))

	defer func() { // catch or finally
		if err := recover(); err != nil { // catch
			logger.ErrorContext(ctx, "encountered panic: %v", err)
			os.Exit(1)
		}
	}()

	err := godotenv.Load()
	if err != nil {
		logger.Debug("Error loading .env file")
	}

	err = sentry.Init(sentry.ClientOptions{
		Dsn: os.Getenv("SENTRY_DSN"),
	})
	if err != nil {
		logger.ErrorContext(ctx, "error intializing sentrymiddleware", slog.Any("error", err))
		os.Exit(1)
	}

	logger = slog.New(slogmiddleware.NewOtelSlogHandler("log", []slog.Level{slog.LevelInfo, slog.LevelWarn, slog.LevelError}, slogmiddleware.NewSentrySlogHandler("error", []slog.Level{slog.LevelError}, logger.Handler())))
	// Flush buffered events before the program terminates.
	defer sentry.Flush(2 * time.Second)

	slog.SetDefault(logger)

	// TODO: explicitly set options here instead of implicit env vars https://github.com/open-telemetry/opentelemetry-go/tree/main/exporters/otlp/otlptrace

	// Configure a new exporter using environment variables for sending data to Honeycomb over gRPC.
	exporter, err := otlptracegrpc.New(ctx)
	if err != nil {
		logger.ErrorContext(ctx, "failed to initialize exporter", slog.Any("error", err))
		os.Exit(1)
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

	dbConfig, err := pgxpool.ParseConfig(os.Getenv("DATABASE_URL"))
	if err != nil {
		logger.ErrorContext(ctx, "could not create db config", slog.Any("error", err))
		os.Exit(1)
	}
	dbConfig.ConnConfig.Tracer = &tracelog.TraceLog{Logger: pgxutil.NewLogger(logger, pgxutil.WithErrorKey("error")), LogLevel: tracelog.LogLevelInfo}

	dbpool, err := pgxpool.NewWithConfig(context.Background(), dbConfig)
	if err != nil {
		logger.ErrorContext(ctx, "unable to create database pool")
		os.Exit(1)
	}
	err = dbpool.Ping(context.Background())
	if err != nil {
		logger.ErrorContext(ctx, "failed to connect to database", slog.Any("error", err))
		os.Exit(1)
	}
	logger.InfoContext(ctx, "database pool statistics", slog.Any("stats", *dbpool.Stat()))
	defer dbpool.Close()

	r2Client := cloudflare.NewClient(os.Getenv("CLOUDFLARE_ACCOUNT_ID"), os.Getenv("CLOUDFLARE_ACCESS_KEY_ID"), os.Getenv("CLOUDFLARE_ACCESS_KEY_SECRET"))
	if err := r2Client.CreateBucket(ctx, util.NBAR2Bucket); err != nil {
		logger.ErrorContext(ctx, "failed to create bucket for file in r2", slog.Any("error", err), slog.String("bucket", util.NBAR2Bucket))
	}
	r2ObjectCacher := r2.R2ObjectCacher{Bucket: "nba", R2Client: r2Client}

	nbaClient := nba.NewClient(r2ObjectCacher, rlhttp.WithLeveledLogger(logger))

	postgresStore := postgres.NewDB(dbpool)

	arenaService := arena.NewService(postgresStore)
	boxscoreService := boxscore.NewService()
	teamSeasonService := team_season.NewService(postgresStore, nbaClient)
	teamService := team.NewService(postgresStore, teamSeasonService, nbaClient)
	franchiseService := franchise.NewService(postgresStore, teamService, teamSeasonService, nbaClient)
	gameRefereeService := game_referee.NewService(postgresStore)
	leagueService := league.NewService(postgresStore)
	playByPlayService := playbyplay.NewService(nbaClient, r2Client, postgresStore)
	refereeService := referee.NewService(postgresStore)
	seasonService := season.NewService(postgresStore, nbaClient)
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
	schedulerService := scheduler.NewService(gameService, seasonService, nbaClient)
	schedulerService.Start(logger)
	defer schedulerService.Stop()

	sentryMiddleware := sentryhttp.New(sentryhttp.Options{
		Repanic: true,
	})

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(chimiddleware.NewStructuredLogger(logger.Handler()))
	r.Use(middleware.Recoverer)
	r.Use(sentryMiddleware.Handle)
	r.Use(otelchi.Middleware("nba", otelchi.WithChiRoutes(r)))

	r.Mount("/games", game.NewHandler(logger, gameService).Routes())
	r.Mount("/players", player.NewHandler(logger, playerService).Routes())
	r.Mount("/teams", team.NewHandler(logger, teamService).Routes())
	r.Mount("/boxscores", boxscore.NewHandler(logger, boxscoreService).Routes())
	r.Mount("/franchises", franchise.NewHandler(logger, franchiseService).Routes())

	logger.InfoContext(ctx, "starting http server")

	err = http.ListenAndServe(":3333", r)
	if err != nil {
		logger.ErrorContext(ctx, "failed to run http server", slog.Any("error", err))
	}
	logger.InfoContext(ctx, "shutting down server")
}
