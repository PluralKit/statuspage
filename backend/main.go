package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"pluralkit/status/api"
	"pluralkit/status/db"
	"pluralkit/status/util"
	"pluralkit/status/webhook"
	"syscall"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"

	_ "github.com/mattn/go-sqlite3"
)

func resetStatus(database *db.DB) {
	ctx := context.Background()
	status := util.Status{
		OverallStatus:   util.StatusOperational,
		ActiveIncidents: make([]string, 0),
	}

	incidents, err := database.GetActiveIncidents(ctx)
	if err != nil {
		slog.Error("error while resetting status!", slog.Any("error", err))
		return
	}

	highestImpact := util.ImpactNone
	for key, val := range incidents.Incidents {
		status.ActiveIncidents = append(status.ActiveIncidents, key)

		if val.Impact.IsGreater(highestImpact) {
			highestImpact = val.Impact
		}
	}

	switch highestImpact {
	case util.ImpactMajor:
		status.OverallStatus = util.StatusMajorOutage
	case util.ImpactMinor:
		status.OverallStatus = util.StatusDegraded
	default:
		status.OverallStatus = util.StatusOperational
	}

	err = database.SaveStatus(ctx, status)
	if err != nil {
		slog.Error("error while saving status to db", slog.Any("error", err))
		return
	}
}

func main() {
	var cfg util.Config
	err := env.Parse(&cfg)
	if err != nil {
		slog.Error("error while loading envs!", slog.Any("error", err))
		os.Exit(1)
	}

	var logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.Level(cfg.LogLevel),
	}))

	//setup our signal handler
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	//setup event channel
	eventChannel := make(chan util.Event)

	//setup discord notif webhook
	doNotifs := cfg.NotificationWebhook != ""
	webhook := webhook.NewDiscordWebhook(cfg)

	logger.Info("setting up database")
	db := db.NewDB(cfg, logger, eventChannel)
	if db == nil {
		os.Exit(1)
	}

	resetStatus(db)

	logger.Info("starting http api on ", slog.String("address", cfg.BindAddr))
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(render.SetContentType(render.ContentTypeJSON))
	r.Use(middleware.Timeout(30 * time.Second))

	apiInstance := api.NewAPI(cfg, logger, db)
	apiInstance.SetupRoutes(r)

	if cfg.RunDev {
		logger.Warn("serving /srv directory, this is intended for development use only!")
		fs := http.FileServer(http.Dir("./srv"))
		r.Handle("/*", fs)
	}

	go func() {
		err := http.ListenAndServe(cfg.BindAddr, r)
		if err != nil {
			logger.Error("error while running http router!", slog.Any("error", err))
			os.Exit(1)
		}
	}()

	go func() {
		//this is really just a notif handler for now, if other status checks get added, probably check them here
		ctx := context.Background()
		for {
			select {
			case <-quit:
				return
			case event := <-eventChannel:
				resetStatus(db)
				if err != nil {
					logger.Error("error while getting status!", slog.Any("error", err))
					continue
				}
				if !doNotifs {
					continue
				}
				switch event.Type {
				case util.EventCreateIncident:
					incident, ok := (event.Modified).(util.Incident)
					if !ok {
						continue
					}

					msgID, err := webhook.SendIncident(incident)
					if err != nil {
						logger.Error("error while sending incident notif!", slog.Any("error", err))
						continue
					}
					err = db.SaveMessageID(ctx, util.WebhookMessage{
						ID:        incident.ID,
						MessageID: msgID,
						Type:      "incident",
					})
					if err != nil {
						logger.Error("error while saving webhook message id", slog.Any("error", err))
					}
				case util.EventCreateUpdate:
					update, ok := (event.Modified).(util.IncidentUpdate)
					if !ok {
						continue
					}
					incident, err := db.GetIncident(ctx, update.IncidentID)
					if err != nil {
						logger.Error("error while getting incident for update notif!", slog.Any("error", err))
						continue
					}

					msgID, err := webhook.SendUpdate(incident, update)
					if err != nil {
						logger.Error("error while sending update notif!", slog.Any("error", err))
						continue
					}
					err = db.SaveMessageID(ctx, util.WebhookMessage{
						ID:        update.ID,
						MessageID: msgID,
						Type:      "update",
					})
					if err != nil {
						logger.Error("error while saving webhook message id", slog.Any("error", err))
					}
				case util.EventEditIncident:
					incident, ok := (event.Modified).(util.Incident)
					if !ok {
						continue
					}

					msgID, err := db.GetMessageID(ctx, incident.ID, "incident")
					if err != nil {
						logger.Error("error while getting msg id!", slog.Any("error", err))
						continue
					}

					err = webhook.EditIncident(msgID, incident)
					if err != nil {
						logger.Error("error while editing incident notif!", slog.Any("error", err))
						continue
					}
				case util.EventEditUpdate:
					update, ok := (event.Modified).(util.IncidentUpdate)
					if !ok {
						continue
					}
					incident, err := db.GetIncident(ctx, update.IncidentID)
					if err != nil {
						logger.Error("error while getting incident for update notif!", slog.Any("error", err))
						continue
					}

					msgID, err := db.GetMessageID(ctx, update.ID, "update")
					if err != nil {
						logger.Error("error while getting msg id!", slog.Any("error", err))
						continue
					}

					err = webhook.EditUpdate(msgID, incident, update)
					if err != nil {
						logger.Error("error while editing update notif!", slog.Any("error", err))
						continue
					}
				}
			}
		}
	}()

	//wait until sigint/sigterm and safely shutdown
	sig := <-quit
	logger.Info("shutting down", slog.String("signal", sig.String()))
	err = db.CloseDB()
	logger.Error("error while closing db", slog.Any("error", err))
}
