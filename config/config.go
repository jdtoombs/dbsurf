// config.go handles loading and saving database connection configurations
// to ~/.config/dbsurf/config.json.
package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// LastUsed will be used for sorting list in recent connections on dashboard
type Connection struct {
	Name       string    `json:"name"`
	ConnString string    `json:"conn_string"`
	DBType     string    `json:"db_type"`
	LastUsed   time.Time `json:"last_used"`
}

type Config struct {
	Connections []Connection `json:"connections"`
}

func getConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	configDir := filepath.Join(home, ".config", "dbsurf")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return "", err
	}
	return filepath.Join(configDir, "config.json"), nil
}

func Load() (*Config, error) {
	path, err := getConfigPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return &Config{Connections: []Connection{}}, nil
	}
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (c *Config) Save() error {
	path, err := getConfigPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

func (c *Config) AddConnection(name, connString, dbType string) {
	for i, conn := range c.Connections {
		if conn.ConnString == connString {
			c.Connections[i].Name = name
			c.Connections[i].LastUsed = time.Now()
			return
		}
	}
	c.Connections = append(c.Connections, Connection{
		Name:       name,
		ConnString: connString,
		DBType:     dbType,
		LastUsed:   time.Now(),
	})
}

func (c *Config) UpdateLastUsed(connString string) {
	for i, conn := range c.Connections {
		if conn.ConnString == connString {
			c.Connections[i].LastUsed = time.Now()
			return
		}
	}
}
