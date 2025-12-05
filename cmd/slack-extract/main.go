package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/chanseok/slackExtract/internal/config"
	"github.com/chanseok/slackExtract/internal/slack"
	"github.com/chanseok/slackExtract/internal/tui"
)

func main() {
	// 1. Load Config
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		fmt.Println("Please check your .env file.")
		os.Exit(1)
	}

	fmt.Println("Slack Extract - Initializing...")

	// 2. Initialize Slack Client
	client, httpClient, err := slack.NewClient(cfg)
	if err != nil {
		fmt.Printf("Error initializing Slack client: %v\n", err)
		os.Exit(1)
	}

	// 3. Fetch Channels (with caching)
	channels, err := slack.FetchChannels(client)
	if err != nil {
		fmt.Printf("Error fetching channels: %v\n", err)
		os.Exit(1)
	}

	// 4. Fetch Users (with caching)
	userMap, err := slack.FetchUsers(client)
	if err != nil {
		fmt.Printf("Warning: Could not fetch users: %v\n", err)
		userMap = make(map[string]string)
	}

	// 5. Run TUI
	initialModel := tui.NewModel(channels, client, httpClient, userMap, cfg)
	p := tea.NewProgram(initialModel, tea.WithAltScreen())
	_, err = p.Run()
	if err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
