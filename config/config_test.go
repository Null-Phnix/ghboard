// config/config_test.go
package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadFromEnv(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "test-token-123")
	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Token != "test-token-123" {
		t.Errorf("expected token 'test-token-123', got %q", cfg.Token)
	}
}

func TestLoadFromFile(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.json")
	os.WriteFile(cfgPath, []byte(`{"token":"file-token-456"}`), 0600)
	t.Setenv("GITHUB_TOKEN", "")
	t.Setenv("GHBOARD_CONFIG", cfgPath)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Token != "file-token-456" {
		t.Errorf("expected 'file-token-456', got %q", cfg.Token)
	}
}
