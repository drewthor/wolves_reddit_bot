package main

import (
	"net/http"

	"github.com/drewthor/wolves_reddit_bot/services"

	"github.com/drewthor/wolves_reddit_bot/resources"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

func main() {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Mount("/games", resources.GameResource{GameService: &services.GameService{}}.Routes())

	http.ListenAndServe(":3333", r)
}
