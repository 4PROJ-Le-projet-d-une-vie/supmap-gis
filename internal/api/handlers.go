package api

import (
	"errors"
	"fmt"
	"github.com/matheodrd/httphelper/handler"
	"net/http"
	"supmap-gis/internal/domain/services"
	"supmap-gis/internal/providers/valhalla"
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

type RouteRequest struct {
	Locations      []valhalla.LocationRequest `json:"locations"`
	Costing        valhalla.Costing           `json:"costing"`
	CostingOptions *valhalla.CostingOptions   `json:"costing_options,omitempty"`
	Language       *string                    `json:"language,omitempty"`
	Alternates     *int                       `json:"alternates,omitempty"`
}

func (r RouteRequest) Validate() error {
	if len(r.Locations) < 2 {
		return errors.New("at least 2 locations must be provided")
	}
	if !r.Costing.IsValid() {
		return errors.New(fmt.Sprintf("costing %q is invalid", r.Costing))
	}
	return nil
}

// ToValhallaRequest converts a API request to a [valhalla.RouteRequest],
// and applies default values if necessary.
func (r RouteRequest) ToValhallaRequest() valhalla.RouteRequest {
	// Default values
	language := "fr-FR"
	alternates := 2

	if r.Language != nil {
		language = *r.Language
	}

	if r.Alternates != nil {
		alternates = *r.Alternates
	}

	return valhalla.RouteRequest{
		Locations:      r.Locations,
		Costing:        r.Costing,
		CostingOptions: *r.CostingOptions,
		Language:       language,
		Alternates:     alternates,
	}
}

func (s *Server) routeHandler() http.HandlerFunc {
	return handler.Handler(func(w http.ResponseWriter, r *http.Request) error {
		req, err := handler.Decode[RouteRequest](r)
		if err != nil {
			return handler.NewErrWithStatus(http.StatusBadRequest, err)
		}

		valhallaReq := req.ToValhallaRequest()

		route, err := s.routingService.CalculateRoute(r.Context(), valhallaReq)
		if err != nil {
			return handler.NewErrWithStatus(http.StatusInternalServerError, err)
		}

		resp := handler.Response[[]services.Trip]{
			Data:    route,
			Message: "success",
		}

		if err := handler.Encode[handler.Response[[]services.Trip]](resp, http.StatusOK, w); err != nil {
			return handler.NewErrWithStatus(http.StatusInternalServerError, err)
		}

		return nil
	})
}
