package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Env   string `yaml:"env"`
	Bybit Bybit  `yaml:"bybit"`
	Trade Trade  `yaml:"trade"`
}

type Bybit struct {
	ApiKey    string `yaml:"api_key"`
	ApiSecret string `yaml:"api_secret"`
	Testnet   bool   `yaml:"testnet"`
}

type Trade struct {
	Symbol    string  `yaml:"symbol"`
	Timeframe string  `yaml:"timeframe"`
	Qty       float64 `yaml:"qty"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
