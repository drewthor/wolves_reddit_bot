package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/drewthor/wolves_reddit_bot/dao"

	"github.com/drewthor/wolves_reddit_bot/service"
	"github.com/jackc/pgx/v4/pgxpool"

	"github.com/drewthor/wolves_reddit_bot/controller"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

func main() {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	dbConfig, err := pgxpool.ParseConfig(os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Println("could not create db config")
		os.Exit(1)
	}

	dbConfig.BeforeConnect = dao.RegisterCustomTypes

	dbpool, err := pgxpool.ConnectConfig(context.Background(), dbConfig)
	if err != nil {
		log.Println(os.Stderr, "unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer dbpool.Close()

	r.Mount("/games", controller.GameController{GameService: &service.GameService{}}.Routes())
	r.Mount("/players", controller.PlayerController{PlayerService: &service.PlayerService{PlayerDAO: &dao.PlayerDAO{DB: dbpool}}}.Routes())
	r.Mount("/teams", controller.TeamController{TeamService: &service.TeamService{TeamDAO: &dao.TeamDAO{DB: dbpool}}}.Routes())

	http.ListenAndServe(":3333", r)
}
