package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/chanseok/slackExtract/internal/manager"
)

func (m Model) renderConfirmView() string {
	var s string

	s += HeaderStyle.Render("📥 Confirm Download") + "\n\n"

	// 1. Target Folder Selection
	s += "📁 Target Folder:\n"
	
	// List subfolders
	for i, folder := range m.SubFolders {
		cursor := "  "
		if m.FolderCursor == i {
			cursor = "❯ "
		}
		
		style := NormalStyle
		if m.FolderCursor == i {
			style = SelectedStyle
		}
		
		label := folder
		if folder == "." {
			label = "./ (Root)"
		} else if folder == "+ New Folder" {
			label = "+ New Folder..."
		}
		
		s += fmt.Sprintf("%s%s\n", cursor, style.Render(label))
	}
	s += "\n"

	// 2. Existing Files Warning
	existingCount := 0
	for id := range m.Selected {
		// Find channel name
		var chName string
		for _, ch := range m.Channels {
			if ch.ID == id {
				chName = ch.Name
				break
			}
		}
		if _, ok := m.ExistingFiles[chName]; ok {
			existingCount++
		}
	}

	if existingCount > 0 {
		s += fmt.Sprintf("⚠️  %d of %d selected channels already exist in export/:\n", existingCount, len(m.Selected))
		
		// Show up to 5 existing files
		shown := 0
		for id := range m.Selected {
			var chName string
			for _, ch := range m.Channels {
				if ch.ID == id {
					chName = ch.Name
					break
				}
			}
			
			if meta, ok := m.ExistingFiles[chName]; ok {
				shown++
				if shown > 5 {
					s += "  ... and more\n"
					break
				}
				
				archivedTag := ""
				if meta.IsArchived {
					archivedTag = " [ARCHIVED]"
				}
				
				lastUpdate := meta.LastUpdated.Format("Jan 02")
				sizeStr := fmt.Sprintf("%.1f KB", float64(meta.FileSize)/1024.0)
				
				s += fmt.Sprintf("  • %s (%s, %s)%s\n", chName, sizeStr, lastUpdate, archivedTag)
			}
		}
		s += "\n"
	} else {
		s += fmt.Sprintf("✅ %d channels selected. No conflicts found.\n\n", len(m.Selected))
	}

	// 3. Action Selection
	s += "Choose Action:\n"
	
	// Adjust cursor base for actions (it starts after folders)
	// But wait, we can use a separate cursor or just one cursor for the whole screen?
	// Let's use a separate ActionCursor and switch focus? 
	// Or just simple: Up/Down navigates folders, Tab switches to Action?
	// For simplicity, let's make it a single list or use specific keys for actions.
	// The design said: "[s] Skip, [i] Incremental, [o] Overwrite, [c] Cancel"
	
	// Let's stick to the design: keys for actions.
	// But we need to select folder first.
	
	s += "  [s] Skip existing\n"
	s += "  [i] Incremental\n"
	s += "  [o] Overwrite all\n"
	s += "  [c] Cancel\n"
	
	s += "\n" + HelpStyle.Render("↑↓: Select Folder | s/i/o: Start Download | c: Cancel")
	
	return s
}

func (m *Model) scanForExistingFiles() {
	// 1. Scan export directory
	scanResult, err := manager.ScanExportDir("export")
	if err != nil {
		// If export dir doesn't exist, that's fine
		m.ExistingFiles = make(map[string]manager.ChannelMeta)
	} else {
		m.ExistingFiles = scanResult.Channels
	}

	// 2. Populate SubFolders
	m.SubFolders = []string{".", "archived", "sales", "project", "uncategorized"} // Default suggestions
	
	// Add actual subfolders found
	entries, err := os.ReadDir("export")
	if err == nil {
		for _, entry := range entries {
			if entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
				// Check if already in list
				found := false
				for _, f := range m.SubFolders {
					if f == entry.Name() {
						found = true
						break
					}
				}
				if !found {
					m.SubFolders = append(m.SubFolders, entry.Name())
				}
			}
		}
	}
	m.SubFolders = append(m.SubFolders, "+ New Folder")
}

func (m Model) updateConfirm(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.FolderCursor > 0 {
				m.FolderCursor--
			}
		case "down", "j":
			if m.FolderCursor < len(m.SubFolders)-1 {
				m.FolderCursor++
			}
		case "c", "esc":
			m.ConfirmMode = false
			return m, nil
		case "s":
			m.DownloadAction = "skip"
			return m.startDownloadSequence()
		case "i":
			m.DownloadAction = "incremental"
			return m.startDownloadSequence()
		case "o":
			m.DownloadAction = "overwrite"
			return m.startDownloadSequence()
		}
	}
	return m, nil
}

func (m Model) startDownloadSequence() (Model, tea.Cmd) {
	m.ConfirmMode = false
	m.IsDownloading = true
	m.StartTime = time.Now()
	m.TotalSelected = len(m.Selected)
	m.ProgressChannel = make(chan ProgressMsg)
	
	// Set target folder based on selection
	selectedFolder := m.SubFolders[m.FolderCursor]
	if selectedFolder == "." {
		m.TargetFolder = "export"
	} else if selectedFolder == "+ New Folder" {
		// TODO: Implement input for new folder
		// For now fallback to export/new
		m.TargetFolder = filepath.Join("export", "new")
	} else {
		m.TargetFolder = filepath.Join("export", selectedFolder)
	}

	return m, tea.Batch(
		startDownload(m),
		waitForUpdate(m.ProgressChannel),
	)
}
