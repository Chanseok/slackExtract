package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/slack-go/slack"
)

func main() {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		// It's okay if .env doesn't exist, we might be using real env vars
		// But for this dev setup, we'll log a warning
		fmt.Println("Warning: .env file not found")
	}

	token := os.Getenv("SLACK_USER_TOKEN")
	if token == "" {
		fmt.Println("Error: SLACK_USER_TOKEN is not set in .env or environment variables.")
		fmt.Println("Please create a .env file with your token: SLACK_USER_TOKEN=xoxp-...")
		os.Exit(1)
	}

	fmt.Println("Slack Extract - Initializing...")
	
	api := slack.New(token)
	authTest, err := api.AuthTest()
	if err != nil {
		fmt.Printf("Error connecting to Slack: %v\n", err)
		fmt.Println("Tip: Check if your 'd' cookie and 'xoxc' token are valid and not expired.")
		os.Exit(1)
	}

	fmt.Printf("Successfully authenticated as: %s (Team: %s)\n", authTest.User, authTest.Team)
}
