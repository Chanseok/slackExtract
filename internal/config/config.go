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
	LLMProvider         string
	LLMAPIKey           string
	LLMModel            string
	LLMBaseURL          string
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
	
	// LLM Configuration (optional)
	llmProvider := os.Getenv("LLM_PROVIDER")
	if llmProvider == "" {
		llmProvider = "openai" // Default to OpenAI
	}
	
	llmAPIKey := os.Getenv("LLM_API_KEY")
	if llmAPIKey == "" {
		llmAPIKey = os.Getenv("OPENAI_API_KEY") // Fallback to OpenAI key
	}
	if llmAPIKey == "" {
		llmAPIKey = os.Getenv("GEMINI_API_KEY") // Fallback to Gemini key
	}
	
	llmModel := os.Getenv("LLM_MODEL")
	llmBaseURL := os.Getenv("LLM_BASE_URL")

	if token == "" || dCookie == "" {
		return nil, fmt.Errorf("SLACK_USER_TOKEN (xoxc-...) and SLACK_DS_COOKIE (xoxd-...) are required")
	}

	return &Config{
		UserToken:           token,
		DSCookie:            dCookie,
		DownloadAttachments: downloadAttachments,
		LLMProvider:         llmProvider,
		LLMAPIKey:           llmAPIKey,
		LLMModel:            llmModel,
		LLMBaseURL:          llmBaseURL,
	}, nil
}
