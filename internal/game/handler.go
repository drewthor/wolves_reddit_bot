package game

import (
	"net/http"
	"strconv"

	"github.com/drewthor/wolves_reddit_bot/internal/boxscore"
	"github.com/drewthor/wolves_reddit_bot/util"
	log "github.com/sirupsen/logrus"

	"github.com/go-chi/chi/v5"
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
	r.Post("/updateGame", h.UpdateGame)

	return r
}

func (h *handler) List(w http.ResponseWriter, r *http.Request) {
	gameDate := r.URL.Query().Get("game-date")

	if gameDate == "" {
		util.WriteJSON(http.StatusBadRequest, "invalid request: missing game_date", w)
		return
	}

	games, err := h.GameService.List(r.Context(), gameDate)
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

	games, err := h.GameService.UpdateSeasonGames(r.Context(), seasonStartYear)
	if err != nil {
		log.WithError(err).Errorf("could not update games for season-start-year: %d", seasonStartYear)
		util.WriteJSON(http.StatusInternalServerError, err, w)
		return
	}

	util.WriteJSON(http.StatusOK, games, w)
}

func (h *handler) UpdateGame(w http.ResponseWriter, r *http.Request) {
	gameID := r.URL.Query().Get("game-id")
	gameDate := r.URL.Query().Get("game-date")
	seasonStartYearStr := r.URL.Query().Get("season-start-year")

	if gameID == "" {
		util.WriteJSON(http.StatusBadRequest, "invalid required game-id", w)
		return
	}

	if gameDate == "" {
		util.WriteJSON(http.StatusBadRequest, "invalid required game-date", w)
		return
	}

	seasonStartYear, err := strconv.Atoi(seasonStartYearStr)
	if err != nil {
		util.WriteJSON(http.StatusBadRequest, "invalid required season-start-year", w)
		return
	}

	games, err := h.GameService.UpdateGame(r.Context(), gameID, gameDate, seasonStartYear)
	if err != nil {
		log.WithError(err).Errorf("could not update game")
		util.WriteJSON(http.StatusInternalServerError, err, w)
		return
	}

	util.WriteJSON(http.StatusOK, games, w)
}
