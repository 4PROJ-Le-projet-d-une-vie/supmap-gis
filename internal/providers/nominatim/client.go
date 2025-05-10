package nominatim

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
}

type ClientOptions struct {
	Timeout time.Duration
}

func DefaultClientOptions() ClientOptions {
	return ClientOptions{
		Timeout: 7 * time.Second,
	}
}

func NewClient(baseURL string, options ...ClientOptions) *Client {
	opts := DefaultClientOptions()
	if len(options) > 0 {
		opts = options[0]
	}

	return &Client{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: opts.Timeout},
	}
}

func (c *Client) Search(ctx context.Context, address string) ([]GeocodeResult, error) {
	reqURL, err := url.Parse(c.baseURL + "/search")
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	query := reqURL.Query()
	query.Set("q", address)
	reqURL.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to build request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Language", "fr-FR")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var result []GeocodeResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result, nil
}

func (c *Client) Reverse(ctx context.Context, lat, lon float64) (*ReverseResult, error) {
	reqURL, err := url.Parse(c.baseURL + "/reverse")
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	query := reqURL.Query()
	query.Set("format", "geojson")
	query.Set("lat", strconv.FormatFloat(lat, 'g', -1, 64))
	query.Set("lon", strconv.FormatFloat(lon, 'g', -1, 64))
	reqURL.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to build request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var result ReverseResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}
