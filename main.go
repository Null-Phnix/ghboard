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

	if cfg.Token == "" {
		fmt.Println("ghboard — GitHub Terminal Dashboard")
		fmt.Println()
		fmt.Println("No GitHub token found.")
		fmt.Println("Create one at: https://github.com/settings/tokens")
		fmt.Println("Required scopes: repo, notifications, read:user")
		fmt.Println()
		fmt.Print("Paste your token: ")

		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		token := strings.TrimSpace(scanner.Text())
		if token == "" {
			fmt.Fprintln(os.Stderr, "No token provided. Exiting.")
			os.Exit(1)
		}

		cfg.Token = token
		if err := config.Save(cfg); err != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to save config: %v\n", err)
		} else {
			fmt.Println("✓ Token saved to ~/.config/ghboard/config.json")
		}
		fmt.Println()
	}

	p := tea.NewProgram(ui.NewApp(cfg), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
