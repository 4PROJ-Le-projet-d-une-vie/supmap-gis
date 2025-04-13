package api

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"supmap-gis/internal/config"
	"supmap-gis/internal/domain/services"
	"sync"
	"time"
)

type Server struct {
	Config           *config.Config
	logger           *slog.Logger
	geocodingService *services.GeocodingService
}

func NewServer(config *config.Config, logger *slog.Logger, geocodingService *services.GeocodingService) *Server {
	return &Server{
		Config:           config,
		logger:           logger,
		geocodingService: geocodingService,
	}
}

func (s *Server) health(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Add("Cache-Control", "no-cache, no-store, must-revalidate;")
	if _, err := w.Write([]byte("API server is started.")); err != nil {
		s.logger.Error(fmt.Sprintf("Error writing response: %v", err))
	}
}

func (s *Server) Start(ctx context.Context) error {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", s.health)
	mux.HandleFunc("GET /geocode", s.geocodeHandler())

	server := &http.Server{
		Addr:    net.JoinHostPort(s.Config.APIServerHost, s.Config.APIServerPort),
		Handler: mux,
	}

	go func() {
		s.logger.Info("API server is running", "port", s.Config.APIServerPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Error("API server failed to listen and serve", "error", err)
		}
	}()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-ctx.Done()

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			s.logger.Error("API server failed to shutdown", "error", err)
		}
	}()

	wg.Wait()
	return nil
}
