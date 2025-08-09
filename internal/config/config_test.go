package config_test

import (
	"os"
	"testing"

	"github.com/sayuyere/storageX/internal/config"
)

func resetConfigSingleton() {
	config.ResetConfigSingleton()
}

func TestLoadConfig_Defaults(t *testing.T) {
	resetConfigSingleton()
	cfg, err := config.LoadConfig("/nonexistent/path/config.json")
	if err == nil && cfg == nil {
		t.Fatal("expected default config, got nil")
	}
	if cfg.ChunkSize == 0 {
		t.Error("expected nonzero chunk size from defaults")
	}
}

func TestLookupSecrets(t *testing.T) {
	resetConfigSingleton()
	os.Setenv("TEST_DROPBOX_TOKEN", "token123")
	cfg := &config.AppConfig{
		Cloud: config.CloudConfig{
			DropboxAccessTokens: []string{"TEST_DROPBOX_TOKEN", ""},
		},
	}
	config.LookupSecrets(cfg)
	if cfg.Cloud.DropboxAccessTokens[0] != "token123" {
		t.Errorf("expected token123, got %q", cfg.Cloud.DropboxAccessTokens[0])
	}
	if cfg.Cloud.DropboxAccessTokens[1] != "" {
		t.Errorf("expected empty string for missing token, got %q", cfg.Cloud.DropboxAccessTokens[1])
	}
}

func TestLoadConfig_JSON(t *testing.T) {
	resetConfigSingleton()
	f, err := os.CreateTemp("", "config-test-*.json")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(f.Name())
	jsonData := `{"chunk_size": 42, "cloud": {"dropbox_access_tokens": ["TEST_DROPBOX_TOKEN"]}, "log": {"debug": true}, "metadata": {"db_path": "test.db"}}`
	if _, err := f.Write([]byte(jsonData)); err != nil {
		t.Fatalf("failed to write temp config: %v", err)
	}
	if err := f.Close(); err != nil {
		t.Fatalf("failed to close temp config: %v", err)
	}
	os.Setenv("TEST_DROPBOX_TOKEN", "tokenXYZ")
	cfg, err := config.LoadConfig(f.Name())
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}
	if cfg.ChunkSize != 42 {
		t.Errorf("expected chunk size 42, got %d", cfg.ChunkSize)
	}
	if cfg.Cloud.DropboxAccessTokens[0] != "tokenXYZ" {
		t.Errorf("expected tokenXYZ, got %q", cfg.Cloud.DropboxAccessTokens[0])
	}
	if !cfg.Log.Debug {
		t.Error("expected debug true")
	}
	if cfg.Meta.DBPath != "test.db" {
		t.Errorf("expected db_path test.db, got %q", cfg.Meta.DBPath)
	}
}
