package api

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"pluralkit/status/util"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

func (a *API) GetIncidents(w http.ResponseWriter, r *http.Request) {
	/*
		TODO: implement this

		it should fetch the <insert arbitrary number> most recent incidents
		allow grabbing more events with a 'before' argument? similar to the PK API?
	*/
}

func (a *API) GetActiveIncidents(w http.ResponseWriter, r *http.Request) {
	list, err := a.Database.GetActiveIncidents(r.Context())
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
	id := []string{chi.URLParam(r, "incidentID")}
	list, err := a.Database.GetIncidents(r.Context(), id)
	if err != nil {
		//TODO: handle not found
		http.Error(w, "error while retrieving incident", 500)
		a.Logger.Error("error while fufilling get incident request", slog.Any("error", err))
		return
	}

	incident := list.Incidents[id[0]]

	if err := render.Render(w, r, &incident); err != nil {
		http.Error(w, "error while rendering json", 500)
		a.Logger.Error("error while rendering json for get incident request", slog.Any("error", err))
		return
	}
}

func (a *API) CreateIncident(w http.ResponseWriter, r *http.Request) {
	data, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "error while getting body data", 500)
		a.Logger.Error("error while getting body data", slog.Any("error", err))
		return
	}

	var incident util.Incident

	err = json.Unmarshal(data, &incident)
	if err != nil {
		http.Error(w, "error while parsing incident data", 400)
		a.Logger.Error("error while parsing incident data", slog.Any("error", err))
	}

	id, err := a.Database.CreateIncident(r.Context(), incident)
	if err != nil {
		http.Error(w, "error while creating incident", 500)
		a.Logger.Error("error while creating incident", slog.Any("error", err))
	}
	w.Write([]byte(id))
}

func (a *API) EditIncident(w http.ResponseWriter, r *http.Request) {
	data, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "error while getting body data", 500)
		a.Logger.Error("error while getting body data", slog.Any("error", err))
		return
	}

	var incidentPatch util.IncidentPatch
	id := chi.URLParam(r, "incidentID")

	err = json.Unmarshal(data, &incidentPatch)
	if err != nil {
		http.Error(w, "error while parsing incident data", 400)
		a.Logger.Error("error while parsing incident data", slog.Any("error", err))
	}

	err = a.Database.EditIncident(r.Context(), id, incidentPatch)
	if err != nil {
		//TODO: handle not found
		http.Error(w, "error while editing incident", 500)
		a.Logger.Error("error while editing incident", slog.Any("error", err))
	}
}

func (a *API) DeleteIncident(w http.ResponseWriter, r *http.Request) {
	var incident util.Incident
	incident.ID = chi.URLParam(r, "incidentID")

	err := a.Database.DeleteIncident(r.Context(), incident)
	if err != nil {
		//TODO: handle not found
		http.Error(w, "error while deleting incident", 500)
		a.Logger.Error("error while deleting incident", slog.Any("error", err))
	}
}

func (a *API) AddUpdate(w http.ResponseWriter, r *http.Request) {
	data, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "error while getting body data", 500)
		a.Logger.Error("error while getting body data", slog.Any("error", err))
		return
	}

	var update util.IncidentUpdate
	update.IncidentID = chi.URLParam(r, "incidentID")

	err = json.Unmarshal(data, &update)
	if err != nil {
		http.Error(w, "error while parsing update data", 400)
		a.Logger.Error("error while parsing update data", slog.Any("error", err))
	}

	id, err := a.Database.CreateUpdate(r.Context(), update)
	if err != nil {
		http.Error(w, "error while creating update", 500)
		a.Logger.Error("error while creating update", slog.Any("error", err))
	}
	w.Write([]byte(id))
}

func (a *API) EditUpdate(w http.ResponseWriter, r *http.Request) {
	data, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "error while getting body data", 500)
		a.Logger.Error("error while getting body data", slog.Any("error", err))
		return
	}

	var update util.UpdatePatch
	id := chi.URLParam(r, "updateID")

	err = json.Unmarshal(data, &update)
	if err != nil {
		http.Error(w, "error while parsing update data", 400)
		a.Logger.Error("error while parsing update data", slog.Any("error", err))
	}

	err = a.Database.EditUpdate(r.Context(), id, update)
	if err != nil {
		//TODO: handle not found
		http.Error(w, "error while editing update", 500)
		a.Logger.Error("error while editing update", slog.Any("error", err))
	}
}

func (a *API) GetUpdate(w http.ResponseWriter, r *http.Request) {
	update, err := a.Database.GetUpdate(r.Context(), chi.URLParam(r, "updateID"))
	if err != nil {
		//TODO: handle not found
		http.Error(w, "error while getting update", 500)
		a.Logger.Error("error while getting update", slog.Any("error", err))
	}

	if err := render.Render(w, r, &update); err != nil {
		http.Error(w, "error while rendering json", 500)
		a.Logger.Error("error while rendering json for get incident request", slog.Any("error", err))
		return
	}
}

func (a *API) DeleteUpdate(w http.ResponseWriter, r *http.Request) {
	var update util.IncidentUpdate
	update.IncidentID = chi.URLParam(r, "incidentID")
	update.ID = chi.URLParam(r, "updateID")

	err := a.Database.DeleteUpdate(r.Context(), update)
	if err != nil {
		//TODO: handle not found
		http.Error(w, "error while deleting update", 500)
		a.Logger.Error("error while deleting update", slog.Any("error", err))
	}
}
