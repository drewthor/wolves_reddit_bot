package team

import (
	"net/http"
	"strconv"

	"github.com/drewthor/wolves_reddit_bot/util"
	log "github.com/sirupsen/logrus"

	"github.com/go-chi/chi/v5"
)

type Handler interface {
	Routes() chi.Router
	List(w http.ResponseWriter, r *http.Request)
	Get(w http.ResponseWriter, r *http.Request)
	UpdateTeams(w http.ResponseWriter, r *http.Request)
}

func NewHandler(teamService Service) Handler {
	return &handler{TeamService: teamService}
}

type handler struct {
	TeamService Service
}

func (h *handler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", h.List)

	r.Get("/{teamID}", h.Get)

	r.Post("/update", h.UpdateTeams)

	return r
}

func (h *handler) List(w http.ResponseWriter, r *http.Request) {
	teams, err := h.TeamService.ListPlayers(r.Context())

	if err != nil {
		log.Error(err)
		util.WriteJSON(http.StatusInternalServerError, err, w)
		return
	}

	util.WriteJSON(http.StatusOK, teams, w)
}

func (h *handler) Get(w http.ResponseWriter, r *http.Request) {
	teamID := chi.URLParam(r, "teamID")

	team, err := h.TeamService.Get(r.Context(), teamID)

	if err != nil {
		log.Error(err)
		util.WriteJSON(http.StatusInternalServerError, err, w)
		return
	}

	util.WriteJSON(http.StatusOK, team, w)
}

func (h *handler) UpdateTeams(w http.ResponseWriter, r *http.Request) {
	seasonStartYear, err := strconv.Atoi(r.URL.Query().Get("season-start-year"))
	if err != nil {
		util.WriteJSON(http.StatusBadRequest, "invalid required season-start-year", w)
		return
	}

	teams, err := h.TeamService.UpdateTeams(r.Context(), seasonStartYear)

	if err != nil {
		log.Error(err)
		util.WriteJSON(http.StatusInternalServerError, err, w)
		return
	}

	util.WriteJSON(http.StatusOK, teams, w)
}
