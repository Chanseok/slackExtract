package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	UserToken           string
	DSCookie            string
	DownloadAttachments bool
}

func Load() (*Config, error) {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		// It's okay if .env doesn't exist, maybe vars are set in environment
		// But for this app, we usually expect .env
		// fmt.Println("Warning: .env file not found")
	}

	token := os.Getenv("SLACK_USER_TOKEN")
	dCookie := os.Getenv("SLACK_DS_COOKIE")
	downloadAttachments := os.Getenv("DOWNLOAD_ATTACHMENTS") == "true"

	if token == "" || dCookie == "" {
		return nil, fmt.Errorf("SLACK_USER_TOKEN (xoxc-...) and SLACK_DS_COOKIE (xoxd-...) are required")
	}

	return &Config{
		UserToken:           token,
		DSCookie:            dCookie,
		DownloadAttachments: downloadAttachments,
	}, nil
}
