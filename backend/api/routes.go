package api

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"pluralkit/status/db"
	"pluralkit/status/util"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
)

type UnixTime struct {
	time.Time
}

func (u *UnixTime) UnmarshalJSON(b []byte) error {
	var timestamp int64
	err := json.Unmarshal(b, &timestamp)
	if err != nil {
		return err
	}
	u.Time = time.Unix(timestamp, 0)
	return nil
}
func (u UnixTime) MarshalJSON() ([]byte, error) {
	if u.Time.IsZero() { //nolint:all
		return []byte("0"), nil
	}
	return []byte(fmt.Sprintf("%d", u.Time.Unix())), nil //nolint:all
}

type Shard struct {
	ShardID            int      `json:"shard_id"`
	ClusterID          int      `json:"cluster_id"`
	Up                 bool     `json:"up"`
	DisconnectionCount int      `json:"disconnection_count"`
	Latency            int      `json:"latency"`
	LastHeartbeat      UnixTime `json:"last_heartbeat"`
	LastConnection     UnixTime `json:"last_connection"`
	LastReconnect      UnixTime `json:"last_reconnect"`
}

type Cluster struct {
	AvgLatency int     `json:"avg_latency"`
	ShardsUp   int     `json:"shards_up"`
	Shards     []Shard `json:"-"`
	Up         bool    `json:"up"`
}

type ClustersInfo struct {
	AvgLatency     int        `json:"avg_latency"`
	MaxConcurrency int        `json:"max_concurrency"`
	NumShards      int        `json:"num_shards"`
	ShardsUp       int        `json:"shards_up"`
	Clusters       []*Cluster `json:"clusters"`
}

type API struct {
	Config     util.Config
	Logger     *slog.Logger
	Database   *db.DB
	Sessions   *scs.SessionManager
	httpClient http.Client

	clustersCache  ClustersInfo
	cacheTimestamp time.Time
	cacheMutex     sync.RWMutex
}

func NewAPI(config util.Config, logger *slog.Logger, database *db.DB) *API {
	moduleLogger := logger.With(slog.String("module", "API"))

	sessionManager := scs.New()
	sessionManager.Lifetime = 24 * time.Hour
	sessionManager.Cookie.Name = "session_id"
	sessionManager.Cookie.HttpOnly = true
	sessionManager.Cookie.Secure = !config.RunDev
	sessionManager.Cookie.SameSite = http.SameSiteLaxMode

	return &API{
		Config:     config,
		Logger:     moduleLogger,
		Database:   database,
		Sessions:   sessionManager,
		httpClient: http.Client{Timeout: 10 * time.Second},
		clustersCache: ClustersInfo{
			Clusters:       make([]*Cluster, 0),
			MaxConcurrency: config.MaxConcurrency,
		},
	}
}

func (a *API) Auth() func(http.Handler) http.Handler {
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

			if split[1] == a.Config.AuthToken {
				next.ServeHTTP(w, r)
				return
			}

			val := a.Sessions.Get(r.Context(), "user_session")
			user, ok := val.(DiscordUser)
			if !ok {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			if user.ID == "" || !slices.Contains(a.Config.AuthorizedUsers, user.ID) {
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func (a *API) Me(w http.ResponseWriter, r *http.Request) {
	val := a.Sessions.Get(r.Context(), "user_session")
	user, ok := val.(DiscordUser)

	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var avatarURL string
	if user.Avatar != "" {
		avatarURL = "https://cdn.discordapp.com/avatars/" + user.ID + "/" + user.Avatar + ".png"
	} else {
		avatarURL = "https://cdn.discordapp.com/embed/avatars/0.png"
	}

	response := map[string]string{
		"id":       user.ID,
		"username": user.Username,
		"avatar":   avatarURL,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (a *API) SetupRoutes(router *chi.Mux) {
	a.oauthInit()
	router.Route("/api/v1", func(r chi.Router) {

		r.Get("/status", a.GetStatus)

		r.Route("/clusters", func(r chi.Router) {
			r.Get("/", a.GetClusters)
			r.Get("/{clusterID}", a.GetShards)
		})

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

		r.Route("/auth/discord", func(r chi.Router) {
			r.Get("/login", a.oauthDiscordLogin)
			r.Get("/callback", a.oauthDiscordCallback)
			r.Get("/logout", a.logout)
		})
		r.Get("/me", a.Me)

		r.Route("/admin", func(r chi.Router) {
			r.Use(a.Auth())

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
