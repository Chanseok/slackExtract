package export

import (
	"html"
	"regexp"
)

// CleanSlackText converts Slack's mrkdwn format to standard Markdown
// and cleans up the text for better LLM processing.
func CleanSlackText(text string, userMap map[string]string, channelMap map[string]string) string {
	// 1. User mentions: <@U12345> or <@U12345|username>
	reUser := regexp.MustCompile(`<@(U[A-Z0-9]+)(?:\|([^>]+))?>`)
	text = reUser.ReplaceAllStringFunc(text, func(m string) string {
		matches := reUser.FindStringSubmatch(m)
		if len(matches) >= 2 {
			userID := matches[1]
			// If display name is provided in the mention itself
			if len(matches) >= 3 && matches[2] != "" {
				return "@" + matches[2]
			}
			// Look up in userMap
			if name, ok := userMap[userID]; ok && name != "" {
				return "@" + name
			}
			// Fallback
			if len(userID) > 4 {
				return "@Guy" + userID[len(userID)-4:]
			}
		}
		return m
	})

	// 2. Channel links: <#C12345> or <#C12345|general>
	reChannel := regexp.MustCompile(`<#(C[A-Z0-9]+)(?:\|([^>]+))?>`)
	text = reChannel.ReplaceAllStringFunc(text, func(m string) string {
		matches := reChannel.FindStringSubmatch(m)
		if len(matches) >= 2 {
			channelID := matches[1]
			// If display name is provided in the link itself
			if len(matches) >= 3 && matches[2] != "" {
				return "#" + matches[2]
			}
			// Look up in channelMap (if provided)
			if channelMap != nil {
				if name, ok := channelMap[channelID]; ok && name != "" {
					return "#" + name
				}
			}
			// Fallback to ID
			return "#" + channelID
		}
		return m
	})

	// 3. URLs: <http://example.com> or <http://example.com|Example>
	reURL := regexp.MustCompile(`<(https?://[^|>]+)(?:\|([^>]+))?>`)
	text = reURL.ReplaceAllStringFunc(text, func(m string) string {
		matches := reURL.FindStringSubmatch(m)
		if len(matches) >= 2 {
			url := matches[1]
			// If display text is provided
			if len(matches) >= 3 && matches[2] != "" {
				return "[" + matches[2] + "](" + url + ")"
			}
			// Just the URL
			return url
		}
		return m
	})

	// 4. Special links: <!here>, <!channel>, <!everyone>
	reSpecial := regexp.MustCompile(`<!([a-z]+)(?:\|([^>]+))?>`)
	text = reSpecial.ReplaceAllStringFunc(text, func(m string) string {
		matches := reSpecial.FindStringSubmatch(m)
		if len(matches) >= 2 {
			special := matches[1]
			switch special {
			case "here":
				return "@here"
			case "channel":
				return "@channel"
			case "everyone":
				return "@everyone"
			default:
				if len(matches) >= 3 && matches[2] != "" {
					return "@" + matches[2]
				}
				return "@" + special
			}
		}
		return m
	})

	// 5. HTML entity decoding
	text = html.UnescapeString(text)

	// 6. Slack-style formatting to Markdown (basic cases)
	// Bold: *text* stays the same (already Markdown compatible for bold/italic)
	// Strikethrough: ~text~ -> ~~text~~ (Markdown uses double tilde)
	reStrike := regexp.MustCompile(`~([^~]+)~`)
	text = reStrike.ReplaceAllString(text, "~~$1~~")

	// Code blocks are already ```...``` in Slack, which is Markdown compatible
	// Inline code `text` is also already compatible

	return text
}

// IsSystemMessage checks if the message is a system/bot message that should be filtered
// for LLM processing (e.g., join/leave notifications)
func IsSystemMessage(subtype string) bool {
	systemSubtypes := map[string]bool{
		"channel_join":     true,
		"channel_leave":    true,
		"channel_purpose":  true,
		"channel_topic":    true,
		"channel_name":     true,
		"channel_archive":  true,
		"channel_unarchive": true,
		"group_join":       true,
		"group_leave":      true,
		"group_purpose":    true,
		"group_topic":      true,
		"group_name":       true,
		"group_archive":    true,
		"group_unarchive":  true,
		"pinned_item":      true,
		"unpinned_item":    true,
		"ekm_access_denied": true,
		"me_message":       false, // Keep /me messages, they have content
		"bot_message":      false, // Keep bot messages, they might have useful info
	}

	filtered, exists := systemSubtypes[subtype]
	return exists && filtered
}
