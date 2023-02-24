package player

import (
	"net/http"
	"strconv"

	"github.com/drewthor/wolves_reddit_bot/util"
	log "github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"

	"github.com/go-chi/chi/v5"
)

type Handler interface {
	Routes() chi.Router
	List(w http.ResponseWriter, r *http.Request)
	Get(w http.ResponseWriter, r *http.Request)
	UpdatePlayers(w http.ResponseWriter, r *http.Request)
}

func NewHandler(playerService Service) Handler {
	return &handler{PlayerService: playerService}
}

type handler struct {
	PlayerService Service
}

func (h handler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", h.List)

	r.Get("/{id}", h.Get)

	r.Post("/update", h.UpdatePlayers)

	return r
}

func (h *handler) List(w http.ResponseWriter, r *http.Request) {
	ctx, span := otel.Tracer("player").Start(r.Context(), "player.handler.List")
	defer span.End()

	players, err := h.PlayerService.ListPlayers(ctx)

	if err != nil {
		log.Error(err)
		util.WriteJSON(http.StatusInternalServerError, err, w)
		return
	}

	util.WriteJSON(http.StatusOK, players, w)
}

func (h *handler) Get(w http.ResponseWriter, r *http.Request) {
	playerID := chi.URLParam(r, "id")

	player, err := h.PlayerService.Get(r.Context(), playerID)

	if err != nil {
		log.Error(err)
		util.WriteJSON(http.StatusInternalServerError, err, w)
		return
	}

	util.WriteJSON(http.StatusOK, player, w)
}

func (h *handler) UpdatePlayers(w http.ResponseWriter, r *http.Request) {
	seasonStartYear, err := strconv.Atoi(r.URL.Query().Get("season-start-year"))
	if err != nil {
		util.WriteJSON(http.StatusBadRequest, "invalid required season-start-year", w)
		return
	}

	players, err := h.PlayerService.UpdatePlayers(r.Context(), seasonStartYear)

	if err != nil {
		log.Error(err)
		util.WriteJSON(http.StatusInternalServerError, err, w)
		return
	}

	util.WriteJSON(http.StatusOK, players, w)
}