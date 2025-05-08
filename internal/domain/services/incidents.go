package services

import (
	"context"
	"fmt"
	"log"
	"math"
	supmapIncidents "supmap-gis/internal/providers/supmap-incidents"
)

type IncidentsClient interface {
	IncidentsInRadius(ctx context.Context, lat, lon float64, radius supmapIncidents.RadiusMeter) ([]supmapIncidents.Incident, error)
}

type IncidentsService struct {
	client IncidentsClient
}

func NewIncidentsService(client IncidentsClient) *IncidentsService {
	return &IncidentsService{client: client}
}

func (s *IncidentsService) IncidentsAroundLocations(ctx context.Context, locations []Point) []Point {
	centerLat, centerLon, radius := computeLocationsBoundingCircle(locations)

	fmt.Println(centerLat, centerLon, radius)

	incidents, err := s.client.IncidentsInRadius(ctx, centerLat, centerLon, radius)
	if err != nil {
		log.Printf("Error getting incidents: %v", err)
		return []Point{}
	}

	fmt.Println(incidents)

	incidentsPoints := make([]Point, 0, len(incidents))
	for _, incident := range incidents {
		incidentsPoints = append(incidentsPoints, Point{
			Lat: incident.Latitude,
			Lon: incident.Longitude,
		})
	}

	return incidentsPoints
}

// computeLocationsBoundingCircle calcule un cercle englobant tous les points de locations.
func computeLocationsBoundingCircle(locations []Point) (centerLat, centerLon float64, radius supmapIncidents.RadiusMeter) {
	var sumLat, sumLon float64
	for _, loc := range locations {
		sumLat += loc.Lat
		sumLon += loc.Lon
	}
	centerLat = sumLat / float64(len(locations))
	centerLon = sumLon / float64(len(locations))

	maxDist := 0.0
	for _, loc := range locations {
		dist := haversine(centerLat, centerLon, loc.Lat, loc.Lon)
		if dist > maxDist {
			maxDist = dist
		}
	}
	return centerLat, centerLon, supmapIncidents.RadiusMeter(maxDist * 1.6)
}

// Haversine distance in meters
func haversine(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371000 // m
	phi1 := lat1 * math.Pi / 180
	phi2 := lat2 * math.Pi / 180
	dPhi := (lat2 - lat1) * math.Pi / 180
	dLambda := (lon2 - lon1) * math.Pi / 180

	a := math.Sin(dPhi/2)*math.Sin(dPhi/2) +
		math.Cos(phi1)*math.Cos(phi2)*math.Sin(dLambda/2)*math.Sin(dLambda/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return R * c
}
