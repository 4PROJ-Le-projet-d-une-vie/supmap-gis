package api

import (
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

type Maneuver struct {
	Type                uint8    `json:"type"`
	Instruction         string   `json:"instruction"`
	StreetNames         []string `json:"street_names"`
	Time                float64  `json:"time"`
	Length              float64  `json:"length"`
	RoundaboutExitCount *uint8   `json:"roundabout_exit_count,omitempty"`
}

type Summary struct {
	Time   float64 `json:"time"`
	Length float64 `json:"length"`
}

type Leg struct {
	Maneuvers []Maneuver       `json:"maneuvers"`
	Summary   Summary          `json:"summary"`
	Shape     []services.Point `json:"shape"`
}

type Trip struct {
	Locations []valhalla.LocationResponse `json:"locations"`
	Legs      []Leg                       `json:"legs"`
	Summary   Summary                     `json:"summary"`
}

func (s *Server) routeHandler() http.HandlerFunc {
	return handler.Handler(func(w http.ResponseWriter, r *http.Request) error {
		req, err := handler.Decode[valhalla.RouteRequest](r)
		if err != nil {
			return handler.NewErrWithStatus(http.StatusBadRequest, err)
		}

		route, err := s.routingService.CalculateRoute(r.Context(), req)
		if err != nil {
			return handler.NewErrWithStatus(http.StatusInternalServerError, err)
		}

		respTrips := make([]Trip, 0, 1)
		// Main trip
		mainTrip, err := convertTrip(route.Trip)
		if err != nil {
			return handler.NewErrWithStatus(http.StatusInternalServerError, err)
		}
		respTrips = append(respTrips, *mainTrip)
		// Alternative trips
		for _, altTrip := range route.Alternates {
			trip, err := convertTrip(altTrip.Trip)
			if err != nil {
				return handler.NewErrWithStatus(http.StatusInternalServerError, err)
			}
			respTrips = append(respTrips, *trip)
		}

		resp := handler.Response[[]Trip]{
			Data:    &respTrips,
			Message: "success",
		}

		if err := handler.Encode[handler.Response[[]Trip]](resp, http.StatusOK, w); err != nil {
			return handler.NewErrWithStatus(http.StatusInternalServerError, err)
		}

		return nil
	})
}

// --- Mapping : Provider -> API ---

func convertTrip(trip valhalla.Trip) (*Trip, error) {
	legs := make([]Leg, len(trip.Legs))
	for i, leg := range trip.Legs {
		convertedLeg, err := convertLeg(leg)
		if err != nil {
			return nil, fmt.Errorf("legs[%d]: %w", i, err)
		}
		legs[i] = *convertedLeg
	}
	return &Trip{
		Locations: trip.Locations,
		Legs:      legs,
		Summary: Summary{
			Time:   trip.Summary.Time,
			Length: trip.Summary.Length,
		},
	}, nil
}

func convertLeg(leg valhalla.Leg) (*Leg, error) {
	maneuvers := make([]Maneuver, len(leg.Maneuvers))
	for i, m := range leg.Maneuvers {
		maneuvers[i] = Maneuver{
			Type:                m.Type,
			Instruction:         m.Instruction,
			StreetNames:         m.StreetNames,
			Time:                m.Time,
			Length:              m.Length,
			RoundaboutExitCount: m.RoundaboutExitCount,
		}
	}

	shape, err := services.DecodePolyline(leg.Shape, 6)
	if err != nil {
		return nil, fmt.Errorf("convertLeg: %w", err)
	}

	return &Leg{
		Maneuvers: maneuvers,
		Summary: Summary{
			Time:   leg.Summary.Time,
			Length: leg.Summary.Length,
		},
		Shape: shape,
	}, nil
}
