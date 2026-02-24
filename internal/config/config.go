package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

const (
	DefaultAPIURL = "https://api.ledgi.uk"
	configDir     = ".ledgi"
	configFile    = "config.json"
)

type Config struct {
	APIKey string `json:"api_key,omitempty"`
	APIURL string `json:"api_url,omitempty"`
}

func configPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, configDir, configFile)
}

// Load reads the config file from ~/.ledgi/config.json.
func Load() (*Config, error) {
	path := configPath()
	if path == "" {
		return &Config{}, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{}, nil
		}
		return nil, err
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// Save writes the config to ~/.ledgi/config.json.
func Save(cfg *Config) error {
	path := configPath()
	if path == "" {
		return os.ErrNotExist
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}

// Resolve returns the API key and URL using the priority:
// flag > env > config file > default.
func Resolve(flagKey, flagURL string) (apiKey, apiURL string) {
	cfg, _ := Load()

	// API Key: flag > env > config
	apiKey = flagKey
	if apiKey == "" {
		apiKey = os.Getenv("LEDGI_API_KEY")
	}
	if apiKey == "" && cfg != nil {
		apiKey = cfg.APIKey
	}

	// API URL: flag > env > config > default
	apiURL = flagURL
	if apiURL == "" {
		apiURL = os.Getenv("LEDGI_API_URL")
	}
	if apiURL == "" && cfg != nil && cfg.APIURL != "" {
		apiURL = cfg.APIURL
	}
	if apiURL == "" {
		apiURL = DefaultAPIURL
	}

	return apiKey, apiURL
}
