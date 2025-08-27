package api

import (
	"log/slog"
	"net/http"
	"pluralkit/status/db"
	"pluralkit/status/util"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
)

type cacheEntry struct {
	data      []byte
	timestamp time.Time
}

type API struct {
	Config     util.Config
	Logger     *slog.Logger
	Database   *db.DB
	httpClient http.Client

	cache      map[string]cacheEntry
	cacheMutex sync.RWMutex
}

func NewAPI(config util.Config, logger *slog.Logger, database *db.DB) *API {
	moduleLogger := logger.With(slog.String("module", "API"))
	return &API{
		Config:     config,
		Logger:     moduleLogger,
		Database:   database,
		httpClient: http.Client{},
		cache:      make(map[string]cacheEntry),
	}
}

// this isn't that secure, and it's not supposed to be.
// backend should be behind a reverse proxy with only local/certain IPs allowed
func BasicTokenAuth(token string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "token not provided", http.StatusUnauthorized)
				return
			}
			split := strings.Split(authHeader, " ")
			if len(split) != 2 || strings.ToLower(split[0]) != "bearer" {
				http.Error(w, "invalid header format", http.StatusUnauthorized)
				return
			}

			if split[1] == token {
				next.ServeHTTP(w, r)
			} else {
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			}
		})
	}
}

func (a *API) SetupRoutes(router *chi.Mux) {
	router.Route("/api/v1", func(r chi.Router) {

		r.Get("/status", a.GetStatus)
		r.Get("/shards", a.GetShardStatus)

		r.Route("/incidents", func(r chi.Router) {
			r.Get("/", a.GetIncidents)
			r.Get("/active", a.GetActiveIncidents)
			r.Route("/{incidentID}", func(r chi.Router) {
				r.Get("/", a.GetIncident)
			})
		})
		r.Route("/updates/{updateID}", func(r chi.Router) {
			r.Get("/", a.GetUpdate)
		})

		r.Route("/admin", func(r chi.Router) {
			if a.Config.AuthToken != "" {
				r.Use(BasicTokenAuth(a.Config.AuthToken))
			} else {
				a.Logger.Warn("auth token is not set! admin endpoint auth disabled!")
			}

			r.Route("/incidents", func(r chi.Router) {
				r.Post("/create", a.CreateIncident)
				r.Route("/{incidentID}", func(r chi.Router) {
					r.Patch("/", a.EditIncident)
					r.Delete("/", a.DeleteIncident)
					r.Post("/update", a.AddUpdate)
				})
			})
			r.Route("/updates/{updateID}", func(r chi.Router) {
				r.Patch("/", a.EditUpdate)
				r.Delete("/", a.DeleteUpdate)
			})
		})

	})
}
