package api

import (
	"io"
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

func (a *API) GetShardStatus(w http.ResponseWriter, r *http.Request) {
	req, err := http.NewRequest(http.MethodGet, a.Config.ShardsEndpoint, nil)
	if err != nil {
		http.Error(w, "error while creating request", 500)
		a.Logger.Error("error while creating request", slog.Any("error", err))
		return
	}
	resp, err := a.httpClient.Do(req)
	if err != nil {
		http.Error(w, "error while getting shard status", 500)
		a.Logger.Error("error while getting shard status", slog.Any("error", err))
		return
	} else if resp.StatusCode != 200 {
		http.Error(w, "error while getting shard status", 500)
		a.Logger.Error("non 200 when getting shard status")
		return
	}

	_, err = io.Copy(w, resp.Body)
	if err != nil {
		http.Error(w, "error while copying response", 500)
		a.Logger.Error("error while copying response", slog.Any("error", err))
		return
	}
	err = resp.Body.Close()
	if err != nil {
		a.Logger.Error("error while closing body", slog.Any("error", err))
	}
}
