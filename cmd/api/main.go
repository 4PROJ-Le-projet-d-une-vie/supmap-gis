package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"supmap-gis/internal/api"
	"supmap-gis/internal/config"
	"supmap-gis/internal/domain/services"
	"supmap-gis/internal/providers/nominatim"
	"syscall"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	conf, err := config.New()
	if err != nil {
		return err
	}

	jsonHandler := slog.NewJSONHandler(os.Stdout, nil)
	logger := slog.New(jsonHandler)

	nominatimURL := fmt.Sprintf("http://%s:%s", conf.NominatimHost, conf.NominatimPort)
	nominatimClient := nominatim.NewClient(nominatimURL)
	logger.Info("Nominatim client initialized", "url", nominatimURL)

	geocodingService := services.NewGeocodingService(nominatimClient)

	server := api.NewServer(conf, logger, geocodingService)
	if err := server.Start(ctx); err != nil {
		return err
	}

	return nil
}
