package supmap_incidents

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
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

type RadiusMeter uint

func (c *Client) IncidentsInRadius(ctx context.Context, lat, lon float64, radius RadiusMeter) ([]Incident, error) {
	reqURL, err := url.Parse(c.baseURL + "/incidents")
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	query := reqURL.Query()
	query.Set("lat", fmt.Sprintf("%f", lat))
	query.Set("lon", fmt.Sprintf("%f", lon))
	query.Set("radius", fmt.Sprintf("%d", radius))
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

	var result []Incident
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result, nil
}
