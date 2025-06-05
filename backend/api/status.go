package api

import (
	"log/slog"
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
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		a.Logger.Error("error while getting status", slog.Any("error", err))
		return
	}

	data := wrapper{
		status,
		time.Now(),
	}

	if err := render.Render(w, r, &data); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		a.Logger.Error("error while rendering json", slog.Any("error", err))
		return
	}
}
