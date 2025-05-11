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
	"supmap-gis/internal/providers/nominatim"
	supmapIncidents "supmap-gis/internal/providers/supmap-incidents"
	"supmap-gis/internal/providers/valhalla"
	"supmap-gis/internal/services"
	"syscall"
)

// @title			SupMap GIS API
// @version		1.0
// @description	Cette API permet de réaliser des calculs et opérations géographiques (routing, geocoding, map-matching).
// @accept			json
// @schemes		http
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

	supmapIncidentsURL := fmt.Sprintf("http://%s:%s", conf.SupmapIncidentsHost, conf.SupmapIncidentsPort)
	supmapIncidentsClient := supmapIncidents.NewClient(supmapIncidentsURL)
	logger.Info("supmap-incidents client initialized", "url", supmapIncidentsURL)

	incidentsService := services.NewIncidentsService(supmapIncidentsClient)

	valhallaURL := fmt.Sprintf("http://%s:%s", conf.ValhallaHost, conf.ValhallaPort)
	valhallaClient := valhalla.NewClient(valhallaURL)
	logger.Info("Valhalla client initialized", "url", valhallaURL)

	routingService := services.NewRoutingService(valhallaClient, incidentsService)

	server := api.NewServer(conf, logger, geocodingService, routingService)
	if err := server.Start(ctx); err != nil {
		return err
	}

	return nil
}
