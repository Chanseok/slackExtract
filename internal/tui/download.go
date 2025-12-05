package tui

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/chanseok/slackExtract/internal/export"
	"github.com/chanseok/slackExtract/internal/slack"
)

func waitForUpdate(sub chan ProgressMsg) tea.Cmd {
	return func() tea.Msg {
		return <-sub
	}
}

func startDownload(m Model) tea.Cmd {
	return func() tea.Msg {
		go func() {
			defer close(m.ProgressChannel)

			currentChannelIdx := 0

			for channelID := range m.Selected {
				currentChannelIdx++
				
				// Find channel name
				channelName := channelID
				for _, ch := range m.Channels {
					if ch.ID == channelID {
						channelName = ch.Name
						break
					}
				}

				// Report start of channel
				m.ProgressChannel <- ProgressMsg{
					ChannelName: channelName,
					Current:     0,
					Total:       0,
					Status:      "Starting...",
					Done:        false,
					AllDone:     false,
				}

				// Fetch History with retry support
				retryCfg := slack.DefaultRetryConfig()
				msgs, err := slack.FetchHistoryWithRetryAndProgress(m.SlackClient, channelID, retryCfg, func(current, total int, status string) {
					m.ProgressChannel <- ProgressMsg{
						ChannelName: channelName,
						Current:     current,
						Total:       total,
						Status:      status,
						Done:        false,
						AllDone:     false,
					}
				})

				if err != nil {
					m.ProgressChannel <- ProgressMsg{
						ChannelName: channelName,
						Err:         fmt.Errorf("failed to fetch history: %w", err),
						Done:        true,
					}
					continue
				}

				// Save to Markdown
				m.ProgressChannel <- ProgressMsg{
					ChannelName: channelName,
					Current:     len(msgs),
					Total:       len(msgs),
					Status:      "Saving to Markdown & Downloading files...",
					Done:        false,
				}

				err = export.SaveToMarkdown(m.HTTPClient, channelName, msgs, m.UserMap, m.Config.DownloadAttachments)
				if err != nil {
					m.ProgressChannel <- ProgressMsg{
						ChannelName: channelName,
						Err:         fmt.Errorf("failed to save: %w", err),
						Done:        true,
					}
				} else {
					m.ProgressChannel <- ProgressMsg{
						ChannelName: channelName,
						Current:     len(msgs),
						Total:       len(msgs),
						Status:      "Done",
						Done:        true,
					}
				}
				
				// Small pause to let the user see "Done"
				time.Sleep(500 * time.Millisecond)
			}

			// All done
			m.ProgressChannel <- ProgressMsg{
				AllDone: true,
			}
		}()
		return nil // The command itself returns nothing immediately, the goroutine sends msgs
	}
}
