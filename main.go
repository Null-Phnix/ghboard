package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/Null-Phnix/ghboard/config"
	"github.com/Null-Phnix/ghboard/ui"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "config error: %v\n", err)
		os.Exit(1)
	}

	if !cfg.HasGitHub() && !cfg.HasGitLab() {
		fmt.Println("ghboard — GitHub + GitLab Terminal Dashboard")
		fmt.Println()
		scanner := bufio.NewScanner(os.Stdin)

		// GitHub setup
		fmt.Println("GitHub:")
		fmt.Println("  Create a token at: https://github.com/settings/tokens")
		fmt.Println("  Required scopes: repo, notifications, read:user")
		fmt.Println()
		fmt.Print("  Paste GitHub token (or press Enter to skip): ")
		scanner.Scan()
		ghToken := strings.TrimSpace(scanner.Text())

		// GitLab setup
		fmt.Println()
		fmt.Println("GitLab:")
		fmt.Println("  Create a token at: https://gitlab.com/-/user_settings/personal_access_tokens")
		fmt.Println("  Required scopes: read_api, read_user")
		fmt.Println()
		fmt.Print("  Paste GitLab token (or press Enter to skip): ")
		scanner.Scan()
		glToken := strings.TrimSpace(scanner.Text())

		var glHost string
		if glToken != "" {
			fmt.Print("  GitLab host [gitlab.com]: ")
			scanner.Scan()
			glHost = strings.TrimSpace(scanner.Text())
			if glHost == "" {
				glHost = "gitlab.com"
			}
		}

		if ghToken == "" && glToken == "" {
			fmt.Fprintln(os.Stderr, "\nNo tokens provided. Exiting.")
			os.Exit(1)
		}

		var providers []config.Provider
		if ghToken != "" {
			providers = append(providers, config.Provider{Type: "github", Token: ghToken})
		}
		if glToken != "" {
			providers = append(providers, config.Provider{Type: "gitlab", Token: glToken, Host: glHost})
		}
		cfg.Providers = providers

		if err := config.Save(cfg); err != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to save config: %v\n", err)
		} else {
			fmt.Println("\n✓ Config saved to ~/.config/ghboard/config.json")
		}
		fmt.Println()
	}

	p := tea.NewProgram(ui.NewApp(cfg), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
