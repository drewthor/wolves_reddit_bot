package controller

import (
	"net/http"

	"github.com/drewthor/wolves_reddit_bot/util"

	"github.com/drewthor/wolves_reddit_bot/service"
	"github.com/go-chi/chi"
)

type GameController struct {
	GameService *service.GameService
}

func (gc GameController) Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", gc.List)

	r.Route("/{gameID}", func(r chi.Router) {
		r.Mount("/boxscore", BoxscoreController{BoxscoreService: &service.BoxscoreService{}}.Routes())
	})

	r.Post("/update", gc.UpdateGames)

	return r
}

func (gc GameController) List(w http.ResponseWriter, r *http.Request) {
	gameDate := chi.URLParam(r, "game-date")

	if gameDate == "" {
		util.WriteJSON(http.StatusBadRequest, "invalid request: missing game_date", w)
		return
	}

	games, err := gc.GameService.Get(gameDate)
	if err != nil {
		util.WriteJSON(http.StatusInternalServerError, err, w)
	}

	util.WriteJSON(http.StatusOK, games, w)
}

func (gc GameController) UpdateGames(w http.ResponseWriter, r *http.Request) {
	games, err := gc.GameService.UpdateGames()

	if err != nil {
		util.WriteJSON(http.StatusInternalServerError, err, w)
	}

	util.WriteJSON(http.StatusOK, games, w)
}
