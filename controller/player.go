package controller

import (
	"net/http"

	"github.com/drewthor/wolves_reddit_bot/util"

	"github.com/drewthor/wolves_reddit_bot/service"
	"github.com/go-chi/chi"
)

type PlayerController struct {
	PlayerService *service.PlayerService
}

func (pc PlayerController) Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", pc.List)

	r.Get("/{id}", pc.Get)

	r.Post("/update", pc.UpdatePlayers)

	return r
}

func (pc PlayerController) List(w http.ResponseWriter, r *http.Request) {
	players, err := pc.PlayerService.GetAll()

	if err != nil {
		util.WriteJSON(http.StatusInternalServerError, err, w)

	}

	util.WriteJSON(http.StatusOK, players, w)
}

func (pc PlayerController) Get(w http.ResponseWriter, r *http.Request) {
	playerID := chi.URLParam(r, "id")

	player, err := pc.PlayerService.Get(playerID)

	if err != nil {
		util.WriteJSON(http.StatusInternalServerError, err, w)

	}

	util.WriteJSON(http.StatusOK, player, w)
}

func (pc PlayerController) UpdatePlayers(w http.ResponseWriter, r *http.Request) {
	players, err := pc.PlayerService.UpdatePlayers()

	if err != nil {
		util.WriteJSON(http.StatusInternalServerError, err, w)
	}

	util.WriteJSON(http.StatusOK, players, w)
}
