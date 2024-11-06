package main

import (
	"fmt"

	"github.com/foreground-eclipse/url-shortener/internal/config"
)

func main() {
	cfg := config.MustLoad()

	fmt.Println(cfg)

	// TODO: init config: cleanenv/viper idk

	// TODO: init logger: slog

	// TODO: init storage: postgres

	// TODO: init router: chi <3, chi render

	// TODO: run server:
}
