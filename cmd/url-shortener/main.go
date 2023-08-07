package main

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	_ "github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/go-chi/render"
	"golang.org/x/exp/slog"
	"net/http"
	"os"
	"url-shortener/internal/config"
	del "url-shortener/internal/http-server/handlers/url/delete"
	"url-shortener/internal/http-server/handlers/url/redirect"
	"url-shortener/internal/http-server/handlers/url/save"
	mwLog "url-shortener/internal/http-server/middleware/logger"
	"url-shortener/internal/storage/sqlite"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	cfg := config.MustLoad()
	fmt.Println(cfg)

	logger := setupLogger(cfg.Env)
	_ = logger

	db, err := sqlite.New(cfg.StoragePath)
	if err != nil {
		logger.Error("failed while initializing storage")
	}
	_ = db

	router := chi.NewRouter()

	router.Use(middleware.Logger)
	router.Use(middleware.RequestID)
	router.Use(middleware.URLFormat)
	router.Use(middleware.Recoverer)
	router.Use(mwLog.New(logger))

	router.Post("/", save.New(logger, db))
	router.Get("/{alias}", redirect.New(logger, db))
	router.Post("/del", del.New(logger, db))

	srv := &http.Server{
		Handler:      router,
		Addr:         cfg.Address,
		ReadTimeout:  cfg.Timeout,
		WriteTimeout: cfg.Timeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	if err := srv.ListenAndServe(); err != nil {
		logger.Error("failed to run a server")
	}

}

func setupLogger(env string) *slog.Logger {
	var logger *slog.Logger

	switch env {
	case envLocal:
		logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case envDev:
		logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case envProd:
		logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}

	return logger
}
