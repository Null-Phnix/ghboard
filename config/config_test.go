// config/config_test.go
package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadFromEnv_GitHub(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "test-token-123")
	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tok := cfg.GitHubToken(); tok != "test-token-123" {
		t.Errorf("expected github token 'test-token-123', got %q", tok)
	}
}

func TestLoadFromEnv_GitLab(t *testing.T) {
	t.Setenv("GITLAB_TOKEN", "glpat-abc")
	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	tok, host := cfg.GitLabToken()
	if tok != "glpat-abc" {
		t.Errorf("expected gitlab token 'glpat-abc', got %q", tok)
	}
	if host != "gitlab.com" {
		t.Errorf("expected host 'gitlab.com', got %q", host)
	}
}

func TestLoadFromEnv_GitLabCustomHost(t *testing.T) {
	t.Setenv("GITLAB_TOKEN", "glpat-xyz")
	t.Setenv("GITLAB_HOST", "gitlab.example.com")
	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, host := cfg.GitLabToken()
	if host != "gitlab.example.com" {
		t.Errorf("expected host 'gitlab.example.com', got %q", host)
	}
}

func TestLoadFromFile_MultiProvider(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.json")
	os.WriteFile(cfgPath, []byte(`{
  "providers": [
    {"type": "github", "token": "ghp-123"},
    {"type": "gitlab", "token": "glpat-456", "host": "gitlab.com"}
  ]
}`), 0600)
	t.Setenv("GITHUB_TOKEN", "")
	t.Setenv("GITLAB_TOKEN", "")
	t.Setenv("GHBOARD_CONFIG", cfgPath)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !cfg.HasGitHub() {
		t.Fatal("expected HasGitHub() = true")
	}
	if !cfg.HasGitLab() {
		t.Fatal("expected HasGitLab() = true")
	}
	if tok := cfg.GitHubToken(); tok != "ghp-123" {
		t.Errorf("expected github token 'ghp-123', got %q", tok)
	}
	if tok, host := cfg.GitLabToken(); tok != "glpat-456" || host != "gitlab.com" {
		t.Errorf("expected gitlab token 'glpat-456' host 'gitlab.com', got %q / %q", tok, host)
	}
}

func TestLoadFromFile_LegacyMigration(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.json")
	os.WriteFile(cfgPath, []byte(`{"token":"old-legacy-token"}`), 0600)
	t.Setenv("GITHUB_TOKEN", "")
	t.Setenv("GHBOARD_CONFIG", cfgPath)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !cfg.HasGitHub() {
		t.Fatal("legacy token should migrate to GitHub provider")
	}
	if tok := cfg.GitHubToken(); tok != "old-legacy-token" {
		t.Errorf("expected 'old-legacy-token', got %q", tok)
	}
}
