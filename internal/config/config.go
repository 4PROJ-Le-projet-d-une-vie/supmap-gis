package config

import (
	"fmt"
	"github.com/caarlos0/env/v11"
)

type Config struct {
	APIServerHost string `env:"API_SERVER_HOST"`
	APIServerPort string `env:"API_SERVER_PORT"`
	NominatimHost string `env:"NOMINATIM_HOST"`
	NominatimPort string `env:"NOMINATIM_PORT"`
	ValhallaHost  string `env:"VALHALLA_HOST"`
	ValhallaPort  string `env:"VALHALLA_PORT"`
}

func New() (*Config, error) {
	cfg, err := env.ParseAs[Config]()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}
	return &cfg, nil
}
