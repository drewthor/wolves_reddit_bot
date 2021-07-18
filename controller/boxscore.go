package controller

import (
	"net/http"

	"github.com/go-chi/chi"

	"github.com/drewthor/wolves_reddit_bot/util"

	"github.com/drewthor/wolves_reddit_bot/service"
)

type BoxscoreController struct {
	BoxscoreService *service.BoxscoreService
}

func (bc BoxscoreController) Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", bc.Get)

	return r
}

func (bc BoxscoreController) Get(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	gameID := chi.URLParam(r, "gameID")
	gameDate := r.FormValue("game_date")

	if gameID == "" {
		util.WriteJSON(http.StatusBadRequest, "invalid request: missing game_id", w)
		return
	}

	if gameDate == "" {
		util.WriteJSON(http.StatusBadRequest, "invalid request: missing game_date", w)
		return
	}

	boxscore, err := bc.BoxscoreService.Get(gameID, gameDate)
	if err != nil {
		util.WriteJSON(http.StatusInternalServerError, err, w)
		return
	}

	util.WriteJSON(http.StatusOK, boxscore, w)
}
