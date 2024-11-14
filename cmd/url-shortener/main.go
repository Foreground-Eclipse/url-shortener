package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/foreground-eclipse/url-shortener/internal/config"
	"github.com/foreground-eclipse/url-shortener/internal/lib/logger/sl"
	"github.com/foreground-eclipse/url-shortener/internal/storage/postgres"
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

	// TODO: init router: chi <3, chi render

	// TODO: run server:
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
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
