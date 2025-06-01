package api

import (
	"net/http"
	"pluralkit/status/util"
	"time"

	"github.com/go-chi/render"
)

type wrapper struct {
	util.Status
	Timestamp time.Time `json:"timestamp"`
}

func (a *API) GetStatus(w http.ResponseWriter, r *http.Request) {
	status, err := a.Database.GetStatus(r.Context())
	if err != nil {
		http.Error(w, "error while getting status", 500)
		return
	}

	data := wrapper{
		status,
		time.Now(),
	}

	if err := render.Render(w, r, &data); err != nil {
		http.Error(w, "error while rendering json", 500)
		return
	}
}
