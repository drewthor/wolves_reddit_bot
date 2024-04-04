package game

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/drewthor/wolves_reddit_bot/internal/boxscore"
	"github.com/drewthor/wolves_reddit_bot/util"
	"go.opentelemetry.io/otel"

	"github.com/go-chi/chi/v5"
)

type Handler interface {
	Routes() chi.Router
	List(w http.ResponseWriter, r *http.Request)
	UpdateGames(w http.ResponseWriter, r *http.Request)
}

func NewHandler(logger *slog.Logger, gameService Service) Handler {
	return &handler{logger: logger, gameService: gameService}
}

type handler struct {
	logger      *slog.Logger
	gameService Service
}

func (h *handler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", h.List)

	r.Route("/{gameID}", func(r chi.Router) {
		r.Mount("/boxscore", boxscore.NewHandler(h.logger, boxscore.NewService()).Routes())
	})

	r.Post("/update", h.UpdateGames)
	r.Post("/updateGame", h.UpdateGame)

	return r
}

func (h *handler) List(w http.ResponseWriter, r *http.Request) {
	ctx, span := otel.Tracer("game").Start(r.Context(), "game.handler.List")
	defer span.End()

	//gameDate := r.URL.Query().Get("game-date")
	//if gameDate == "" {
	//	util.WriteJSON(http.StatusBadRequest, "invalid request: missing game_date", w)
	//	return
	//}

	//logger := h.logger.With(slog.String("game_date", gameDate))

	games, err := h.gameService.List(ctx)
	if err != nil {
		h.logger.ErrorContext(ctx, "failed to get games", slog.Any("error", err))
		util.WriteJSON(http.StatusInternalServerError, err, w)
		return
	}

	util.WriteJSON(http.StatusOK, games, w)
}

func (h *handler) UpdateGames(w http.ResponseWriter, r *http.Request) {
	ctx, span := otel.Tracer("game").Start(r.Context(), "game.handler.UpdateGames")
	defer span.End()

	seasonStartYear, err := strconv.Atoi(r.URL.Query().Get("season-start-year"))
	if err != nil {
		util.WriteJSON(http.StatusBadRequest, "invalid required season-start-year", w)
		return
	}

	logger := h.logger.With(slog.Int("season_start_year", seasonStartYear))

	games, err := h.gameService.UpdateSeasonGames(ctx, logger, seasonStartYear)
	if err != nil {
		logger.ErrorContext(ctx, "could not update games", slog.Any("error", err))
		util.WriteJSON(http.StatusInternalServerError, err, w)
		return
	}

	util.WriteJSON(http.StatusOK, games, w)
}

func (h *handler) UpdateGame(w http.ResponseWriter, r *http.Request) {
	ctx, span := otel.Tracer("game").Start(r.Context(), "game.handler.UpdateGame")
	defer span.End()

	gameID := r.URL.Query().Get("game-id")
	if gameID == "" {
		util.WriteJSON(http.StatusBadRequest, "invalid required game-id", w)
		return
	}

	seasonStartYearStr := r.URL.Query().Get("season-start-year")
	seasonStartYear, err := strconv.Atoi(seasonStartYearStr)
	if err != nil {
		util.WriteJSON(http.StatusBadRequest, "invalid required season-start-year", w)
		return
	}

	logger := h.logger.With(slog.Int("season_start_year", seasonStartYear), slog.String("game_id", gameID))

	games, err := h.gameService.UpdateGame(ctx, logger, gameID, seasonStartYear)
	if err != nil {
		logger.ErrorContext(ctx, "could not update game", slog.Any("error", err))
		util.WriteJSON(http.StatusInternalServerError, err, w)
		return
	}

	util.WriteJSON(http.StatusOK, games, w)
}
