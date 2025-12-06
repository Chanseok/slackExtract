package export

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/chanseok/slackExtract/internal/slack"
	slackgo "github.com/slack-go/slack"
)

// SaveToMarkdown saves messages to a Markdown file
func SaveToMarkdown(httpClient *http.Client, channelName string, msgs []slack.Message, userMap map[string]string, downloadAttachments bool, targetFolder string, appendMode bool) error {
	// Create target folder if it doesn't exist
	if err := os.MkdirAll(targetFolder, 0755); err != nil {
		return fmt.Errorf("failed to create target folder: %w", err)
	}

	// Sanitize channel name for filename
	safeName := sanitizeFilename(channelName)
	filePath := filepath.Join(targetFolder, safeName+".md")

	// Check if file exists when in append mode
	if appendMode {
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			// File doesn't exist, switch to normal mode
			appendMode = false
		}
	}

	// Open or create file
	var file *os.File
	var err error
	
	if appendMode {
		file, err = os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("failed to open file for append: %w", err)
		}
	} else {
		file, err = os.Create(filePath)
		if err != nil {
			return fmt.Errorf("failed to create file: %w", err)
		}
		
		// Write header only for new files
		fmt.Fprintf(file, "# %s\n\n", channelName)
		fmt.Fprintf(file, "Exported: %s\n\n", time.Now().Format("2006-01-02 15:04:05"))
		fmt.Fprintf(file, "---\n\n")
	}
	defer file.Close()

	// Sort messages from oldest to newest
	sort.Slice(msgs, func(i, j int) bool {
		return msgs[i].Timestamp < msgs[j].Timestamp
	})

	// Write messages
	for _, msg := range msgs {
		if err := writeMessage(file, msg, userMap, httpClient, channelName, downloadAttachments, targetFolder, 0); err != nil {
			return fmt.Errorf("failed to write message: %w", err)
		}

		// Write thread replies if any
		if len(msg.Replies) > 0 {
			for _, reply := range msg.Replies {
				replyMsg := slack.Message{Message: reply}
				if err := writeMessage(file, replyMsg, userMap, httpClient, channelName, downloadAttachments, targetFolder, 1); err != nil {
					return fmt.Errorf("failed to write reply: %w", err)
				}
			}
		}
	}

	return nil
}

func writeMessage(file *os.File, msg slack.Message, userMap map[string]string, httpClient *http.Client, channelName string, downloadAttachments bool, targetFolder string, indentLevel int) error {
	// Get user name
	userName := getUserName(msg.Message, userMap)

	// Parse timestamp
	msgTime, err := slack.ParseTimestamp(msg.Timestamp)
	if err != nil {
		msgTime = time.Now()
	}

	// Format time
	timeStr := msgTime.Format("2006-01-02 15:04:05")

	// Clean text
	text := cleanSlackText(msg.Text, userMap)

	// Handle indentation for thread replies
	indent := ""
	if indentLevel > 0 {
		indent = "> "
	}

	// Write message header
	if indentLevel == 0 {
		fmt.Fprintf(file, "### %s - %s\n\n", userName, timeStr)
	} else {
		fmt.Fprintf(file, "%s**%s** - %s\n%s\n", indent, userName, timeStr, indent)
	}

	// Write message text
	if text != "" {
		lines := strings.Split(text, "\n")
		for _, line := range lines {
			fmt.Fprintf(file, "%s%s\n", indent, line)
		}
		fmt.Fprintln(file)
	}

	// Handle file attachments
	if len(msg.Files) > 0 {
		for _, f := range msg.Files {
			if downloadAttachments {
				// Download file
				localPath, err := downloadFile(httpClient, f, channelName, targetFolder)
				if err != nil {
					fmt.Fprintf(file, "%s📎 [%s](%s) *(download failed)*\n", indent, f.Name, f.URLPrivate)
				} else {
					// Check if it's an image
					if isImage(f.Mimetype) {
						fmt.Fprintf(file, "%s![%s](%s)\n", indent, f.Name, localPath)
					} else {
						fmt.Fprintf(file, "%s📎 [%s](%s)\n", indent, f.Name, localPath)
					}
				}
			} else {
				// Just link to URL
				fmt.Fprintf(file, "%s📎 [%s](%s)\n", indent, f.Name, f.URLPrivate)
			}
		}
		fmt.Fprintln(file)
	}

	// Add separator for top-level messages
	if indentLevel == 0 {
		fmt.Fprintln(file, "---\n")
	}

	return nil
}

func getUserName(msg slackgo.Message, userMap map[string]string) string {
	// Try to get from userMap
	if name, ok := userMap[msg.User]; ok && name != "" {
		return name
	}

	// Try bot username
	if msg.BotID != "" && msg.Username != "" {
		return msg.Username
	}

	// Fallback to "Guy XXXX" format
	if msg.User != "" {
		if len(msg.User) >= 4 {
			return "Guy " + msg.User[len(msg.User)-4:]
		}
		return "Guy " + msg.User
	}

	return "Unknown"
}

func cleanSlackText(text string, userMap map[string]string) string {
	// Replace user mentions <@U123> with real names
	re := regexp.MustCompile(`<@([A-Z0-9]+)>`)
	text = re.ReplaceAllStringFunc(text, func(match string) string {
		userID := match[2 : len(match)-1]
		if name, ok := userMap[userID]; ok && name != "" {
			return "@" + name
		}
		if len(userID) >= 4 {
			return "@Guy " + userID[len(userID)-4:]
		}
		return "@" + userID
	})

	// Replace channel links <#C123|general> with #general
	re = regexp.MustCompile(`<#[A-Z0-9]+\|([^>]+)>`)
	text = re.ReplaceAllString(text, "#$1")

	// Replace URLs <url|text> with [text](url)
	re = regexp.MustCompile(`<(https?://[^|>]+)\|([^>]+)>`)
	text = re.ReplaceAllString(text, "[$2]($1)")

	// Replace bare URLs <url> with [url](url)
	re = regexp.MustCompile(`<(https?://[^>]+)>`)
	text = re.ReplaceAllString(text, "[$1]($1)")

	// Decode HTML entities
	text = strings.ReplaceAll(text, "&lt;", "<")
	text = strings.ReplaceAll(text, "&gt;", ">")
	text = strings.ReplaceAll(text, "&amp;", "&")

	return text
}

func downloadFile(httpClient *http.Client, file slackgo.File, channelName string, targetFolder string) (string, error) {
	// Create attachments directory
	attachDir := filepath.Join(targetFolder, "attachments", channelName)
	if err := os.MkdirAll(attachDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create attachments directory: %w", err)
	}

	// Create unique filename
	filename := fmt.Sprintf("%s_%s", file.ID, sanitizeFilename(file.Name))
	filePath := filepath.Join(attachDir, filename)

	// Check if file already exists
	if _, err := os.Stat(filePath); err == nil {
		// File already exists, return relative path
		relPath, _ := filepath.Rel(targetFolder, filePath)
		return relPath, nil
	}

	// Download file
	url := file.URLPrivateDownload
	if url == "" {
		url = file.URLPrivate
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("bad status: %s", resp.Status)
	}

	// Create file
	out, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer out.Close()

	// Write content
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return "", err
	}

	// Return relative path
	relPath, _ := filepath.Rel(targetFolder, filePath)
	return relPath, nil
}

func isImage(mimetype string) bool {
	return strings.HasPrefix(mimetype, "image/")
}

func sanitizeFilename(name string) string {
	// Replace filesystem-unsafe characters
	replacements := map[string]string{
		"/":  "_",
		"\\": "_",
		":":  "_",
		"*":  "_",
		"?":  "_",
		"\"": "_",
		"<":  "_",
		">":  "_",
		"|":  "_",
	}
	
	result := name
	for old, new := range replacements {
		result = strings.ReplaceAll(result, old, new)
	}
	
	return result
}
