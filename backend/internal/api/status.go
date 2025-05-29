package api

import (
	"net/http"
	"time"

	"github.com/go-chi/render"
)

func (a *API) GetStatus(w http.ResponseWriter, r *http.Request) {
	incidents := make([]string, 0, 1)
	incidents = append(incidents, "testing")
	incidents = append(incidents, "testing2")
	tmp_stat := Status{
		Status:          StatusMajorOutage,
		Impact:          ImpactMajor,
		ActiveIncidents: incidents,
		Timestamp:       time.Now(),
	}
	if err := render.Render(w, r, &tmp_stat); err != nil {
		// TODO: handle render errors
		return
	}
}
