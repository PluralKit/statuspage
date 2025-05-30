package api

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/go-chi/render"
)

func (a *API) GetIncidents(w http.ResponseWriter, r *http.Request) {
	// if err := render.Render(w, r, &incidentlist); err != nil {
	// 	http.Error(w, "error while rendering json", 500)
	// 	return
	// }
}

func (a *API) GetActiveIncidents(w http.ResponseWriter, r *http.Request) {
	list, err := a.Database.GetIncidents(context.Background(), a.Status.ActiveIncidents)
	if err != nil {
		http.Error(w, "error while retrieving incidents", 500)
		a.Logger.Error("error while fufilling active incidents request", slog.Any("error", err))
		return
	}

	if err := render.Render(w, r, &list); err != nil {
		http.Error(w, "error while rendering json", 500)
		a.Logger.Error("error while rendering json for active incidents request", slog.Any("error", err))
		return
	}
}

func (a *API) GetIncident(w http.ResponseWriter, r *http.Request) {

}

func (a *API) CreateIncident(w http.ResponseWriter, r *http.Request) {

}

func (a *API) EditIncident(w http.ResponseWriter, r *http.Request) {

}

func (a *API) DeleteIncident(w http.ResponseWriter, r *http.Request) {

}

func (a *API) AddUpdate(w http.ResponseWriter, r *http.Request) {

}

func (a *API) EditUpdate(w http.ResponseWriter, r *http.Request) {

}

func (a *API) GetUpdate(w http.ResponseWriter, r *http.Request) {

}

func (a *API) DeleteUpdate(w http.ResponseWriter, r *http.Request) {

}
