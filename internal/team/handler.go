package team

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
	UpdateTeams(w http.ResponseWriter, r *http.Request)
}

func NewHandler(logger *slog.Logger, teamService Service) Handler {
	return &handler{logger: logger, teamService: teamService}
}

type handler struct {
	logger      *slog.Logger
	teamService Service
}

func (h *handler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", h.List)

	r.Get("/{teamID}", h.Get)

	r.Post("/update", h.UpdateTeams)

	return r
}

func (h *handler) List(w http.ResponseWriter, r *http.Request) {
	ctx, span := otel.Tracer("team").Start(r.Context(), "team.handler.List")
	defer span.End()

	teams, err := h.teamService.ListTeams(ctx)

	if err != nil {
		h.logger.Error("failed to list teams", slog.Any("error", err))
		util.WriteJSON(http.StatusInternalServerError, err, w)
		return
	}

	util.WriteJSON(http.StatusOK, teams, w)
}

func (h *handler) Get(w http.ResponseWriter, r *http.Request) {
	ctx, span := otel.Tracer("team").Start(r.Context(), "team.handler.Get")
	defer span.End()

	teamID := chi.URLParam(r, "teamID")
	logger := h.logger.With(slog.String("team_id", teamID))

	team, err := h.teamService.Get(ctx, teamID)

	if err != nil {
		logger.Error("failed to get team", slog.Any("error", err))
		util.WriteJSON(http.StatusInternalServerError, err, w)
		return
	}

	util.WriteJSON(http.StatusOK, team, w)
}

func (h *handler) UpdateTeams(w http.ResponseWriter, r *http.Request) {
	ctx, span := otel.Tracer("team").Start(r.Context(), "team.handler.UpdateTeams")
	defer span.End()

	seasonStartYear, err := strconv.Atoi(r.URL.Query().Get("season-start-year"))
	if err != nil {
		util.WriteJSON(http.StatusBadRequest, "invalid required season-start-year", w)
		return
	}

	logger := h.logger.With(slog.Int("season_start_year", seasonStartYear))

	teams, err := h.teamService.UpdateTeams(ctx, seasonStartYear)

	if err != nil {
		logger.Error("failed to update teams", slog.Any("error", err))
		util.WriteJSON(http.StatusInternalServerError, err, w)
		return
	}

	util.WriteJSON(http.StatusOK, teams, w)
}
