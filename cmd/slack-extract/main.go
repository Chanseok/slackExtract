package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"regexp"
	"sort"
	"strconv"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/joho/godotenv"
	"github.com/slack-go/slack"
)

// --- Bubble Tea Model & Messages ---

type model struct {
	channels  []slack.Channel
	cursor    int
	selected  map[string]struct{} // Set of selected channel IDs
	err       error
	quitting  bool
	windowMin int
	height    int
}

func initialModel(client *slack.Client) (model, error) {
	// Fetch channels (public and private)
	// types: public_channel, private_channel, mpim, im
	params := &slack.GetConversationsParameters{
		Types: []string{"public_channel", "private_channel", "mpim", "im"},
		Limit: 1000, // Fetch up to 1000 channels per page
	}

	var allChannels []slack.Channel
	for {
		channels, nextCursor, err := client.GetConversations(params)
		if err != nil {
			return model{}, err
		}
		allChannels = append(allChannels, channels...)

		if nextCursor == "" {
			break
		}
		params.Cursor = nextCursor
	}

	// Sort channels by name
	sort.Slice(allChannels, func(i, j int) bool {
		return allChannels[i].Name < allChannels[j].Name
	})

	return model{
		channels: allChannels,
		selected: make(map[string]struct{}),
	}, nil
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.height = msg.Height - 5 // Reserve lines for header/footer
		if m.height < 1 {
			m.height = 1
		}
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
				if m.cursor < m.windowMin {
					m.windowMin = m.cursor
				}
			}
		case "down", "j":
			if m.cursor < len(m.channels)-1 {
				m.cursor++
				if m.cursor >= m.windowMin+m.height {
					m.windowMin = m.cursor - m.height + 1
				}
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

	height := m.height
	if height == 0 {
		height = 20 // Default height
	}

	start := m.windowMin
	end := start + height
	if end > len(m.channels) {
		end = len(m.channels)
	}

	for i := start; i < end; i++ {
		ch := m.channels[i]
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
	p := tea.NewProgram(m, tea.WithAltScreen())
	finalModel, err := p.Run()
	if err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}

	// --- Export Logic ---
	m = finalModel.(model)
	if len(m.selected) == 0 {
		fmt.Println("No channels selected.")
		return
	}

	fmt.Printf("\nStarting export for %d channels...\n", len(m.selected))

	// Fetch users for mapping
	fmt.Println("Fetching user list for mapping...")
	userMap, err := fetchUsers(api)
	if err != nil {
		fmt.Printf("Warning: Could not fetch users: %v\n", err)
		userMap = make(map[string]string)
	} else {
		fmt.Printf("  -> Fetched %d users.\n", len(userMap))
	}

	for channelID := range m.selected {
		// Find channel name for display
		channelName := channelID
		for _, ch := range m.channels {
			if ch.ID == channelID {
				channelName = ch.Name
				break
			}
		}

		fmt.Printf("Processing #%s (%s)...\n", channelName, channelID)
		msgs, err := fetchHistory(api, channelID)
		if err != nil {
			fmt.Printf("  Error fetching history: %v\n", err)
			continue
		}
		fmt.Printf("  -> Successfully fetched %d messages.\n", len(msgs))

		// Save to file
		err = saveToFile(channelName, msgs, userMap)
		if err != nil {
			fmt.Printf("  Error saving file: %v\n", err)
		} else {
			fmt.Printf("  -> Saved to export/%s.md\n", channelName)
		}
	}
}

func fetchUsers(client *slack.Client) (map[string]string, error) {
	cacheFile := "users.json"
	userMap := make(map[string]string)

	// 1. Try to load from cache
	if _, err := os.Stat(cacheFile); err == nil {
		data, err := os.ReadFile(cacheFile)
		if err == nil {
			if err := json.Unmarshal(data, &userMap); err == nil {
				fmt.Println("Loaded user list from cache (users.json).")
				return userMap, nil
			}
		}
	}

	// 2. Fetch from API (with pagination)
	fmt.Println("Fetching user list from Slack API...")
	var allUsers []slack.User
	limit := 1000
	cursor := ""

	for {
		users, nextCursor, err := client.GetUsersPaginated(slack.GetUsersOptionLimit(limit), slack.GetUsersOptionCursor(cursor))
		if err != nil {
			return nil, err
		}
		allUsers = append(allUsers, users...)

		if nextCursor == "" {
			break
		}
		cursor = nextCursor
		fmt.Printf("  ...fetched %d users so far\n", len(allUsers))
	}

	for _, u := range allUsers {
		userMap[u.ID] = u.RealName
	}

	// 3. Save to cache
	data, err := json.MarshalIndent(userMap, "", "  ")
	if err == nil {
		_ = os.WriteFile(cacheFile, data, 0644)
		fmt.Println("Saved user list to cache (users.json).")
	}

	return userMap, nil
}

func saveToFile(channelName string, msgs []slack.Message, userMap map[string]string) error {
	// Create export directory if not exists
	if err := os.MkdirAll("export", 0755); err != nil {
		return err
	}

	f, err := os.Create(fmt.Sprintf("export/%s.md", channelName))
	if err != nil {
		return err
	}
	defer f.Close()

	// Reverse messages to show oldest first
	for i := len(msgs) - 1; i >= 0; i-- {
		msg := msgs[i]
		
		// 1. Format Time
		floatTs, _ := strconv.ParseFloat(msg.Timestamp, 64)
		ts := time.Unix(int64(floatTs), 0)
		timeStr := ts.Format("2006-01-02 15:04:05")

		// 2. Resolve User Name
		userName := userMap[msg.User]
		if userName == "" {
			userName = msg.User // Fallback to ID
			if msg.BotID != "" {
				userName = fmt.Sprintf("%s (Bot)", msg.Username)
			}
		}

		// 3. Replace Mentions in Text
		text := resolveMentions(msg.Text, userMap)

		// Simple Markdown format
		_, err := fmt.Fprintf(f, "### %s (%s)\n%s\n\n---\n\n", userName, timeStr, text)
		if err != nil {
			return err
		}
	}
	return nil
}

func resolveMentions(text string, userMap map[string]string) string {
	// Regex to find <@U...>
	re := regexp.MustCompile(`<@(U[A-Z0-9]+)>`)
	return re.ReplaceAllStringFunc(text, func(match string) string {
		id := match[2 : len(match)-1] // remove <@ and >
		if name, ok := userMap[id]; ok {
			return "@" + name
		}
		return match
	})
}

func fetchHistory(client *slack.Client, channelID string) ([]slack.Message, error) {
	var allMessages []slack.Message
	params := &slack.GetConversationHistoryParameters{
		ChannelID: channelID,
		Limit:     1000,
	}

	for {
		history, err := client.GetConversationHistory(params)
		if err != nil {
			return nil, err
		}
		allMessages = append(allMessages, history.Messages...)

		if !history.HasMore {
			break
		}
		params.Cursor = history.ResponseMetaData.NextCursor
	}
	return allMessages, nil
}
