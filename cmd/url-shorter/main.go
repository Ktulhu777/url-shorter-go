package main

import (
	"log/slog"
	"net/http"
	"os"
	"url-shorter/internal/config"
	"url-shorter/internal/http-server/handlers/auth/register"
	"url-shorter/internal/http-server/handlers/delete"
	"url-shorter/internal/http-server/handlers/redirect"
	"url-shorter/internal/http-server/handlers/url/save"
	mwLogger "url-shorter/internal/http-server/middleware/logger"
	myMiddleware "url-shorter/internal/http-server/middleware/authentication"
	"url-shorter/internal/lib/logger/sl"
	"url-shorter/internal/storage/sqlite"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	cfg := config.MustLoad()

	log := setupLogger(cfg.Env)

	log.Info("starting url shorter", slog.String("env", cfg.Env))
	log.Debug("debug message are enabled")

	storage, err := sqlite.NewStorage(cfg.StoragePath)
	if err != nil {
		log.Error("failed to init storage", sl.Err(err))
		os.Exit(1)
	}

	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(mwLogger.New(log))
	router.Use(middleware.RealIP)
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)

	router.Get("/{alias}", redirect.New(log, storage))

	authMiddleware := myMiddleware.BasicAuthMiddleware(log, storage)
	router.Route("/url", func(r chi.Router) {
		r.Use(authMiddleware)
		r.Post("/", save.New(log, storage))
		r.Delete("/{id}", delete.New(log, storage))
	})

	router.Post("/register", register.New(log, storage))

	log.Info("starting server", slog.String("address", cfg.Address))

	srv := &http.Server{
		Addr: cfg.Address,
		Handler: router,
		ReadTimeout: cfg.HTTPServer.Timeout,
		WriteTimeout: cfg.HTTPServer.Timeout,
		IdleTimeout: cfg.HTTPServer.Idle_timeout,
	}

	if err := srv.ListenAndServe(); err != nil {
		log.Error("failed to start server")
	}

	log.Error("server stopped")

}


func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envDev:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}

	return log
}
