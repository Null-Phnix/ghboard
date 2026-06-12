// config/config.go
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Provider struct {
	Type  string `json:"type"`
	Token string `json:"token"`
	Host  string `json:"host,omitempty"`
}

type Config struct {
	Token     string     `json:"token"`     // legacy single-token field
	Providers []Provider `json:"providers,omitempty"`
}

func configPath() string {
	if p := os.Getenv("GHBOARD_CONFIG"); p != "" {
		return p
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "ghboard", "config.json")
}

func Load() (*Config, error) {
	// 1. Env vars take priority
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		return &Config{
			Providers: []Provider{{Type: "github", Token: token}},
		}, nil
	}
	if token := os.Getenv("GITLAB_TOKEN"); token != "" {
		host := os.Getenv("GITLAB_HOST")
		if host == "" {
			host = "gitlab.com"
		}
		return &Config{
			Providers: []Provider{{Type: "gitlab", Token: token, Host: host}},
		}, nil
	}

	// 2. Config file
	path := configPath()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{}, nil
		}
		return nil, fmt.Errorf("reading config: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}
	// Migrate legacy single-token config
	if cfg.Token != "" && len(cfg.Providers) == 0 {
		cfg.Providers = []Provider{{Type: "github", Token: cfg.Token}}
		cfg.Token = ""
	}
	return &cfg, nil
}

func (c *Config) GitHubToken() string {
	for _, p := range c.Providers {
		if p.Type == "github" && p.Token != "" {
			return p.Token
		}
	}
	return c.Token // fallback
}

func (c *Config) GitLabToken() (token, host string) {
	for _, p := range c.Providers {
		if p.Type == "gitlab" && p.Token != "" {
			host := p.Host
			if host == "" {
				host = "gitlab.com"
			}
			return p.Token, host
		}
	}
	return "", ""
}

func (c *Config) HasGitHub() bool { return c.GitHubToken() != "" }
func (c *Config) HasGitLab() bool {
	t, _ := c.GitLabToken()
	return t != ""
}

func Save(cfg *Config) error {
	path := configPath()
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}
