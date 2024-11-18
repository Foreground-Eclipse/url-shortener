package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/foreground-eclipse/url-shortener/internal/config"
	mwLogger "github.com/foreground-eclipse/url-shortener/internal/http-server/middleware/logger"
	"github.com/foreground-eclipse/url-shortener/internal/lib/logger/handlers/slogpretty"
	"github.com/foreground-eclipse/url-shortener/internal/lib/logger/sl"
	"github.com/foreground-eclipse/url-shortener/internal/storage/postgres"
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

	fmt.Println(cfg)

	log := setupLogger(cfg.Env)

	log.Info("starting url shortener", slog.String("env", cfg.Env))
	log.Debug("debug messages are on")

	storage, err := postgres.New()
	if err != nil {
		log.Error("failed to init storage", sl.Err(err))
		os.Exit(1)
	}

	if err := storage.Init(); err != nil {
		log.Error("failed to init url table ", sl.Err(err))
		os.Exit(1)
	}

	router := chi.NewRouter()
	// middleware

	router.Use(middleware.RequestID)
	router.Use(mwLogger.New(log))
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)
	// router.Use(middleware.RealIP) // unsure

	// TODO: run server:
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = setupPrettySlog()
	case envDev:
		{
			log = slog.New(
				slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
			)
		}
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}
	return log

}

func setupPrettySlog() *slog.Logger {
	opts := slogpretty.PrettyHandlerOptions{
		SlogOpts: &slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
	}

	handler := opts.NewPrettyHandler(os.Stdout)

	return slog.New(handler)
}
