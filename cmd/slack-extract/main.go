package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/chanseok/slackExtract/internal/config"
	"github.com/chanseok/slackExtract/internal/export"
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
	client, err := slack.NewClient(cfg)
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

	// 4. Run TUI
	initialModel := tui.NewModel(channels)
	p := tea.NewProgram(initialModel, tea.WithAltScreen())
	finalModel, err := p.Run()
	if err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}

	// 5. Export Logic
	m := finalModel.(tui.Model)
	if len(m.Selected) == 0 {
		fmt.Println("No channels selected.")
		return
	}

	fmt.Printf("\nStarting export for %d channels...\n", len(m.Selected))

	// Fetch users for mapping (with caching)
	userMap, err := slack.FetchUsers(client)
	if err != nil {
		fmt.Printf("Warning: Could not fetch users: %v\n", err)
		userMap = make(map[string]string)
	}

	for channelID := range m.Selected {
		// Find channel name for display
		channelName := channelID
		for _, ch := range m.Channels {
			if ch.ID == channelID {
				channelName = ch.Name
				break
			}
		}

		fmt.Printf("Processing #%s (%s)...\n", channelName, channelID)
		msgs, err := slack.FetchHistory(client, channelID)
		if err != nil {
			fmt.Printf("  Error fetching history: %v\n", err)
			continue
		}
		fmt.Printf("  -> Successfully fetched %d messages.\n", len(msgs))

		// Save to file
		err = export.SaveToMarkdown(channelName, msgs, userMap)
		if err != nil {
			fmt.Printf("  Error saving file: %v\n", err)
		} else {
			fmt.Printf("  -> Saved to export/%s.md\n", channelName)
		}
	}
}
