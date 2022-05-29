package game

import (
	"net/http"
	"strconv"

	"github.com/drewthor/wolves_reddit_bot/internal/boxscore"
	"github.com/drewthor/wolves_reddit_bot/util"
	log "github.com/sirupsen/logrus"

	"github.com/go-chi/chi"
)

type Handler interface {
	Routes() chi.Router
	List(w http.ResponseWriter, r *http.Request)
	UpdateGames(w http.ResponseWriter, r *http.Request)
}

func NewHandler(gameService Service) Handler {
	return &handler{GameService: gameService}
}

type handler struct {
	GameService Service
}

func (h *handler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", h.List)

	r.Route("/{gameID}", func(r chi.Router) {
		r.Mount("/boxscore", boxscore.NewHandler(boxscore.NewService()).Routes())
	})

	r.Post("/update", h.UpdateGames)

	return r
}

func (h *handler) List(w http.ResponseWriter, r *http.Request) {
	gameDate := r.URL.Query().Get("game-date")

	if gameDate == "" {
		util.WriteJSON(http.StatusBadRequest, "invalid request: missing game_date", w)
		return
	}

	games, err := h.GameService.Get(gameDate)
	if err != nil {
		log.WithError(err).Errorf("failed to get games for date: %s", gameDate)
		util.WriteJSON(http.StatusInternalServerError, err, w)
		return
	}

	util.WriteJSON(http.StatusOK, games, w)
}

func (h *handler) UpdateGames(w http.ResponseWriter, r *http.Request) {
	seasonStartYear, err := strconv.Atoi(r.URL.Query().Get("season-start-year"))
	if err != nil {
		util.WriteJSON(http.StatusBadRequest, "invalid required season-start-year", w)
		return
	}

	games, err := h.GameService.UpdateGames(seasonStartYear)
	if err != nil {
		log.WithError(err).Errorf("could not update games for season-start-year: %d", seasonStartYear)
		util.WriteJSON(http.StatusInternalServerError, err, w)
		return
	}

	util.WriteJSON(http.StatusOK, games, w)
}
