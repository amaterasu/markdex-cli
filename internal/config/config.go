package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

type Config struct {
	APIBase string `json:"apiBase"`
}

func configPath() string {
	d, err := os.UserConfigDir()
	if err != nil {
		return ".markdex-config.json"
	}
	return filepath.Join(d, "markdex", "config.json")
}

func Load() (*Config, error) {
	b, err := os.ReadFile(configPath())
	if err != nil {
		return &Config{}, err
	}
	var c Config
	return &c, json.Unmarshal(b, &c)
}

func Save(c *Config) error {
	if c.APIBase == "" {
		return errors.New("apiBase required")
	}
	p := configPath()
	_ = os.MkdirAll(filepath.Dir(p), 0o755)
	b, _ := json.MarshalIndent(c, "", "  ")
	return os.WriteFile(p, b, 0o600)
}
