package manager

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

var (
	dateHeaderRegex = regexp.MustCompile(`## 📅 (\d{4}-\d{2}-\d{2})`)
	timeHeaderRegex = regexp.MustCompile(`### .* - (\d{2}:\d{2}:\d{2})`)
)

// ScanExportDir scans the export directory for existing channel files
func ScanExportDir(exportRoot string) (*ScanResult, error) {
	result := &ScanResult{
		Channels: make(map[string]ChannelMeta),
	}

	err := filepath.Walk(exportRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			// Skip .meta and .analysis directories
			if strings.HasPrefix(info.Name(), ".") {
				return filepath.SkipDir
			}
			return nil
		}

		if filepath.Ext(path) != ".md" {
			return nil
		}

		// Parse channel name from filename
		filename := filepath.Base(path)
		channelName := strings.TrimSuffix(filename, ".md")

		// Check if archived
		relPath, _ := filepath.Rel(exportRoot, path)
		isArchived := strings.Contains(filepath.Dir(relPath), "archived")

		// Parse last message time
		lastMsgTime, msgCount, err := parseFileMetadata(path)
		if err != nil {
			// Log error but continue?
			fmt.Printf("Warning: failed to parse metadata for %s: %v\n", path, err)
		}

		meta := ChannelMeta{
			ChannelName:     channelName,
			FilePath:        relPath,
			FileSize:        info.Size(),
			LastUpdated:     info.ModTime(),
			LastMessageTime: lastMsgTime,
			MessageCount:    msgCount, // This is expensive to count exactly, maybe estimate or skip for now
			IsArchived:      isArchived,
		}

		result.Channels[channelName] = meta
		return nil
	})

	return result, err
}

// parseFileMetadata reads the file to find the last message timestamp and estimate message count
func parseFileMetadata(path string) (time.Time, int, error) {
	f, err := os.Open(path)
	if err != nil {
		return time.Time{}, 0, err
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return time.Time{}, 0, err
	}

	// 1. Estimate message count (rough count of "### " lines)
	// For large files, scanning the whole file might be slow.
	// Let's skip exact count for now or do a fast scan if file is small (< 1MB)
	msgCount := 0
	if info.Size() < 1024*1024 { // 1MB
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			if strings.HasPrefix(scanner.Text(), "### ") {
				msgCount++
			}
		}
		// Reset file pointer for timestamp parsing
		f.Seek(0, 0)
	}

	// 2. Find last timestamp
	// Read last 20KB
	const tailSize = 20 * 1024
	offset := int64(0)
	if info.Size() > tailSize {
		offset = info.Size() - tailSize
	}
	
	_, err = f.Seek(offset, 0)
	if err != nil {
		return time.Time{}, msgCount, err
	}

	content, err := io.ReadAll(f)
	if err != nil {
		return time.Time{}, msgCount, err
	}
	
	strContent := string(content)

	// Find last date
	dateMatches := dateHeaderRegex.FindAllStringSubmatch(strContent, -1)
	var lastDateStr string
	var lastDateIdx int
	
	if len(dateMatches) > 0 {
		lastMatch := dateMatches[len(dateMatches)-1]
		lastDateStr = lastMatch[1]
		// Find index of this match in strContent to search for time after it
		lastDateIdx = strings.LastIndex(strContent, lastMatch[0])
	} else {
		// If no date found in tail, maybe the file is small and we read it all?
		// Or the date header is further up.
		// Fallback: try to find date in the whole file if we haven't read it all?
		// For now, return zero time if not found
		return time.Time{}, msgCount, nil
	}

	// Find last time AFTER the date
	timeContent := strContent[lastDateIdx:]
	timeMatches := timeHeaderRegex.FindAllStringSubmatch(timeContent, -1)
	
	if len(timeMatches) > 0 {
		lastTimeStr := timeMatches[len(timeMatches)-1][1]
		fullTimeStr := fmt.Sprintf("%s %s", lastDateStr, lastTimeStr)
		t, err := time.Parse("2006-01-02 15:04:05", fullTimeStr)
		if err == nil {
			return t, msgCount, nil
		}
	}

	return time.Time{}, msgCount, nil
}
