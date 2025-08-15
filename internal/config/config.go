package config

import (
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Config represents persisted configuration values.
type Config struct {
	APIBase string `toml:"apiBase"`
	UserID  string `toml:"userId"`
}

func configDir() string {
	// Explicitly use XDG-style path regardless of platform as requested.
	home, err := os.UserHomeDir()
	if err != nil {
		return ".config/markdex" // fallback relative
	}
	return filepath.Join(home, ".config", "markdex")
}

func configPath() string { return filepath.Join(configDir(), "config.toml") }

// Path returns the absolute path to the configuration file (may be relative if home directory lookup failed).
func Path() string { return configPath() }

// Load reads configuration from config.toml. It returns an empty Config and the error
// if the file cannot be read (callers commonly ignore the error to allow empty defaults).
func Load() (*Config, error) {
	vp := viper.New()
	vp.SetConfigFile(configPath())
	vp.SetConfigType("toml")
	if err := vp.ReadInConfig(); err != nil {
		// On error (e.g., file missing), still return a config with sensible defaults.
		return &Config{UserID: "default"}, err
	}
	c := &Config{
		APIBase: vp.GetString("apiBase"),
		UserID:  vp.GetString("userId"),
	}
	if c.UserID == "" {
		c.UserID = "default"
	}
	return c, nil
}

// Save persists the configuration to config.toml using TOML format via Viper.
func Save(c *Config) error {
	if err := os.MkdirAll(configDir(), 0o755); err != nil {
		return err
	}
	vp := viper.New()
	vp.Set("apiBase", c.APIBase)
	vp.Set("userId", c.UserID)
	vp.SetConfigFile(configPath())
	vp.SetConfigType("toml")
	// WriteConfig will create or truncate the file.
	return vp.WriteConfig()
}
