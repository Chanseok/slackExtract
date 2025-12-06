package meta

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const (
	MetaDirName   = ".meta"
	IndexFileName = "index.json"
)

// Manager handles metadata operations
type Manager struct {
	baseDir string
	mu      sync.RWMutex
	index   *Index
}

// NewManager creates a new metadata manager
func NewManager(baseDir string) (*Manager, error) {
	m := &Manager{
		baseDir: baseDir,
	}
	if err := m.loadIndex(); err != nil {
		// If index doesn't exist, create a new one
		m.index = NewIndex()
	}
	return m, nil
}

// loadIndex loads the index from disk
func (m *Manager) loadIndex() error {
	indexPath := filepath.Join(m.baseDir, MetaDirName, IndexFileName)
	data, err := os.ReadFile(indexPath)
	if err != nil {
		return err
	}

	var index Index
	if err := json.Unmarshal(data, &index); err != nil {
		return err
	}
	m.index = &index
	return nil
}

// SaveIndex saves the current index to disk
func (m *Manager) SaveIndex() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.index.LastUpdated = time.Now()

	metaDir := filepath.Join(m.baseDir, MetaDirName)
	if err := os.MkdirAll(metaDir, 0755); err != nil {
		return fmt.Errorf("failed to create meta directory: %w", err)
	}

	indexPath := filepath.Join(metaDir, IndexFileName)
	data, err := json.MarshalIndent(m.index, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal index: %w", err)
	}

	return os.WriteFile(indexPath, data, 0644)
}

// UpdateChannelDownload updates metadata after a channel download
func (m *Manager) UpdateChannelDownload(channelID, channelName, relPath string, msgCount int, lastMsgAt time.Time) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.index.Channels == nil {
		m.index.Channels = make(map[string]*Channel)
	}

	ch, exists := m.index.Channels[channelID]
	if !exists {
		ch = &Channel{
			ID:   channelID,
			Name: channelName,
		}
		m.index.Channels[channelID] = ch
	}

	ch.Path = relPath
	ch.MessageCount = msgCount
	ch.LastMessageAt = lastMsgAt
	ch.LastDownloadedAt = time.Now()
}

// EnsureChannel ensures a channel exists in the index
func (m *Manager) EnsureChannel(id, name string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.index.Channels == nil {
		m.index.Channels = make(map[string]*Channel)
	}

	if _, exists := m.index.Channels[id]; !exists {
		m.index.Channels[id] = &Channel{
			ID:   id,
			Name: name,
		}
	}
}

// UpdateChannelAnalysis updates metadata after an analysis
func (m *Manager) UpdateChannelAnalysis(channelID string, meta *AnalysisMeta) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	ch, exists := m.index.Channels[channelID]
	if !exists {
		return fmt.Errorf("channel %s not found in index", channelID)
	}

	ch.Analysis = meta
	return nil
}

// GetChannel returns metadata for a specific channel
func (m *Manager) GetChannel(channelID string) (*Channel, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	ch, exists := m.index.Channels[channelID]
	return ch, exists
}

// GetChannelByName returns metadata for a specific channel by name
func (m *Manager) GetChannelByName(name string) (*Channel, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, ch := range m.index.Channels {
		if ch.Name == name {
			return ch, true
		}
	}
	return nil, false
}

