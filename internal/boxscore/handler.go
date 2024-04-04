package boxscore

import (
	"log/slog"
	"net/http"

	"github.com/drewthor/wolves_reddit_bot/util"
	"github.com/go-chi/chi/v5"
	"go.opentelemetry.io/otel"
)

type Handler interface {
	Routes() chi.Router
	Get(w http.ResponseWriter, r *http.Request)
}

func NewHandler(logger *slog.Logger, boxscoreService Service) Handler {
	return &handler{logger: logger, boxscoreService: boxscoreService}
}

type handler struct {
	logger          *slog.Logger
	boxscoreService Service
}

func (h *handler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", h.Get)

	return r
}

func (h *handler) Get(w http.ResponseWriter, r *http.Request) {
	ctx, span := otel.Tracer("boxscore").Start(r.Context(), "boxscore.handler.Get")
	defer span.End()

	r.ParseForm()
	gameID := chi.URLParam(r, "gameID")
	if gameID == "" {
		util.WriteJSON(http.StatusBadRequest, "invalid request: missing game_id", w)
		return
	}

	gameDate := r.FormValue("game_date")
	if gameDate == "" {
		util.WriteJSON(http.StatusBadRequest, "invalid request: missing game_date", w)
		return
	}

	logger := h.logger.With(slog.String("game_id", gameID), slog.String("game_date", gameDate))

	boxscore, err := h.boxscoreService.Get(ctx, gameID, gameDate)
	if err != nil {
		logger.ErrorContext(ctx, "failed to get boxscore", slog.Any("error", err))
		util.WriteJSON(http.StatusInternalServerError, err, w)
		return
	}

	util.WriteJSON(http.StatusOK, boxscore, w)
}
