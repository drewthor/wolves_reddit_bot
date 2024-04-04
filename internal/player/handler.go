package player

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/drewthor/wolves_reddit_bot/util"
	"go.opentelemetry.io/otel"

	"github.com/go-chi/chi/v5"
)

type Handler interface {
	Routes() chi.Router
	List(w http.ResponseWriter, r *http.Request)
	Get(w http.ResponseWriter, r *http.Request)
	UpdatePlayers(w http.ResponseWriter, r *http.Request)
}

func NewHandler(logger *slog.Logger, playerService Service) Handler {
	return &handler{logger: logger, playerService: playerService}
}

type handler struct {
	logger        *slog.Logger
	playerService Service
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

	players, err := h.playerService.ListPlayers(ctx)

	if err != nil {
		h.logger.ErrorContext(ctx, "failed to list players", slog.Any("error", err))
		util.WriteJSON(http.StatusInternalServerError, err, w)
		return
	}

	util.WriteJSON(http.StatusOK, players, w)
}

func (h *handler) Get(w http.ResponseWriter, r *http.Request) {
	ctx, span := otel.Tracer("player").Start(r.Context(), "player.handler.Get")
	defer span.End()

	playerID := chi.URLParam(r, "id")
	logger := h.logger.With(slog.String("player_id", playerID))

	player, err := h.playerService.Get(ctx, playerID)

	if err != nil {
		logger.ErrorContext(ctx, "failed to get player", slog.Any("error", err))
		util.WriteJSON(http.StatusInternalServerError, err, w)
		return
	}

	util.WriteJSON(http.StatusOK, player, w)
}

func (h *handler) UpdatePlayers(w http.ResponseWriter, r *http.Request) {
	ctx, span := otel.Tracer("player").Start(r.Context(), "player.handler.UpdatePlayers")
	defer span.End()

	seasonStartYear, err := strconv.Atoi(r.URL.Query().Get("season-start-year"))
	if err != nil {
		util.WriteJSON(http.StatusBadRequest, "invalid required season-start-year", w)
		return
	}

	players, err := h.playerService.UpdatePlayers(ctx, seasonStartYear)

	if err != nil {
		h.logger.ErrorContext(ctx, "failed to update players", slog.Any("error", err))
		util.WriteJSON(http.StatusInternalServerError, err, w)
		return
	}

	util.WriteJSON(http.StatusOK, players, w)
}
