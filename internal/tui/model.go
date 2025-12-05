package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/slack-go/slack"
)

type Model struct {
	Channels  []slack.Channel
	Cursor    int
	Selected  map[string]struct{} // Set of selected channel IDs
	Err       error
	Quitting  bool
	WindowMin int
	Height    int
}

func NewModel(channels []slack.Channel) Model {
	return Model{
		Channels: channels,
		Selected: make(map[string]struct{}),
	}
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
		switch msg.String() {
		case "ctrl+c", "q":
			m.Quitting = true
			return m, tea.Quit
		case "up", "k":
			if m.Cursor > 0 {
				m.Cursor--
				if m.Cursor < m.WindowMin {
					m.WindowMin = m.Cursor
				}
			}
		case "down", "j":
			if m.Cursor < len(m.Channels)-1 {
				m.Cursor++
				if m.Cursor >= m.WindowMin+m.Height {
					m.WindowMin = m.Cursor - m.Height + 1
				}
			}
		case " ":
			if len(m.Channels) > 0 {
				id := m.Channels[m.Cursor].ID
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

func (m Model) View() string {
	if m.Err != nil {
		return fmt.Sprintf("Error: %v\n", m.Err)
	}
	if m.Quitting {
		return "Bye!\n"
	}

	s := "Select channels to export (Press Space to select, Enter to confirm):\n\n"

	height := m.Height
	if height == 0 {
		height = 20 // Default height
	}

	start := m.WindowMin
	end := start + height
	if end > len(m.Channels) {
		end = len(m.Channels)
	}

	for i := start; i < end; i++ {
		ch := m.Channels[i]
		cursor := " " // no cursor
		if m.Cursor == i {
			cursor = ">" // cursor!
		}

		checked := " " // not selected
		if _, ok := m.Selected[ch.ID]; ok {
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
