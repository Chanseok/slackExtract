package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/chanseok/slackExtract/internal/config"
	"github.com/chanseok/slackExtract/internal/llm"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	// Load config
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	if cfg.LLMAPIKey == "" {
		fmt.Println("Error: LLM API key is required for analysis.")
		fmt.Println("Please add it to your .env file:")
		fmt.Println("")
		fmt.Println("For OpenAI:")
		fmt.Println("  LLM_PROVIDER=openai")
		fmt.Println("  LLM_API_KEY=sk-...")
		fmt.Println("")
		fmt.Println("For Gemini:")
		fmt.Println("  LLM_PROVIDER=gemini")
		fmt.Println("  LLM_API_KEY=AIza...")
		os.Exit(1)
	}

	// Initialize LLM client
	fmt.Printf("Using LLM Provider: %s, Model: %s\n", cfg.LLMProvider, cfg.LLMModel)
	llmClient := llm.NewClient(llm.Config{
		Provider: cfg.LLMProvider,
		APIKey:   cfg.LLMAPIKey,
		Model:    cfg.LLMModel,
		BaseURL:  cfg.LLMBaseURL,
	})

	analyzer := llm.NewChannelAnalyzer(llmClient)

	// Process each file
	for _, arg := range os.Args[1:] {
		if err := analyzeFile(arg, analyzer); err != nil {
			fmt.Printf("Error analyzing %s: %v\n", arg, err)
		}
	}
}

func analyzeFile(filePath string, analyzer *llm.ChannelAnalyzer) error {
	fmt.Printf("\n📊 Analyzing: %s\n", filePath)

	// Read file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Extract channel name from filename
	base := filepath.Base(filePath)
	channelName := strings.TrimSuffix(base, ".md")

	fmt.Println("  🔍 Extracting topics...")
	
	// Perform analysis
	result, err := analyzer.AnalyzeChannel(channelName, string(content))
	if err != nil {
		return fmt.Errorf("analysis failed: %w", err)
	}

	// Count messages (rough estimate from ### headers)
	result.TotalMessages = strings.Count(string(content), "\n### ")

	fmt.Printf("  ✅ Found %d topics, %d contributors\n", len(result.Topics), len(result.Contributors))

	// Save report
	outputDir := filepath.Dir(filePath)
	if err := llm.SaveAnalysisReport(result, outputDir); err != nil {
		return fmt.Errorf("failed to save report: %w", err)
	}

	return nil
}

func printUsage() {
	fmt.Println("Usage: slack-analyze <file.md> [file2.md ...]")
	fmt.Println("")
	fmt.Println("Analyzes exported Slack channel files using LLM.")
	fmt.Println("")
	fmt.Println("Examples:")
	fmt.Println("  slack-analyze export/general.md")
	fmt.Println("  slack-analyze export/*.md")
	fmt.Println("")
	fmt.Println("Required environment variables:")
	fmt.Println("  LLM_API_KEY or OPENAI_API_KEY - Your OpenAI API key")
	fmt.Println("")
	fmt.Println("Optional environment variables:")
	fmt.Println("  LLM_MODEL    - Model to use (default: gpt-4o-mini)")
	fmt.Println("  LLM_BASE_URL - API base URL (for non-OpenAI providers)")
}
