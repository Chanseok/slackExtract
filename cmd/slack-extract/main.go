package main

import (
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"sort"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/joho/godotenv"
	"github.com/slack-go/slack"
)

// --- Bubble Tea Model & Messages ---

type model struct {
	channels []slack.Channel
	cursor   int
	selected map[string]struct{} // Set of selected channel IDs
	err      error
	quitting bool
}

func initialModel(client *slack.Client) (model, error) {
	// Fetch channels (public and private)
	// types: public_channel, private_channel, mpim, im
	params := &slack.GetConversationsParameters{
		Types: []string{"public_channel", "private_channel", "mpim", "im"},
		Limit: 1000, // Fetch up to 1000 channels for now
	}

	channels, _, err := client.GetConversations(params)
	if err != nil {
		return model{}, err
	}

	// Sort channels by name
	sort.Slice(channels, func(i, j int) bool {
		return channels[i].Name < channels[j].Name
	})

	return model{
		channels: channels,
		selected: make(map[string]struct{}),
	}, nil
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.channels)-1 {
				m.cursor++
			}
		case " ":
			if len(m.channels) > 0 {
				id := m.channels[m.cursor].ID
				if _, ok := m.selected[id]; ok {
					delete(m.selected, id)
				} else {
					m.selected[id] = struct{}{}
				}
			}
		case "enter":
			// TODO: Start export process
			m.quitting = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m model) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v\n", m.err)
	}
	if m.quitting {
		return "Bye!\n"
	}

	s := "Select channels to export (Press Space to select, Enter to confirm):\n\n"

	for i, ch := range m.channels {
		cursor := " " // no cursor
		if m.cursor == i {
			cursor = ">" // cursor!
		}

		checked := " " // not selected
		if _, ok := m.selected[ch.ID]; ok {
			checked = "x" // selected!
		}

		name := ch.Name
		if name == "" {
			name = "DM/Group" // Fallback for unnamed channels
		}

		s += fmt.Sprintf("%s [%s] #%s\n", cursor, checked, name)
	}

	s += "\nPress q to quit.\n"
	return s
}

func main() {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Warning: .env file not found")
	}

	token := os.Getenv("SLACK_USER_TOKEN")
	dCookie := os.Getenv("SLACK_DS_COOKIE")

	if token == "" || dCookie == "" {
		fmt.Println("Error: SLACK_USER_TOKEN (xoxc-...) and SLACK_DS_COOKIE (xoxd-...) are required.")
		fmt.Println("Please check your .env file.")
		os.Exit(1)
	}

	fmt.Println("Slack Extract - Initializing...")

	// Create a custom HTTP client with the cookie
	jar, _ := cookiejar.New(nil)
	u, _ := url.Parse("https://slack.com")
	jar.SetCookies(u, []*http.Cookie{
		{
			Name:   "d",
			Value:  dCookie,
			Path:   "/",
			Domain: ".slack.com",
		},
	})

	client := &http.Client{
		Jar: jar,
	}

	// Initialize Slack API with custom client
	api := slack.New(token, slack.OptionHTTPClient(client))

	// Auth Test
	authTest, err := api.AuthTest()
	if err != nil {
		fmt.Printf("Error connecting to Slack: %v\n", err)
		fmt.Println("Tip: Check if your 'd' cookie and 'xoxc' token are valid and not expired.")
		os.Exit(1)
	}
	fmt.Printf("Successfully authenticated as: %s (Team: %s)\n", authTest.User, authTest.Team)

	// Initialize Bubble Tea Model
	m, err := initialModel(api)
	if err != nil {
		fmt.Printf("Error fetching channels: %v\n", err)
		os.Exit(1)
	}

	// Run Bubble Tea Program
	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
