package boxscore

import (
	"net/http"

	"github.com/drewthor/wolves_reddit_bot/util"
	"github.com/go-chi/chi"
	log "github.com/sirupsen/logrus"
)

type Handler interface {
	Routes() chi.Router
	Get(w http.ResponseWriter, r *http.Request)
}

func NewHandler(boxscoreService Service) Handler {
	return &handler{BoxscoreService: boxscoreService}
}

type handler struct {
	BoxscoreService Service
}

func (h *handler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", h.Get)

	return r
}

func (h *handler) Get(w http.ResponseWriter, r *http.Request) {
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

	boxscore, err := h.BoxscoreService.Get(gameID, gameDate)
	if err != nil {
		log.Error(err)
		util.WriteJSON(http.StatusInternalServerError, err, w)
		return
	}

	util.WriteJSON(http.StatusOK, boxscore, w)
}
