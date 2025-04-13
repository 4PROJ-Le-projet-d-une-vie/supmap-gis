package api

import (
	"fmt"
	"net/http"
	"supmap-gis/internal/domain/services"
)

type Response[T any] struct {
	Data    *T     `json:"data,omitempty"`
	Message string `json:"message,omitempty"`
}

func (s *Server) geocodeHandler() http.HandlerFunc {
	return handler(func(w http.ResponseWriter, r *http.Request) error {
		address := r.URL.Query().Get("address")
		if address == "" {
			return NewErrWithStatus(http.StatusBadRequest, fmt.Errorf("missing 'address' query parameter"))
		}

		result, err := s.geocodingService.Search(r.Context(), address)
		if err != nil {
			return NewErrWithStatus(http.StatusInternalServerError, fmt.Errorf("geocoding address: %w", err))
		}

		resp := Response[[]services.Place]{
			Data:    &result,
			Message: "success",
		}

		if err := encode[Response[[]services.Place]](resp, http.StatusOK, w); err != nil {
			return NewErrWithStatus(http.StatusInternalServerError, err)
		}

		return nil
	})
}
