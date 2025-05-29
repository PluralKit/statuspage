package api

import (
	"net/http"
	"time"

	"github.com/go-chi/render"
)

func (a *API) GetIncidents(w http.ResponseWriter, r *http.Request) {
	updates := make([]IncidentUpdate, 0)
	updates = append(updates, IncidentUpdate{
		ID:        "testing-update",
		Text:      "uhhhh there's a fox in the servers biting on wires???",
		Timestamp: time.Now().Add(-15 * time.Minute),
	})
	updates = append(updates, IncidentUpdate{
		ID:        "testing-update2",
		Text:      "we have lured the fox out with a sandwich! working on repairing the wires now",
		Timestamp: time.Now().Add(-5 * time.Minute),
	})
	incident := Incident{
		ID:                  "testing",
		Timestamp:           time.Now().Add(-35 * time.Minute),
		Status:              StatusIdentified,
		Impact:              ImpactMinor,
		Updates:             updates,
		Name:                "Avatars not loading",
		Description:         "Some avatars are not loading.",
		LastUpdate:          time.Now().Add(-5 * time.Minute),
		ResolutionTimestamp: time.Time{},
	}
	incidentB := Incident{
		ID:                  "testing2",
		Timestamp:           time.Now().Add(-5 * time.Minute),
		Status:              StatusInvestigating,
		Impact:              ImpactMajor,
		Name:                "Myriad fell asleep :(",
		Description:         "PluralKit currently isn't working because Myriad is taking a nap.",
		LastUpdate:          time.Now().Add(-5 * time.Minute),
		ResolutionTimestamp: time.Time{},
	}

	incidents := make(map[string]Incident)
	incidents["testing"] = incident
	incidents["testing2"] = incidentB

	tmp_incidentlist := IncidentList{
		Timestamp: time.Now(),
		Incidents: incidents,
	}

	if err := render.Render(w, r, &tmp_incidentlist); err != nil {
		// TODO: handle render errors
		return
	}
}
