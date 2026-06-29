package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	APIEndpoint    string `json:"api_endpoint"`
	IngestEndpoint string `json:"ingest_endpoint"`
	WebClientURL   string `json:"web_client_url"`
	AuthToken      string `json:"auth_token"`
	DeviceID       string `json:"device_id"`
	DefaultForward string `json:"default_forward_url"`
}

func Defaults() *Config {
	return &Config{
		APIEndpoint:    "https://glitch-hunt-central-api.my-basement.cloud",
		IngestEndpoint: "https://glitch-hunt-ingestion.my-basement.cloud",
		WebClientURL:   "https://glitch-hunt.my-basement.cloud",
		DefaultForward: "http://localhost:8080/webhooks",
		DeviceID:       "",
		AuthToken:      "",
	}
}

func ConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".config", "hunt")
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.json"), nil
}

func Load() (*Config, error) {
	cfg := Defaults()
	path, err := ConfigPath()
	if err != nil {
		return cfg, nil
	}
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, err
	}
	defer f.Close()
	if err := json.NewDecoder(f).Decode(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func Save(cfg *Config) error {
	path, err := ConfigPath()
	if err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	encoder := json.NewEncoder(f)
	encoder.SetIndent("", "  ")
	return encoder.Encode(cfg)
}
