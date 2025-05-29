package main

import (
	"log/slog"
	"net/http"
	"os"
	"pluralkit/status/internal/api"
	"pluralkit/status/internal/util"

	"github.com/caarlos0/env/v11"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/render"
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

	logger.Info("starting http api on ", slog.String("address", cfg.BindAddr))
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(render.SetContentType(render.ContentTypeJSON))
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"https://*", "http://*"},
	})) //tmp for dev
	apiInstance := api.NewAPI(cfg, logger)
	apiInstance.SetupRoutes(r)

	http.ListenAndServe(cfg.BindAddr, r)
}
