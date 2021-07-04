package controller

import (
	"net/http"

	"github.com/drewthor/wolves_reddit_bot/util"

	"github.com/drewthor/wolves_reddit_bot/service"
	"github.com/go-chi/chi"
)

type TeamController struct {
	TeamService *service.TeamService
}

func (tc TeamController) Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", tc.List)

	r.Get("/{teamID}", tc.Get)

	r.Post("/update", tc.UpdateTeams)

	return r
}

func (tc TeamController) List(w http.ResponseWriter, r *http.Request) {
	teams, err := tc.TeamService.GetAll()

	if err != nil {
		util.WriteJSON(http.StatusInternalServerError, err, w)

	}

	util.WriteJSON(http.StatusOK, teams, w)

}

func (tc TeamController) Get(w http.ResponseWriter, r *http.Request) {
	teamID := chi.URLParam(r, "teamID")

	team, err := tc.TeamService.Get(teamID)

	if err != nil {
		util.WriteJSON(http.StatusInternalServerError, err, w)

	}

	util.WriteJSON(http.StatusOK, team, w)
}

func (tc TeamController) UpdateTeams(w http.ResponseWriter, r *http.Request) {
	teams, err := tc.TeamService.UpdateTeams()

	if err != nil {
		util.WriteJSON(http.StatusInternalServerError, err, w)
	}

	util.WriteJSON(http.StatusOK, teams, w)
}
