package tui

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/chanseok/slackExtract/internal/config"
	"github.com/chanseok/slackExtract/internal/manager"
	"github.com/chanseok/slackExtract/internal/meta"
	"github.com/slack-go/slack"
)

// ProgressMsg is sent by the worker to update the UI
type ProgressMsg struct {
	ChannelName string
	Current     int    // Number of items processed (messages fetched, files downloaded)
	Total       int    // Total items to process (if known)
	Status      string // Description of current action
	Done        bool   // True if this channel is finished
	AllDone     bool   // True if all selected channels are finished
	Err         error
}

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

	// Grouping Mode
	GroupingMode bool
	GroupName    string

	// Filter Menu
	FilterMenuMode bool
	FilterCursor   int
	ShowPublic     bool
	ShowPrivate    bool
	ShowArchived   bool
	ShowDMs        bool

	// Confirm Download Mode
	ConfirmMode    bool
	TargetFolder   string
	SubFolders     []string
	FolderCursor   int
	ExistingFiles  map[string]manager.ChannelMeta
	DownloadAction string // "skip", "incremental", "overwrite"
	ActionCursor   int    // 0: Skip, 1: Incremental, 2: Overwrite, 3: Cancel

	// Metadata Manager
	MetaManager    *meta.Manager

	// Progress / Download State
	SlackClient      *slack.Client
	HTTPClient       *http.Client
	UserMap          map[string]string
	Config           *config.Config
	IsDownloading    bool
	ProgressChannel  chan ProgressMsg // Channel to receive updates from worker
	CurrentChannel   string
	ProgressCurrent  int
	ProgressTotal    int
	ProgressStatus   string
	StartTime        time.Time
	FinishedChannels int
	TotalSelected    int
}

func NewModel(channels []slack.Channel, client *slack.Client, httpClient *http.Client, userMap map[string]string, cfg *config.Config, metaManager *meta.Manager) Model {
	m := Model{
		Channels:       channels,
		Selected:       make(map[string]struct{}),
		ShowPublic:     true,
		ShowPrivate:    true,
		ShowArchived:   true,
		ShowDMs:        true,
		SlackClient:    client,
		HTTPClient:     httpClient,
		UserMap:        userMap,
		Config:         cfg,
		TargetFolder:   "export",
		DownloadAction: "skip",
		MetaManager:    metaManager,
	}
	m.updateFilter()
	return m
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Handle Confirm Mode
	if m.ConfirmMode {
		newModel, cmd := m.updateConfirm(msg)
		return newModel, cmd
	}

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

		if m.GroupingMode {
			switch msg.String() {
			case "esc":
				m.GroupingMode = false
				m.GroupName = ""
			case "enter":
				m.GroupingMode = false
				// Proceed to confirm mode with the group name
				m.ConfirmMode = true
				m.scanForExistingFiles()
			case "backspace":
				if len(m.GroupName) > 0 {
					m.GroupName = m.GroupName[:len(m.GroupName)-1]
				}
			default:
				if len(msg.String()) == 1 {
					m.GroupName += msg.String()
				}
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
		case "g":
			if len(m.Selected) > 0 {
				m.GroupingMode = true
				m.GroupName = ""
				return m, nil
			}
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
			if len(m.Selected) > 0 && !m.IsDownloading {
				// Switch to Confirm Mode instead of starting download immediately
				m.ConfirmMode = true
				m.scanForExistingFiles()
				return m, nil
			}
		}

	case ProgressMsg:
		if msg.AllDone {
			m.Quitting = true
			return m, tea.Quit
		}
		m.CurrentChannel = msg.ChannelName
		m.ProgressCurrent = msg.Current
		m.ProgressTotal = msg.Total
		m.ProgressStatus = msg.Status
		if msg.Done {
			m.FinishedChannels++
		}
		return m, waitForUpdate(m.ProgressChannel)
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
	if m.IsDownloading {
		return m.renderProgress()
	}
	if m.ConfirmMode {
		return m.renderConfirmView()
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

	// Helper for status bar items
	item := func(key, val string) string {
		return StatusBarKeyStyle.Render(key) + StatusBarValueStyle.Render(val)
	}

	bar := lipgloss.JoinHorizontal(lipgloss.Top,
		item(" Total: ", fmt.Sprintf("%d ", totalChannels)),
		item(" Visible: ", fmt.Sprintf("%d ", visibleChannels)),
		item(" Selected: ", fmt.Sprintf("%d ", selectedCount)),
	)
	s += StatusBarStyle.Render(bar) + "\n\n"

	// Search bar
	if m.SearchMode {
		s += fmt.Sprintf("🔍 Search: %s█\n\n", m.SearchQuery)
	} else if m.SearchQuery != "" {
		s += fmt.Sprintf("🔍 Filter: '%s' (Press / to edit, Esc to clear)\n\n", m.SearchQuery)
	}

	// Grouping Input (overlay)
	if m.GroupingMode {
		return fmt.Sprintf("\n\n  📂 Enter Group Name (Folder): %s█\n\n  (Press Enter to confirm, Esc to cancel)", m.GroupName)
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

		// Metadata Status
		metaStatus := ""
		if m.MetaManager != nil {
			if chMeta, ok := m.MetaManager.GetChannel(ch.ID); ok {
				if !chMeta.LastDownloadedAt.IsZero() {
					metaStatus += " ⬇️ "
				}
				if chMeta.Analysis != nil && !chMeta.Analysis.LastAnalyzedAt.IsZero() {
					metaStatus += " 📝"
				}
			}
		}

		// Apply style
		var line string
		if m.Cursor == i {
			line = SelectedStyle.Render(fmt.Sprintf("%s%s %s %s%s%s", cursor, checked, icon, name, memberInfo, metaStatus))
		} else {
			style := GetChannelStyle(ch.IsPrivate, ch.IsIM, ch.IsMpIM, ch.IsArchived, false)
			line = fmt.Sprintf("%s%s %s ", cursor, checked, icon) + style.Render(name) + NormalStyle.Render(memberInfo+metaStatus)
		}

		s += line + "\n"
	}

	if len(m.FilteredIdx) == 0 {
		s += "\n  No channels found.\n"
	}

	// Help
	helpText := "↑↓/jk: Move • Space: Select • a: All • /: Search • f: Filter • g: Group • Enter: Download • q: Quit"
	s += "\n" + HelpStyle.Render(helpText)

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

func (m Model) renderProgress() string {
	// Calculate ETA
	elapsed := time.Since(m.StartTime)
	var eta string
	if m.FinishedChannels > 0 {
		avgTimePerChannel := elapsed / time.Duration(m.FinishedChannels)
		remaining := time.Duration(m.TotalSelected-m.FinishedChannels) * avgTimePerChannel
		eta = fmt.Sprintf("ETA: %s", remaining.Round(time.Second))
	} else {
		eta = "ETA: Calculating..."
	}

	// Progress Bar for Channels
	percent := 0.0
	if m.TotalSelected > 0 {
		percent = float64(m.FinishedChannels) / float64(m.TotalSelected)
	}
	width := 50
	filled := int(percent * float64(width))
	bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)

	s := fmt.Sprintf("\n  🚀 Exporting Channels... (%d/%d)\n\n", m.FinishedChannels, m.TotalSelected)
	s += fmt.Sprintf("  [%s] %.0f%%\n", bar, percent*100)
	s += fmt.Sprintf("  %s\n\n", eta)

	s += fmt.Sprintf("  Current: %s\n", m.CurrentChannel)
	s += fmt.Sprintf("  Status:  %s\n", m.ProgressStatus)
	if m.ProgressTotal > 0 {
		s += fmt.Sprintf("  Items:   %d / %d\n", m.ProgressCurrent, m.ProgressTotal)
	} else {
		s += fmt.Sprintf("  Items:   %d\n", m.ProgressCurrent)
	}

	return s
}
