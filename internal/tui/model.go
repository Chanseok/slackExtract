package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/slack-go/slack"
)

type Model struct {
	Channels     []slack.Channel
	Cursor       int
	Selected     map[string]struct{} // Set of selected channel IDs
	Err          error
	Quitting     bool
	WindowMin    int
	Height       int
	SearchMode   bool
	SearchQuery  string
	FilteredIdx  []int // Indices of filtered channels

	// Filter Menu
	FilterMenuMode bool
	FilterCursor   int
	ShowPublic     bool
	ShowPrivate    bool
	ShowArchived   bool
	ShowDMs        bool
}

func NewModel(channels []slack.Channel) Model {
	m := Model{
		Channels:     channels,
		Selected:     make(map[string]struct{}),
		ShowPublic:   true,
		ShowPrivate:  true,
		ShowArchived: true,
		ShowDMs:      true,
	}
	m.updateFilter()
	return m
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Height = msg.Height - 5 // Reserve lines for header/footer
		if m.Height < 1 {
			m.Height = 1
		}
		return m, nil

	case tea.KeyMsg:
		// Filter Menu Mode
		if m.FilterMenuMode {
			switch msg.String() {
			case "esc", "enter", "f":
				m.FilterMenuMode = false
				m.updateFilter()
			case "up", "k":
				if m.FilterCursor > 0 {
					m.FilterCursor--
				}
			case "down", "j":
				if m.FilterCursor < 3 { // 4 options: 0-3
					m.FilterCursor++
				}
			case " ":
				// Toggle the selected filter
				switch m.FilterCursor {
				case 0:
					m.ShowPublic = !m.ShowPublic
				case 1:
					m.ShowPrivate = !m.ShowPrivate
				case 2:
					m.ShowArchived = !m.ShowArchived
				case 3:
					m.ShowDMs = !m.ShowDMs
				}
				m.updateFilter()
			}
			return m, nil
		}

		if m.SearchMode {
			switch msg.String() {
			case "esc":
				m.SearchMode = false
				m.SearchQuery = ""
				m.updateFilter()
			case "enter":
				m.SearchMode = false
			case "backspace":
				if len(m.SearchQuery) > 0 {
					m.SearchQuery = m.SearchQuery[:len(m.SearchQuery)-1]
					m.updateFilter()
				}
			default:
				if len(msg.String()) == 1 {
					m.SearchQuery += msg.String()
					m.updateFilter()
				}
			}
			return m, nil
		}

		switch msg.String() {
		case "ctrl+c", "q":
			m.Quitting = true
			return m, tea.Quit
		case "/":
			m.SearchMode = true
			return m, nil
		case "f":
			m.FilterMenuMode = true
			return m, nil
		case "a":
			// Toggle all visible channels
			m.toggleAllVisible()
		case "up", "k":
			if m.Cursor > 0 {
				m.Cursor--
				if m.Cursor < m.WindowMin {
					m.WindowMin = m.Cursor
				}
			}
		case "down", "j":
			maxIdx := len(m.FilteredIdx) - 1
			if m.Cursor < maxIdx {
				m.Cursor++
				if m.Cursor >= m.WindowMin+m.Height {
					m.WindowMin = m.Cursor - m.Height + 1
				}
			}
		case " ":
			if len(m.FilteredIdx) > 0 && m.Cursor < len(m.FilteredIdx) {
				actualIdx := m.FilteredIdx[m.Cursor]
				id := m.Channels[actualIdx].ID
				if _, ok := m.Selected[id]; ok {
					delete(m.Selected, id)
				} else {
					m.Selected[id] = struct{}{}
				}
			}
		case "enter":
			// Start export process
			m.Quitting = true
			return m, tea.Quit
		}
	}
	return m, nil
}

// updateFilter updates the FilteredIdx based on SearchQuery and attribute filters
func (m *Model) updateFilter() {
	m.FilteredIdx = []int{}
	for i, ch := range m.Channels {
		// Apply attribute filters
		if !m.shouldShowChannel(ch) {
			continue
		}
		// Apply search query filter
		if m.SearchQuery != "" && !containsIgnoreCase(ch.Name, m.SearchQuery) {
			continue
		}
		m.FilteredIdx = append(m.FilteredIdx, i)
	}
	// Reset cursor if out of bounds
	if m.Cursor >= len(m.FilteredIdx) {
		m.Cursor = 0
		m.WindowMin = 0
	}
}

// shouldShowChannel checks if a channel should be shown based on filter settings
func (m *Model) shouldShowChannel(ch slack.Channel) bool {
	// 1. Archived Channel Logic
	// If the channel is archived, it is controlled ONLY by ShowArchived.
	// This allows users to see "All Archived Channels" by checking only ShowArchived,
	// regardless of whether the channel was originally Public or Private.
	if ch.IsArchived {
		return m.ShowArchived
	}

	// 2. Active Channel Logic (Non-Archived)
	// Active channels are filtered by their specific types (Public, Private, DM).
	isDM := ch.IsIM || ch.IsMpIM
	isPrivate := ch.IsPrivate && !isDM
	isPublic := !ch.IsPrivate && !isDM

	if isDM {
		return m.ShowDMs
	}
	if isPrivate {
		return m.ShowPrivate
	}
	if isPublic {
		return m.ShowPublic
	}

	return true
}

// toggleAllVisible selects or deselects all visible (filtered) channels
func (m *Model) toggleAllVisible() {
	if len(m.FilteredIdx) == 0 {
		return
	}
	// Check if all visible are selected
	allSelected := true
	for _, idx := range m.FilteredIdx {
		id := m.Channels[idx].ID
		if _, ok := m.Selected[id]; !ok {
			allSelected = false
			break
		}
	}
	// Toggle
	if allSelected {
		// Deselect all
		for _, idx := range m.FilteredIdx {
			delete(m.Selected, m.Channels[idx].ID)
		}
	} else {
		// Select all
		for _, idx := range m.FilteredIdx {
			m.Selected[m.Channels[idx].ID] = struct{}{}
		}
	}
}

// containsIgnoreCase checks if s contains substr (case-insensitive)
func containsIgnoreCase(s, substr string) bool {
	s = toLower(s)
	substr = toLower(substr)
	return contains(s, substr)
}

func toLower(s string) string {
	result := make([]rune, len(s))
	for i, r := range s {
		if r >= 'A' && r <= 'Z' {
			result[i] = r + 32
		} else {
			result[i] = r
		}
	}
	return string(result)
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || indexOfSubstr(s, substr) >= 0)
}

func indexOfSubstr(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

func (m Model) View() string {
	if m.Err != nil {
		return fmt.Sprintf("Error: %v\n", m.Err)
	}
	if m.Quitting {
		return "Starting export...\n"
	}

	var s string

	// Header
	header := HeaderStyle.Render("🚀 Slack Extract - Channel Selector")
	s += header + "\n"

	// Status bar
	totalChannels := len(m.Channels)
	visibleChannels := len(m.FilteredIdx)
	selectedCount := len(m.Selected)
	statusText := fmt.Sprintf(" Channels: %d total | %d visible | %d selected ", totalChannels, visibleChannels, selectedCount)
	s += StatusBarStyle.Render(statusText) + "\n\n"

	// Search bar
	if m.SearchMode {
		s += fmt.Sprintf("🔍 Search: %s█\n\n", m.SearchQuery)
	} else if m.SearchQuery != "" {
		s += fmt.Sprintf("🔍 Filter: '%s' (Press / to edit, Esc to clear)\n\n", m.SearchQuery)
	}

	// Filter Menu (overlay)
	if m.FilterMenuMode {
		s += m.renderFilterMenu()
		return s
	}

	// Channel list
	height := m.Height
	if height == 0 {
		height = 20
	}

	start := m.WindowMin
	end := start + height
	if end > len(m.FilteredIdx) {
		end = len(m.FilteredIdx)
	}

	for i := start; i < end; i++ {
		actualIdx := m.FilteredIdx[i]
		ch := m.Channels[actualIdx]

		cursor := "  "
		if m.Cursor == i {
			cursor = "❯ "
		}

		checked := "[ ]"
		if _, ok := m.Selected[ch.ID]; ok {
			checked = "[✓]"
		}

		icon := GetChannelIcon(ch.IsPrivate, ch.IsIM, ch.IsMpIM, ch.IsArchived)

		name := ch.Name
		if name == "" {
			name = "DM/Group"
		}

		// Member count
		memberInfo := ""
		if ch.NumMembers > 0 {
			memberInfo = fmt.Sprintf(" (%d members)", ch.NumMembers)
		}

		// Apply style
		var line string
		if m.Cursor == i {
			line = SelectedStyle.Render(fmt.Sprintf("%s%s %s %s%s", cursor, checked, icon, name, memberInfo))
		} else {
			style := GetChannelStyle(ch.IsPrivate, ch.IsIM, ch.IsMpIM, ch.IsArchived, false)
			line = fmt.Sprintf("%s%s %s ", cursor, checked, icon) + style.Render(name) + NormalStyle.Render(memberInfo)
		}

		s += line + "\n"
	}

	if len(m.FilteredIdx) == 0 {
		s += "\n  No channels found.\n"
	}

	// Help
	s += "\n" + HelpStyle.Render("↑↓/jk: Navigate | Space: Select | a: Toggle All | /: Search | f: Filter | Enter: Download | q: Quit")

	return s
}

// renderFilterMenu renders the filter settings menu
func (m Model) renderFilterMenu() string {
	var menu string

	menu += FilterTitleStyle.Render("⚙️  Filter Settings") + "\n\n"

	options := []struct {
		label   string
		enabled bool
	}{
		{"🔓 Show Active Public Channels", m.ShowPublic},
		{"🔒 Show Active Private Channels", m.ShowPrivate},
		{"🗄️  Show All Archived Channels", m.ShowArchived},
		{"💬 Show Active DMs / Groups", m.ShowDMs},
	}

	for i, opt := range options {
		cursor := "  "
		if m.FilterCursor == i {
			cursor = "❯ "
		}

		checkbox := "[ ]"
		if opt.enabled {
			checkbox = "[✓]"
		}

		var line string
		if m.FilterCursor == i {
			line = FilterSelectedStyle.Render(fmt.Sprintf("%s%s %s", cursor, checkbox, opt.label))
		} else {
			line = FilterItemStyle.Render(fmt.Sprintf("%s%s %s", cursor, checkbox, opt.label))
		}
		menu += line + "\n"
	}

	menu += "\n" + HelpStyle.Render("↑↓: Navigate | Space: Toggle | Esc/Enter/f: Close")

	return FilterMenuStyle.Render(menu)
}
