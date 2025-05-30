package api

import (
	"log/slog"
	"pluralkit/status/internal/db"
	"pluralkit/status/internal/util"

	"github.com/go-chi/chi/v5"
)

type API struct {
	Config   util.Config
	Logger   *slog.Logger
	Status   *util.Status
	Database *db.DB
}

func NewAPI(config util.Config, logger *slog.Logger, status *util.Status, database *db.DB) *API {
	moduleLogger := logger.With(slog.String("module", "API"))
	return &API{
		Config:   config,
		Logger:   moduleLogger,
		Status:   status,
		Database: database,
	}
}

func (a *API) SetupRoutes(router *chi.Mux) {
	router.Route("/api/v1", func(r chi.Router) {

		r.Route("/incidents", func(r chi.Router) {
			r.Get("/", a.GetIncidents)
			r.Get("/active", a.GetActiveIncidents)
			r.Post("/create", a.CreateIncident)

			r.Route("/{incidentID}", func(r chi.Router) {
				r.Get("/", a.GetIncident)
				r.Patch("/", a.EditIncident)
				r.Delete("/", a.DeleteIncident)

				r.Route("/update", func(r chi.Router) {
					r.Post("/", a.AddUpdate)
					r.Route("/{updateID}", func(r chi.Router) {
						r.Get("/", a.GetUpdate)
						r.Patch("/", a.EditUpdate)
						r.Delete("/", a.DeleteUpdate)
					})
				})
			})
		})

		r.Get("/status", a.GetStatus)
	})
}
