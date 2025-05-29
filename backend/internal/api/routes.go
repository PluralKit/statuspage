package api

import (
	"log/slog"
	"pluralkit/status/internal/util"

	"github.com/go-chi/chi/v5"
)

type API struct {
	Config util.Config
	Logger *slog.Logger
}

func NewAPI(config util.Config, logger *slog.Logger) *API {
	moduleLogger := logger.With(slog.String("module", "API"))
	return &API{
		Config: config,
		Logger: moduleLogger,
	}
}

func (a *API) SetupRoutes(router *chi.Mux) {
	router.Route("/api/v1", func(r chi.Router) {
		r.Get("/incidents", a.GetIncidents)
		r.Get("/status", a.GetStatus)
	})
}
