package slack

import (
	"fmt"
	"strings"
	"time"

	"github.com/slack-go/slack"
)

// RetryConfig holds configuration for retry behavior
type RetryConfig struct {
	MaxRetries     int           // Maximum number of retries
	InitialBackoff time.Duration // Initial backoff duration
	MaxBackoff     time.Duration // Maximum backoff duration
}

// DefaultRetryConfig returns sensible defaults for Slack API
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries:     5,
		InitialBackoff: 1 * time.Second,
		MaxBackoff:     60 * time.Second,
	}
}

// isRateLimitError checks if the error is a rate limit error and extracts retry delay
func isRateLimitError(err error) (bool, time.Duration) {
	if err == nil {
		return false, 0
	}

	// slack-go returns RateLimitedError for 429 responses
	if rateLimitErr, ok := err.(*slack.RateLimitedError); ok {
		return true, rateLimitErr.RetryAfter
	}

	// Also check for string matching as a fallback
	errStr := err.Error()
	if strings.Contains(errStr, "rate_limited") || strings.Contains(errStr, "429") {
		return true, 30 * time.Second // Default wait if we can't parse
	}

	return false, 0
}

// withRetry executes a function with retry logic for rate limits
func withRetry[T any](cfg RetryConfig, operation string, fn func() (T, error)) (T, error) {
	var result T
	var lastErr error
	backoff := cfg.InitialBackoff

	for attempt := 0; attempt <= cfg.MaxRetries; attempt++ {
		result, lastErr = fn()
		if lastErr == nil {
			return result, nil
		}

		isRateLimit, retryAfter := isRateLimitError(lastErr)
		if !isRateLimit {
			// Not a rate limit error, return immediately
			return result, lastErr
		}

		if attempt == cfg.MaxRetries {
			return result, fmt.Errorf("max retries (%d) exceeded for %s: %w", cfg.MaxRetries, operation, lastErr)
		}

		// Use Retry-After if provided, otherwise use exponential backoff
		waitTime := retryAfter
		if waitTime == 0 {
			waitTime = backoff
		}

		// Cap at max backoff
		if waitTime > cfg.MaxBackoff {
			waitTime = cfg.MaxBackoff
		}

		fmt.Printf("⏳ Rate limited on %s. Waiting %v before retry (%d/%d)...\n",
			operation, waitTime.Round(time.Second), attempt+1, cfg.MaxRetries)

		time.Sleep(waitTime)

		// Exponential backoff for next attempt
		backoff *= 2
		if backoff > cfg.MaxBackoff {
			backoff = cfg.MaxBackoff
		}
	}

	return result, lastErr
}

// FetchChannelsWithRetry fetches channels with automatic retry on rate limit
func FetchChannelsWithRetry(client *slack.Client, cfg RetryConfig, refresh bool) ([]slack.Channel, error) {
	return withRetry(cfg, "GetConversations", func() ([]slack.Channel, error) {
		return FetchChannels(client, refresh)
	})
}

// FetchUsersWithRetry fetches users with automatic retry on rate limit
func FetchUsersWithRetry(client *slack.Client, cfg RetryConfig, refresh bool) (map[string]string, error) {
	return withRetry(cfg, "GetUsers", func() (map[string]string, error) {
		return FetchUsers(client, refresh)
	})
}

// FetchHistoryWithRetryAndProgress fetches channel history with retry and progress callback
func FetchHistoryWithRetryAndProgress(client *slack.Client, channelID string, cfg RetryConfig, callback ProgressCallback, oldest string) ([]Message, error) {
	var allMessages []Message
	params := &slack.GetConversationHistoryParameters{
		ChannelID: channelID,
		Limit:     200, // Smaller batches to reduce rate limit impact
		Oldest:    oldest,
	}

	if callback != nil {
		callback(0, 0, "Fetching history...")
	}

	for {
		// Fetch history with retry
		history, err := withRetry(cfg, "GetConversationHistory", func() (*slack.GetConversationHistoryResponse, error) {
			return client.GetConversationHistory(params)
		})
		if err != nil {
			return nil, err
		}

		for i, msg := range history.Messages {
			richMsg := Message{Message: msg}

			// Fetch thread replies if any
			if msg.ReplyCount > 0 {
				if callback != nil {
					callback(len(allMessages)+i, 0, fmt.Sprintf("Fetching thread (%d replies)...", msg.ReplyCount))
				}

				// Fetch replies with retry
				replies, err := withRetry(cfg, "GetConversationReplies", func() ([]slack.Message, error) {
					msgs, _, _, err := client.GetConversationReplies(&slack.GetConversationRepliesParameters{
						ChannelID: channelID,
						Timestamp: msg.Timestamp,
						Limit:     200,
					})
					return msgs, err
				})

				if err == nil && len(replies) > 1 {
					richMsg.Replies = replies[1:]
				}
			}
			allMessages = append(allMessages, richMsg)
		}

		if callback != nil {
			callback(len(allMessages), 0, "Fetching history...")
		}

		if !history.HasMore {
			break
		}
		params.Cursor = history.ResponseMetaData.NextCursor

		// Small delay between pagination to be nice to the API
		time.Sleep(100 * time.Millisecond)
	}

	return allMessages, nil
}
