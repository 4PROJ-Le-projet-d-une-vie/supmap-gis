package services

import (
	"fmt"
	"io"
	"math"
)

// DecodePolyline decodes a Google encoded Polyline string into a slice of coordinates.
// The precision parameter defines the number of decimal digits used when encoding.
// If precision is zero or negative, the function defaults to 6 digits (1e-6 precision).
//
// It returns a slice of [Point] structs, each containing latitude and longitude in
// decimal degrees, or an error if the input string is malformed.
func DecodePolyline(encoded string, precision int) ([]Point, error) {
	if precision <= 0 {
		precision = 6
	}
	factor := math.Pow10(precision)

	idx := 0
	lat, lng := 0, 0
	var coords []Point

	for idx < len(encoded) {
		dLat, err := readDelta(encoded, &idx)
		if err != nil {
			return nil, fmt.Errorf("decodePolyline: failed reading latitude at idx %d: %w", idx, err)
		}
		lat += dLat

		dLng, err := readDelta(encoded, &idx)
		if err != nil {
			return nil, fmt.Errorf("decodePolyline: failed reading longitude at idx %d: %w", idx, err)
		}
		lng += dLng

		coords = append(coords, Point{
			Lat: float64(lat) / factor,
			Lon: float64(lng) / factor,
		})
	}

	return coords, nil
}

func readDelta(s string, idx *int) (int, error) {
	var result, shift, b int
	for {
		if *idx >= len(s) {
			return 0, io.ErrUnexpectedEOF
		}
		b = int(s[*idx]) - 63
		*idx++
		result |= (b & 0x1f) << shift
		shift += 5
		if b < 0x20 {
			break
		}
	}
	if result&1 != 0 {
		return ^(result >> 1), nil
	}
	return result >> 1, nil
}
