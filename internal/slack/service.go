package slack

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"

	"github.com/slack-go/slack"
)

// Message wraps slack.Message to include full reply history
type Message struct {
	slack.Message
	Replies []slack.Message
}

// CachedChannel stores channel info for local caching
type CachedChannel struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	IsArchived bool   `json:"is_archived"`
	IsPrivate  bool   `json:"is_private"`
	IsChannel  bool   `json:"is_channel"`
	IsGroup    bool   `json:"is_group"`
	IsIM       bool   `json:"is_im"`
	IsMpIM     bool   `json:"is_mpim"`
	IsMember   bool   `json:"is_member"`
	NumMembers int    `json:"num_members"`
	Topic      string `json:"topic"`
	Purpose    string `json:"purpose"`
	Created    int64  `json:"created"`
}

func FetchChannels(client *slack.Client) ([]slack.Channel, error) {
	cacheFile := "channels.json"

	// 1. Try to load from cache
	if _, err := os.Stat(cacheFile); err == nil {
		data, err := os.ReadFile(cacheFile)
		if err == nil {
			var cachedChannels []CachedChannel
			if err := json.Unmarshal(data, &cachedChannels); err == nil {
				fmt.Printf("Loaded %d channels from cache (channels.json).\n", len(cachedChannels))
				// Convert CachedChannel to slack.Channel
				channels := make([]slack.Channel, len(cachedChannels))
				for i, cc := range cachedChannels {
					ch := slack.Channel{}
					ch.ID = cc.ID
					ch.Name = cc.Name
					ch.IsArchived = cc.IsArchived
					ch.IsPrivate = cc.IsPrivate
					ch.IsChannel = cc.IsChannel
					ch.IsGroup = cc.IsGroup
					ch.IsIM = cc.IsIM
					ch.IsMpIM = cc.IsMpIM
					ch.IsMember = cc.IsMember
					ch.NumMembers = cc.NumMembers
					ch.Topic = slack.Topic{Value: cc.Topic}
					ch.Purpose = slack.Purpose{Value: cc.Purpose}
					ch.Created = slack.JSONTime(cc.Created)
					channels[i] = ch
				}
				return channels, nil
			}
		}
	}

	// 2. Fetch from API (with pagination)
	fmt.Println("Fetching channel list from Slack API...")
	params := &slack.GetConversationsParameters{
		Types: []string{"public_channel", "private_channel", "mpim", "im"},
		Limit: 1000,
	}

	var allChannels []slack.Channel
	for {
		channels, nextCursor, err := client.GetConversations(params)
		if err != nil {
			return nil, err
		}
		allChannels = append(allChannels, channels...)

		if nextCursor == "" {
			break
		}
		params.Cursor = nextCursor
		fmt.Printf("  ...fetched %d channels so far\n", len(allChannels))
	}
	fmt.Printf("  -> Fetched %d channels total.\n", len(allChannels))

	// 3. Save to cache
	cachedChannels := make([]CachedChannel, len(allChannels))
	for i, ch := range allChannels {
		cachedChannels[i] = CachedChannel{
			ID:         ch.ID,
			Name:       ch.Name,
			IsArchived: ch.IsArchived,
			IsPrivate:  ch.IsPrivate,
			IsChannel:  ch.IsChannel,
			IsGroup:    ch.IsGroup,
			IsIM:       ch.IsIM,
			IsMpIM:     ch.IsMpIM,
			IsMember:   ch.IsMember,
			NumMembers: ch.NumMembers,
			Topic:      ch.Topic.Value,
			Purpose:    ch.Purpose.Value,
			Created:    int64(ch.Created),
		}
	}
	data, err := json.MarshalIndent(cachedChannels, "", "  ")
	if err == nil {
		_ = os.WriteFile(cacheFile, data, 0644)
		fmt.Println("Saved channel list to cache (channels.json).")
	}

	// Sort channels by name
	sort.Slice(allChannels, func(i, j int) bool {
		return allChannels[i].Name < allChannels[j].Name
	})

	return allChannels, nil
}

func FetchUsers(client *slack.Client) (map[string]string, error) {
	cacheFile := "users.json"
	userMap := make(map[string]string)

	// 1. Try to load from cache
	if _, err := os.Stat(cacheFile); err == nil {
		data, err := os.ReadFile(cacheFile)
		if err == nil {
			if err := json.Unmarshal(data, &userMap); err == nil {
				fmt.Println("Loaded user list from cache (users.json).")
				return userMap, nil
			}
		}
	}

	// 2. Fetch from API
	fmt.Println("Fetching user list from Slack API...")
	
	// slack-go GetUsers fetches all users (handles pagination internally)
	allUsers, err := client.GetUsers()
	if err != nil {
		return nil, err
	}
	fmt.Printf("  -> Fetched %d users total.\n", len(allUsers))

	for _, u := range allUsers {
		userMap[u.ID] = u.RealName
	}

	// 3. Save to cache
	data, err := json.MarshalIndent(userMap, "", "  ")
	if err == nil {
		_ = os.WriteFile(cacheFile, data, 0644)
		fmt.Println("Saved user list to cache (users.json).")
	}

	return userMap, nil
}

func FetchHistory(client *slack.Client, channelID string) ([]Message, error) {
	var allMessages []Message
	params := &slack.GetConversationHistoryParameters{
		ChannelID: channelID,
		Limit:     1000,
	}

	fmt.Println("  Fetching messages...")
	for {
		history, err := client.GetConversationHistory(params)
		if err != nil {
			return nil, err
		}
		
		for _, msg := range history.Messages {
			richMsg := Message{Message: msg}
			
			// Fetch thread replies if any
			if msg.ReplyCount > 0 {
				fmt.Printf("    Fetching %d replies for thread %s...\n", msg.ReplyCount, msg.Timestamp)
				replies, _, _, err := client.GetConversationReplies(&slack.GetConversationRepliesParameters{
					ChannelID: channelID,
					Timestamp: msg.Timestamp,
					Limit:     1000, // Assuming threads aren't huge, but might need pagination too
				})
				if err == nil && len(replies) > 1 {
					// replies[0] is the parent message itself, so we skip it
					richMsg.Replies = replies[1:]
				} else if err != nil {
					fmt.Printf("    Warning: Failed to fetch replies: %v\n", err)
				}
			}
			allMessages = append(allMessages, richMsg)
		}

		if !history.HasMore {
			break
		}
		params.Cursor = history.ResponseMetaData.NextCursor
	}
	return allMessages, nil
}

// ProgressCallback is a function type for reporting progress
type ProgressCallback func(current int, total int, status string)

func FetchHistoryWithProgress(client *slack.Client, channelID string, callback ProgressCallback, oldest string) ([]Message, error) {
	var allMessages []Message
	params := &slack.GetConversationHistoryParameters{
		ChannelID: channelID,
		Limit:     1000,
		Oldest:    oldest, // Support incremental fetch
	}

	if callback != nil {
		callback(0, 0, "Fetching history...")
	}

	for {
		history, err := client.GetConversationHistory(params)
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
				replies, _, _, err := client.GetConversationReplies(&slack.GetConversationRepliesParameters{
					ChannelID: channelID,
					Timestamp: msg.Timestamp,
					Limit:     1000,
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
	}
	return allMessages, nil
}
