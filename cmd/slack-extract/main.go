package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"

	"github.com/joho/godotenv"
	"github.com/slack-go/slack"
)

func main() {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Warning: .env file not found")
	}

	token := os.Getenv("SLACK_USER_TOKEN")
	dCookie := os.Getenv("SLACK_DS_COOKIE")

	if token == "" || dCookie == "" {
		fmt.Println("Error: SLACK_USER_TOKEN (xoxc-...) and SLACK_DS_COOKIE (xoxd-...) are required.")
		fmt.Println("Please check your .env file.")
		os.Exit(1)
	}

	fmt.Println("Slack Extract - Initializing...")

	// Create a custom HTTP client with the cookie
	jar, _ := cookiejar.New(nil)
	u, _ := url.Parse("https://slack.com")
	jar.SetCookies(u, []*http.Cookie{
		{
			Name:   "d",
			Value:  dCookie,
			Path:   "/",
			Domain: ".slack.com",
		},
	})

	client := &http.Client{
		Jar: jar,
	}

	// Initialize Slack API with custom client
	api := slack.New(token, slack.OptionHTTPClient(client))

	authTest, err := api.AuthTest()
	if err != nil {
		fmt.Printf("Error connecting to Slack: %v\n", err)
		fmt.Println("Tip: Check if your 'd' cookie and 'xoxc' token are valid and not expired.")
		os.Exit(1)
	}

	fmt.Printf("Successfully authenticated as: %s (Team: %s)\n", authTest.User, authTest.Team)
}
