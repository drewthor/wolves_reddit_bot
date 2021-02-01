package resources

import (
	"net/http"

	"github.com/drewthor/wolves_reddit_bot/util"

	"github.com/drewthor/wolves_reddit_bot/services"
	"github.com/go-chi/chi"
)

type GameResource struct {
	GameService *services.GameService
}

func (gr GameResource) Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", gr.List)

	r.Route("/{gameID}", func(r chi.Router) {
		r.Mount("/boxscore", BoxscoreResource{BoxscoreService: &services.BoxscoreService{}}.Routes())
	})

	return r
}

func (gr GameResource) List(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	gameDate := r.FormValue("game_date")

	if gameDate == "" {
		util.WriteJSON(http.StatusBadRequest, "invalid request: missing game_date", w)
		return
	}

	util.WriteJSON(http.StatusOK, gr.GameService.Get(gameDate), w)

}
