package main

import (
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"pluralkit/status/internal/api"
	"pluralkit/status/internal/db"
	"pluralkit/status/internal/util"
	"syscall"

	"github.com/caarlos0/env/v11"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/render"

	_ "github.com/mattn/go-sqlite3"
)

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

	logger.Info("setting up database")
	db := db.NewDB(cfg, logger)
	if db == nil {
		os.Exit(1)
	}

	logger.Info("starting http api on ", slog.String("address", cfg.BindAddr))
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(render.SetContentType(render.ContentTypeJSON))
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"https://*", "http://*"},
	})) //tmp for dev
	apiInstance := api.NewAPI(cfg, logger, db)
	apiInstance.SetupRoutes(r)

	go func() {
		err := http.ListenAndServe(cfg.BindAddr, r)
		if err != nil {
			logger.Error("error while running http router!", slog.Any("error", err))
		}
	}()

	//wait until sigint/sigterm and safely shutdown
	sig := <-quit
	logger.Info("shutting down", slog.String("signal", sig.String()))
	db.CloseDB()
}
