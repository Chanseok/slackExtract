package manager

import (
	"time"
)

// ChannelMeta represents metadata for an exported channel file
type ChannelMeta struct {
	ChannelName     string
	ChannelID       string    // Might be empty if scanned from file system
	FilePath        string    // Relative path from export root
	FileSize        int64
	MessageCount    int       // Estimated or parsed
	LastUpdated     time.Time // File modification time
	LastMessageTime time.Time // Parsed from content
	IsArchived      bool      // True if file is in "archived" folder
}

// ScanResult holds the result of scanning the export directory
type ScanResult struct {
	Channels map[string]ChannelMeta // Key: ChannelName
}
