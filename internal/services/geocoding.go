package services

import (
	"context"
	"fmt"
	"strconv"
	"supmap-gis/internal/providers/nominatim"
)

type GeocodingClient interface {
	Search(ctx context.Context, address string) ([]nominatim.GeocodeResult, error)
	Reverse(ctx context.Context, lat, lon float64) (*nominatim.ReverseResult, error)
}

type GeocodingService struct {
	client GeocodingClient
}

func NewGeocodingService(client GeocodingClient) *GeocodingService {
	return &GeocodingService{client: client}
}

type Place struct {
	Lat         float64 `json:"lat"`
	Lon         float64 `json:"lon"`
	Name        string  `json:"name"`
	DisplayName string  `json:"display_name"`
}

func (s *GeocodingService) Search(ctx context.Context, address string) ([]Place, error) {
	resp, err := s.client.Search(ctx, address)
	if err != nil {
		return nil, fmt.Errorf("searching address %q: %w", address, err)
	}

	if len(resp) == 0 {
		return []Place{}, nil
	}

	places := make([]Place, len(resp))
	for i, geocodeResult := range resp {
		lat, err := strconv.ParseFloat(geocodeResult.Lat, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse latitude: %w", err)
		}
		lon, err := strconv.ParseFloat(geocodeResult.Lon, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse longitude: %w", err)
		}

		places[i] = Place{
			Lat:         lat,
			Lon:         lon,
			Name:        geocodeResult.Name,
			DisplayName: geocodeResult.DisplayName,
		}
	}

	return places, nil
}

func (s *GeocodingService) Reverse(ctx context.Context, lat, lon float64) (*nominatim.ReverseResult, error) {
	return s.client.Reverse(ctx, lat, lon)
}
