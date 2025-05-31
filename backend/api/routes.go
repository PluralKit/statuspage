package api

import (
	"log/slog"
	"pluralkit/status/db"
	"pluralkit/status/util"

	"github.com/go-chi/chi/v5"
)

type API struct {
	Config   util.Config
	Logger   *slog.Logger
	Database *db.DB
}

func NewAPI(config util.Config, logger *slog.Logger, database *db.DB) *API {
	moduleLogger := logger.With(slog.String("module", "API"))
	return &API{
		Config:   config,
		Logger:   moduleLogger,
		Database: database,
	}
}

func (a *API) SetupRoutes(router *chi.Mux) {
	router.Route("/api/v1", func(r chi.Router) {

		r.Get("/status", a.GetStatus)

		r.Route("/incidents", func(r chi.Router) {
			r.Get("/", a.GetIncidents)
			r.Get("/active", a.GetActiveIncidents)
			r.Route("/{incidentID}", func(r chi.Router) {
				r.Get("/", a.GetIncident)
				r.Route("/update", func(r chi.Router) {
					r.Route("/{updateID}", func(r chi.Router) {
						r.Get("/", a.GetUpdate)
					})
				})
			})
		})

		r.Route("/admin", func(r chi.Router) {
			r.Route("/incidents", func(r chi.Router) {
				r.Post("/create", a.CreateIncident)
				r.Route("/{incidentID}", func(r chi.Router) {
					r.Patch("/", a.EditIncident)
					r.Delete("/", a.DeleteIncident)
					r.Route("/update", func(r chi.Router) {
						r.Post("/", a.AddUpdate)
						r.Route("/{updateID}", func(r chi.Router) {
							r.Patch("/", a.EditUpdate)
							r.Delete("/", a.DeleteUpdate)
						})
					})
				})
			})
		})

	})
}
