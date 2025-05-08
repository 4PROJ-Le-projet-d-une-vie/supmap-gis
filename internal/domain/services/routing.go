package services

import (
	"context"
	"fmt"
	"supmap-gis/internal/providers/valhalla"
)

type RoutingClient interface {
	CalculateRoute(ctx context.Context, routeRequest valhalla.RouteRequest) (*valhalla.RouteResponse, error)
}

type RoutingService struct {
	client           RoutingClient
	incidentsService *IncidentsService
}

func NewRoutingService(client RoutingClient, incidentsService *IncidentsService) *RoutingService {
	return &RoutingService{client: client, incidentsService: incidentsService}
}

func (s *RoutingService) CalculateRoute(ctx context.Context, routeRequest valhalla.RouteRequest) (*[]Trip, error) {
	locationsPoints := extractPointsFromLocations(routeRequest.Locations)
	incidents := s.incidentsService.IncidentsAroundLocations(ctx, locationsPoints)
	excludes := pointsToExcludeLocations(incidents)

	// Add incidents coordinates to the locations to avoid
	routeRequest.ExcludeLocations = append(routeRequest.ExcludeLocations, excludes...)
	fmt.Println(routeRequest.ExcludeLocations)

	vRoute, err := s.client.CalculateRoute(ctx, routeRequest)
	if err != nil {
		return nil, fmt.Errorf("calculate route: %w", err)
	}

	respTrips := make([]Trip, 0, 1)
	// Main trip
	mainTrip, err := MapValhallaTrip(vRoute.Trip)
	if err != nil {
		return nil, fmt.Errorf("MapValhallaTrip: %w", err)
	}
	respTrips = append(respTrips, *mainTrip)
	// Alternative trips
	for _, altTrip := range vRoute.Alternates {
		trip, err := MapValhallaTrip(altTrip.Trip)
		if err != nil {
			return nil, fmt.Errorf("MapValhallaTrip: %w", err)
		}
		respTrips = append(respTrips, *trip)
	}

	return &respTrips, nil
}

func extractPointsFromLocations(locations []valhalla.LocationRequest) []Point {
	points := make([]Point, 0, len(locations))
	for _, loc := range locations {
		points = append(points, Point{
			Lat: loc.Lat,
			Lon: loc.Lon,
		})
	}
	return points
}

// --- DTOs ---

type Point struct {
	Lat float64 `json:"latitude"`
	Lon float64 `json:"longitude"`
}

type Maneuver struct {
	Type                uint8    `json:"type"`
	Instruction         string   `json:"instruction"`
	StreetNames         []string `json:"street_names"`
	Time                float64  `json:"time"`
	Length              float64  `json:"length"`
	BeginShapeIndex     uint     `json:"begin_shape_index"`
	EndShapeIndex       uint     `json:"end_shape_index"`
	RoundaboutExitCount *uint8   `json:"roundabout_exit_count,omitempty"`
}

type Summary struct {
	Time   float64 `json:"time"`
	Length float64 `json:"length"`
}

type Leg struct {
	Maneuvers []Maneuver `json:"maneuvers"`
	Summary   Summary    `json:"summary"`
	Shape     []Point    `json:"shape"`
}

type Trip struct {
	Locations []valhalla.LocationResponse `json:"locations"`
	Legs      []Leg                       `json:"legs"`
	Summary   Summary                     `json:"summary"`
}

// --- Mapping Valhalla -> DTO ---

// MapValhallaTrip maps Valhalla's [valhalla.Trip] struct to a service DTO [Trip] struct.
func MapValhallaTrip(vt valhalla.Trip) (*Trip, error) {
	legs := make([]Leg, len(vt.Legs))
	for i, leg := range vt.Legs {
		convertedLeg, err := mapValhallaLeg(leg)
		if err != nil {
			return nil, fmt.Errorf("legs[%d]: %w", i, err)
		}
		legs[i] = *convertedLeg
	}
	return &Trip{
		Locations: vt.Locations,
		Legs:      legs,
		Summary: Summary{
			Time:   vt.Summary.Time,
			Length: vt.Summary.Length,
		},
	}, nil
}

// mapValhallaLeg maps Valhalla's [valhalla.Leg] struct to a service DTO [Leg] struct.
func mapValhallaLeg(vl valhalla.Leg) (*Leg, error) {
	maneuvers := make([]Maneuver, len(vl.Maneuvers))
	for i, m := range vl.Maneuvers {
		maneuvers[i] = Maneuver{
			Type:                m.Type,
			Instruction:         m.Instruction,
			StreetNames:         m.StreetNames,
			Time:                m.Time,
			Length:              m.Length,
			BeginShapeIndex:     m.BeginShapeIndex,
			EndShapeIndex:       m.EndShapeIndex,
			RoundaboutExitCount: m.RoundaboutExitCount,
		}

		// initialize StreetNames to an empty slice instead of returning a nil value
		if maneuvers[i].StreetNames == nil {
			maneuvers[i].StreetNames = []string{}
		}
	}

	shape, err := DecodePolyline(vl.Shape, 6)
	if err != nil {
		return nil, fmt.Errorf("mapValhallaLeg: %w", err)
	}

	return &Leg{
		Maneuvers: maneuvers,
		Summary: Summary{
			Time:   vl.Summary.Time,
			Length: vl.Summary.Length,
		},
		Shape: shape,
	}, nil
}

// Conversion []Point -> []valhalla.ExcludeLocations
func pointsToExcludeLocations(points []Point) []valhalla.ExcludeLocations {
	excludes := make([]valhalla.ExcludeLocations, 0, len(points))
	for _, pt := range points {
		excludes = append(excludes, valhalla.ExcludeLocations{
			Lat: pt.Lat,
			Lon: pt.Lon,
		})
	}
	return excludes
}
