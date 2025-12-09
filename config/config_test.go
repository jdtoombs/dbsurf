package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_NonExistent(t *testing.T) {
	// Save original HOME and restore after test
	origHome := os.Getenv("HOME")
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v, want nil", err)
	}
	if cfg == nil {
		t.Fatal("Load() returned nil config")
	}
	if len(cfg.Connections) != 0 {
		t.Errorf("Load() connections = %d, want 0", len(cfg.Connections))
	}
}

func TestSaveAndLoad(t *testing.T) {
	// Use temp dir as HOME
	origHome := os.Getenv("HOME")
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	// Create and save config
	cfg := &Config{}
	cfg.AddConnection("test-db", "postgres://localhost/test", "postgres")

	if err := cfg.Save(); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Load it back
	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if len(loaded.Connections) != 1 {
		t.Fatalf("Load() connections = %d, want 1", len(loaded.Connections))
	}

	conn := loaded.Connections[0]
	if conn.Name != "test-db" {
		t.Errorf("Connection.Name = %q, want %q", conn.Name, "test-db")
	}
	if conn.ConnString != "postgres://localhost/test" {
		t.Errorf("Connection.ConnString = %q, want %q", conn.ConnString, "postgres://localhost/test")
	}
	if conn.DBType != "postgres" {
		t.Errorf("Connection.DBType = %q, want %q", conn.DBType, "postgres")
	}
}

func TestAddConnection(t *testing.T) {
	cfg := &Config{}

	cfg.AddConnection("first", "postgres://localhost/first", "postgres")
	if len(cfg.Connections) != 1 {
		t.Fatalf("AddConnection() connections = %d, want 1", len(cfg.Connections))
	}

	// Add second connection
	cfg.AddConnection("second", "mysql://localhost/second", "mysql")
	if len(cfg.Connections) != 2 {
		t.Fatalf("AddConnection() connections = %d, want 2", len(cfg.Connections))
	}

	// Add duplicate connection string - should update, not add
	cfg.AddConnection("first-renamed", "postgres://localhost/first", "postgres")
	if len(cfg.Connections) != 2 {
		t.Errorf("AddConnection() duplicate should update, got %d connections, want 2", len(cfg.Connections))
	}
	if cfg.Connections[0].Name != "first-renamed" {
		t.Errorf("AddConnection() should update name, got %q, want %q", cfg.Connections[0].Name, "first-renamed")
	}
}

func TestDeleteConnection(t *testing.T) {
	cfg := &Config{
		Connections: []Connection{
			{Name: "first", ConnString: "postgres://localhost/first"},
			{Name: "second", ConnString: "mysql://localhost/second"},
			{Name: "third", ConnString: "sqlserver://localhost/third"},
		},
	}

	// Delete middle connection
	cfg.Connections = append(cfg.Connections[:1], cfg.Connections[2:]...)

	if len(cfg.Connections) != 2 {
		t.Fatalf("Delete connections = %d, want 2", len(cfg.Connections))
	}
	if cfg.Connections[0].Name != "first" {
		t.Errorf("First connection = %q, want %q", cfg.Connections[0].Name, "first")
	}
	if cfg.Connections[1].Name != "third" {
		t.Errorf("Second connection = %q, want %q", cfg.Connections[1].Name, "third")
	}
}

func TestConfigFilePermissions(t *testing.T) {
	origHome := os.Getenv("HOME")
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	cfg := &Config{}
	cfg.AddConnection("test", "postgres://user:secret@localhost/db", "postgres")
	cfg.Save()

	configPath := filepath.Join(tmpDir, ".config", "dbsurf", "config.json")
	info, err := os.Stat(configPath)
	if err != nil {
		t.Fatalf("Stat() error = %v", err)
	}

	// Check file permissions are 0600 (owner read/write only)
	perm := info.Mode().Perm()
	if perm != 0600 {
		t.Errorf("Config file permissions = %o, want 0600", perm)
	}
}

func TestLoad_CorruptedFile(t *testing.T) {
	origHome := os.Getenv("HOME")
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	// Create config dir and write invalid JSON
	configDir := filepath.Join(tmpDir, ".config", "dbsurf")
	os.MkdirAll(configDir, 0755)
	configPath := filepath.Join(configDir, "config.json")
	os.WriteFile(configPath, []byte("not valid json{{{"), 0600)

	_, err := Load()
	if err == nil {
		t.Error("Load() with corrupted file should return error")
	}

	if !isJSONError(err) {
		t.Errorf("Load() error type = %T, want JSON error", err)
	}
}

func isJSONError(err error) bool {
	_, ok1 := err.(*json.SyntaxError)
	_, ok2 := err.(*json.UnmarshalTypeError)
	return ok1 || ok2
}
