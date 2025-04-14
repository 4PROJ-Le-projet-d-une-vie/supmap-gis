package api

import (
	"fmt"
	"github.com/matheodrd/httphelper/handler"
	"net/http"
	"supmap-gis/internal/domain/services"
)

func (s *Server) geocodeHandler() http.HandlerFunc {
	return handler.Handler(func(w http.ResponseWriter, r *http.Request) error {
		address := r.URL.Query().Get("address")
		if address == "" {
			return handler.NewErrWithStatus(http.StatusBadRequest, fmt.Errorf("missing 'address' query parameter"))
		}

		result, err := s.geocodingService.Search(r.Context(), address)
		if err != nil {
			return handler.NewErrWithStatus(http.StatusInternalServerError, fmt.Errorf("geocoding address: %w", err))
		}

		resp := handler.Response[[]services.Place]{
			Data:    &result,
			Message: "success",
		}

		if err := handler.Encode[handler.Response[[]services.Place]](resp, http.StatusOK, w); err != nil {
			return handler.NewErrWithStatus(http.StatusInternalServerError, err)
		}

		return nil
	})
}
