package franchise

import (
	"log/slog"
	"net/http"

	"github.com/drewthor/wolves_reddit_bot/util"
	"go.opentelemetry.io/otel"

	"github.com/go-chi/chi/v5"
)

type Handler interface {
	Routes() chi.Router
	List(w http.ResponseWriter, r *http.Request)
	Get(w http.ResponseWriter, r *http.Request)
	UpdateFranchises(w http.ResponseWriter, r *http.Request)
}

func NewHandler(logger *slog.Logger, franchiseService Service) Handler {
	return &handler{logger: logger, franchiseService: franchiseService}
}

type handler struct {
	logger           *slog.Logger
	franchiseService Service
}

func (h *handler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", h.List)

	r.Get("/{teamID}", h.Get)

	r.Post("/update", h.UpdateFranchises)

	return r
}

func (h *handler) List(w http.ResponseWriter, r *http.Request) {
	_, span := otel.Tracer("team").Start(r.Context(), "franchise.handler.List")
	defer span.End()
}

func (h *handler) Get(w http.ResponseWriter, r *http.Request) {
	_, span := otel.Tracer("team").Start(r.Context(), "franchise.handler.Get")
	defer span.End()
}

func (h *handler) UpdateFranchises(w http.ResponseWriter, r *http.Request) {
	ctx, span := otel.Tracer("team").Start(r.Context(), "franchise.handler.UpdateFranchises")
	defer span.End()

	franchises, err := h.franchiseService.UpdateFranchises(ctx, h.logger)
	if err != nil {
		h.logger.ErrorContext(ctx, "failed to update franchises", slog.Any("error", err))
		util.WriteJSON(http.StatusInternalServerError, err, w)
		return
	}

	util.WriteJSON(http.StatusOK, franchises, w)
}
