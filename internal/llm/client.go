package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Provider represents the LLM provider type
type Provider string

const (
	ProviderOpenAI Provider = "openai"
	ProviderGemini Provider = "gemini"
)

// Client represents an LLM API client
type Client struct {
	Provider   Provider
	APIKey     string
	BaseURL    string
	Model      string
	HTTPClient *http.Client
}

// Config holds LLM configuration
type Config struct {
	Provider string // "openai" or "gemini"
	APIKey   string
	BaseURL  string // Custom base URL (optional)
	Model    string // Model name
}

// NewClient creates a new LLM client
func NewClient(cfg Config) *Client {
	provider := Provider(strings.ToLower(cfg.Provider))
	if provider == "" {
		provider = ProviderOpenAI
	}

	// Set defaults based on provider
	baseURL := cfg.BaseURL
	model := cfg.Model

	switch provider {
	case ProviderGemini:
		if baseURL == "" {
			baseURL = "https://generativelanguage.googleapis.com/v1beta"
		}
		if model == "" {
			model = "gemini-1.5-flash"
		}
	default: // OpenAI
		provider = ProviderOpenAI
		if baseURL == "" {
			baseURL = "https://api.openai.com/v1"
		}
		if model == "" {
			model = "gpt-4o-mini"
		}
	}

	return &Client{
		Provider: provider,
		APIKey:   cfg.APIKey,
		BaseURL:  baseURL,
		Model:    model,
		HTTPClient: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

// ChatMessage represents a message in the conversation
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Chat sends a chat completion request (routes to appropriate provider)
func (c *Client) Chat(messages []ChatMessage, temperature float64, maxTokens int) (string, error) {
	if c.APIKey == "" {
		return "", fmt.Errorf("LLM API key is not configured")
	}

	switch c.Provider {
	case ProviderGemini:
		return c.chatGemini(messages, temperature, maxTokens)
	default:
		return c.chatOpenAI(messages, temperature, maxTokens)
	}
}

// ============ OpenAI Implementation ============

// OpenAI request/response structures
type openAIChatRequest struct {
	Model       string        `json:"model"`
	Messages    []ChatMessage `json:"messages"`
	Temperature float64       `json:"temperature,omitempty"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
}

type openAIChatResponse struct {
	ID      string `json:"id"`
	Choices []struct {
		Message      ChatMessage `json:"message"`
		FinishReason string      `json:"finish_reason"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error,omitempty"`
}

func (c *Client) chatOpenAI(messages []ChatMessage, temperature float64, maxTokens int) (string, error) {
	reqBody := openAIChatRequest{
		Model:       c.Model,
		Messages:    messages,
		Temperature: temperature,
		MaxTokens:   maxTokens,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", c.BaseURL+"/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.APIKey)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	var chatResp openAIChatResponse
	if err := json.Unmarshal(body, &chatResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if chatResp.Error != nil {
		return "", fmt.Errorf("API error: %s", chatResp.Error.Message)
	}

	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("no response from LLM")
	}

	return chatResp.Choices[0].Message.Content, nil
}

// ============ Gemini Implementation ============

// Gemini request/response structures
type geminiRequest struct {
	Contents         []geminiContent        `json:"contents"`
	GenerationConfig *geminiGenerationConfig `json:"generationConfig,omitempty"`
	SystemInstruction *geminiContent         `json:"systemInstruction,omitempty"`
}

type geminiContent struct {
	Role  string       `json:"role,omitempty"`
	Parts []geminiPart `json:"parts"`
}

type geminiPart struct {
	Text string `json:"text"`
}

type geminiGenerationConfig struct {
	Temperature     float64 `json:"temperature,omitempty"`
	MaxOutputTokens int     `json:"maxOutputTokens,omitempty"`
}

type geminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
		FinishReason string `json:"finishReason"`
	} `json:"candidates"`
	Error *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Status  string `json:"status"`
	} `json:"error,omitempty"`
}

func (c *Client) chatGemini(messages []ChatMessage, temperature float64, maxTokens int) (string, error) {
	// Convert messages to Gemini format
	var contents []geminiContent
	var systemInstruction *geminiContent

	for _, msg := range messages {
		if msg.Role == "system" {
			// Gemini uses systemInstruction for system prompts
			systemInstruction = &geminiContent{
				Parts: []geminiPart{{Text: msg.Content}},
			}
			continue
		}

		role := "user"
		if msg.Role == "assistant" {
			role = "model"
		}

		contents = append(contents, geminiContent{
			Role:  role,
			Parts: []geminiPart{{Text: msg.Content}},
		})
	}

	reqBody := geminiRequest{
		Contents:          contents,
		SystemInstruction: systemInstruction,
	}

	if temperature > 0 || maxTokens > 0 {
		reqBody.GenerationConfig = &geminiGenerationConfig{
			Temperature:     temperature,
			MaxOutputTokens: maxTokens,
		}
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Gemini API URL format: /models/{model}:generateContent?key={apiKey}
	url := fmt.Sprintf("%s/models/%s:generateContent?key=%s", c.BaseURL, c.Model, c.APIKey)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	var geminiResp geminiResponse
	if err := json.Unmarshal(body, &geminiResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if geminiResp.Error != nil {
		return "", fmt.Errorf("Gemini API error: %s", geminiResp.Error.Message)
	}

	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("no response from Gemini")
	}

	return geminiResp.Candidates[0].Content.Parts[0].Text, nil
}

// SimpleChat is a convenience method for single-turn conversations
func (c *Client) SimpleChat(systemPrompt, userPrompt string) (string, error) {
	messages := []ChatMessage{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: userPrompt},
	}
	return c.Chat(messages, 0.3, 4096)
}

// DetectLanguage detects the language of the given text
func (c *Client) DetectLanguage(text string) (string, error) {
	// Truncate if too long
	if len(text) > 500 {
		text = text[:500]
	}

	prompt := `Detect the language of the following text. 
Respond with ONLY the ISO 639-1 language code (e.g., "en" for English, "nl" for Dutch, "ko" for Korean).
If mixed languages, respond with the primary language.

Text: ` + text

	messages := []ChatMessage{
		{Role: "user", Content: prompt},
	}

	result, err := c.Chat(messages, 0, 10)
	if err != nil {
		return "en", err // Default to English on error
	}

	return strings.TrimSpace(strings.ToLower(result)), nil
}

// TranslateToEnglish translates non-English text to English
func (c *Client) TranslateToEnglish(text, sourceLang string) (string, error) {
	prompt := fmt.Sprintf(`Translate the following %s text to English.
Provide ONLY the translation, no explanations.

Text: %s`, getLanguageName(sourceLang), text)

	return c.SimpleChat("You are a professional translator.", prompt)
}

// TranslateToKorean translates text to Korean
func (c *Client) TranslateToKorean(text string) (string, error) {
	prompt := `Translate the following text to Korean.
Provide ONLY the translation, no explanations.

Text: ` + text

	return c.SimpleChat("You are a professional translator specializing in technical content.", prompt)
}

func getLanguageName(code string) string {
	languages := map[string]string{
		"en": "English",
		"nl": "Dutch",
		"ko": "Korean",
		"de": "German",
		"fr": "French",
		"es": "Spanish",
		"ja": "Japanese",
		"zh": "Chinese",
	}
	if name, ok := languages[code]; ok {
		return name
	}
	return code
}
