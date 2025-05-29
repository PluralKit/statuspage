package api

import (
	"net/http"
	"pluralkit/status/internal/util"
	"time"

	"github.com/go-chi/render"
)

func (a *API) GetIncidents(w http.ResponseWriter, r *http.Request) {
	updates := make([]*util.IncidentUpdate, 0)
	updates = append(updates, &util.IncidentUpdate{
		ID:        "testing-update",
		Text:      "uhhhh there's a fox in the servers biting on wires???",
		Timestamp: time.Now().Add(-15 * time.Minute),
	})
	updates = append(updates, &util.IncidentUpdate{
		ID:        "testing-update2",
		Text:      "we have lured the fox out with a sandwich! working on repairing the wires now",
		Timestamp: time.Now().Add(-5 * time.Minute),
	})
	incident := util.Incident{
		ID:                  "testing",
		Timestamp:           time.Now().Add(-35 * time.Minute),
		Status:              util.StatusIdentified,
		Impact:              util.ImpactMinor,
		Updates:             updates,
		Name:                "Avatars not loading",
		Description:         "Some avatars are not loading.",
		LastUpdate:          time.Now().Add(-5 * time.Minute),
		ResolutionTimestamp: time.Time{},
	}
	incidentB := util.Incident{
		ID:                  "testing2",
		Timestamp:           time.Now().Add(-5 * time.Minute),
		Status:              util.StatusInvestigating,
		Impact:              util.ImpactMajor,
		Name:                "Myriad fell asleep :(",
		Description:         "PluralKit currently isn't working because Myriad is taking a nap.",
		LastUpdate:          time.Now().Add(-5 * time.Minute),
		ResolutionTimestamp: time.Time{},
	}

	incidents := make(map[string]util.Incident)
	incidents["testing"] = incident
	incidents["testing2"] = incidentB

	tmp_incidentlist := util.IncidentList{
		Timestamp: time.Now(),
		Incidents: incidents,
	}

	if err := render.Render(w, r, &tmp_incidentlist); err != nil {
		// TODO: handle render errors
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
