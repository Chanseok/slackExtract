package slack

import (
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"

	"github.com/chanseok/slackExtract/internal/config"
	"github.com/slack-go/slack"
)

func NewClient(cfg *config.Config) (*slack.Client, *http.Client, error) {
	// Create a custom HTTP client with the cookie
	jar, _ := cookiejar.New(nil)
	u, _ := url.Parse("https://slack.com")
	jar.SetCookies(u, []*http.Cookie{
		{
			Name:   "d",
			Value:  cfg.DSCookie,
			Path:   "/",
			Domain: ".slack.com",
		},
	})

	httpClient := &http.Client{
		Jar: jar,
	}

	// Initialize Slack API with custom client
	api := slack.New(cfg.UserToken, slack.OptionHTTPClient(httpClient))

	// Auth Test
	authTest, err := api.AuthTest()
	if err != nil {
		return nil, nil, fmt.Errorf("error connecting to Slack: %w", err)
	}
	fmt.Printf("Successfully authenticated as: %s (Team: %s)\n", authTest.User, authTest.Team)

	return api, httpClient, nil
}
