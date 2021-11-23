package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/drewthor/wolves_reddit_bot/schedule"

	"github.com/drewthor/wolves_reddit_bot/dao"

	"github.com/drewthor/wolves_reddit_bot/service"
	"github.com/jackc/pgx/v4/pgxpool"

	"github.com/drewthor/wolves_reddit_bot/controller"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

func main() {
	s := schedule.NewScheduler()

	s.StartAsync()

	r := chi.NewRouter()
	r.Use(middleware.Heartbeat("/ping"))
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	dbConfig, err := pgxpool.ParseConfig(os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Println("could not create db config")
		os.Exit(1)
	}

	dbpool, err := pgxpool.ConnectConfig(context.Background(), dbConfig)
	if err != nil {
		log.Println(os.Stderr, "unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer dbpool.Close()

	arenaService := &service.ArenaService{ArenaDAO: &dao.ArenaDAO{DB: dbpool}}
	gameRefereeService := &service.GameRefereeService{GameRefereeDAO: &dao.GameRefereeDAO{DB: dbpool}}
	refereeService := &service.RefereeService{RefereeDAO: &dao.RefereeDAO{DB: dbpool}}
	seasonStageService := &service.SeasonStageService{}
	teamSeasonService := &service.TeamSeasonService{TeamSeasonDao: &dao.TeamSeasonDAO{DB: dbpool}}
	teamService := &service.TeamService{TeamDAO: &dao.TeamDAO{DB: dbpool}, TeamSeasonService: teamSeasonService}
	gameService := &service.GameService{
		GameDAO:            &dao.GameDAO{DB: dbpool},
		ArenaService:       arenaService,
		GameRefereeService: gameRefereeService,
		RefereeService:     refereeService,
		SeasonStageService: seasonStageService,
		TeamService:        teamService}

	r.Mount("/games", controller.GameController{GameService: gameService}.Routes())
	r.Mount("/players", controller.PlayerController{PlayerService: &service.PlayerService{PlayerDAO: &dao.PlayerDAO{DB: dbpool}}}.Routes())
	r.Mount("/teams", controller.TeamController{TeamService: teamService}.Routes())

	http.ListenAndServe(":3333", r)
}
