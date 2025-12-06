package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"time"

	"github.com/chanseok/slackExtract/internal/config"
	"github.com/chanseok/slackExtract/internal/llm"
	"github.com/chanseok/slackExtract/internal/meta"
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

	// Initialize MetaManager
	var metaManager *meta.Manager
	exportRoot := findExportRoot(os.Args[1])
	if exportRoot != "" {
		var err error
		metaManager, err = meta.NewManager(exportRoot)
		if err != nil {
			fmt.Printf("Warning: Failed to initialize metadata manager: %v\n", err)
		} else {
			fmt.Printf("Metadata manager initialized at: %s\n", exportRoot)
		}
	}

	// Process each file
	for _, arg := range os.Args[1:] {
		if err := processArg(arg, analyzer, metaManager); err != nil {
			fmt.Printf("Error processing %s: %v\n", arg, err)
		}
	}

	if metaManager != nil {
		if err := metaManager.SaveIndex(); err != nil {
			fmt.Printf("Warning: Failed to save metadata index: %v\n", err)
		}
	}
}

func processArg(path string, analyzer *llm.ChannelAnalyzer, mm *meta.Manager) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}

	if info.IsDir() {
		entries, err := os.ReadDir(path)
		if err != nil {
			return err
		}
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".md") {
				fullPath := filepath.Join(path, entry.Name())
				if err := analyzeFile(fullPath, analyzer, mm); err != nil {
					fmt.Printf("Error analyzing %s: %v\n", fullPath, err)
				}
			}
		}
		return nil
	}

	return analyzeFile(path, analyzer, mm)
}

func findExportRoot(path string) string {
	abs, err := filepath.Abs(path)
	if err != nil {
		return ""
	}

	dir := filepath.Dir(abs)
	// Check up to 3 levels up
	for i := 0; i < 3; i++ {
		if _, err := os.Stat(filepath.Join(dir, ".meta")); err == nil {
			return dir
		}
		// Also check if we are in "export" dir even if .meta doesn't exist yet
		if filepath.Base(dir) == "export" {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return ""
}

func analyzeFile(filePath string, analyzer *llm.ChannelAnalyzer, mm *meta.Manager) error {
	// Extract channel name from filename
	base := filepath.Base(filePath)
	channelName := strings.TrimSuffix(base, ".md")

	// Check if analysis already exists
	outputDir, reportPath, err := getOutputPaths(filePath)
	if err != nil {
		return fmt.Errorf("failed to determine output path: %w", err)
	}

	if _, err := os.Stat(reportPath); err == nil {
		fmt.Printf("⏭️  Skipping %s (Analysis already exists)\n", channelName)
		return nil
	}

	fmt.Printf("\n📊 Analyzing: %s\n", filePath)

	// Read file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

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
	if err := llm.SaveAnalysisReport(result, outputDir); err != nil {
		return fmt.Errorf("failed to save report: %w", err)
	}

	// Update metadata if manager is available
	if mm != nil {
		// Find channel by name
		ch, exists := mm.GetChannelByName(channelName)
		if exists {
			analysisMeta := &meta.AnalysisMeta{
				LastAnalyzedAt: time.Now(),
				Model:          analyzer.GetClientModel(), // Need to expose this
				Provider:       analyzer.GetClientProvider(), // Need to expose this
				InputTokens:    result.Usage.PromptTokens,
				OutputTokens:   result.Usage.CompletionTokens,
				Cost:           result.EstimatedCost,
				Language:       "ko", // Assuming Korean summary
			}
			if err := mm.UpdateChannelAnalysis(ch.ID, analysisMeta); err != nil {
				fmt.Printf("Warning: Failed to update metadata for %s: %v\n", channelName, err)
			}
		} else {
			// If channel not found in index (e.g. manually exported or index missing), we can't update by ID easily
			// unless we assume channelName is ID or we add a way to add by name (which is risky without ID)
			// For now, just warn
			fmt.Printf("Warning: Channel %s not found in metadata index. Skipping metadata update.\n", channelName)
		}
	}

	return nil
}

func getOutputPaths(filePath string) (string, string, error) {
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return "", "", fmt.Errorf("failed to get absolute path: %w", err)
	}

	dir := filepath.Dir(absPath)
	var exportDir string
	var categoryName string

	// Try to find "export" directory in the path
	if strings.HasSuffix(dir, "export") {
		// File is directly under export/ (e.g. export/channel.md)
		exportDir = dir
		categoryName = "" // No category
	} else if strings.HasSuffix(filepath.Dir(dir), "export") {
		// File is in a subdirectory (e.g. export/category/channel.md)
		exportDir = filepath.Dir(dir)
		categoryName = filepath.Base(dir)
	} else {
		// Fallback: use current dir as base
		exportDir = dir
		categoryName = ""
	}

	outputDir := filepath.Join(exportDir, ".analysis", categoryName)

	// Extract channel name from filename
	base := filepath.Base(filePath)
	channelName := strings.TrimSuffix(base, ".md")

	reportPath := filepath.Join(outputDir, fmt.Sprintf("%s_analysis.md", channelName))

	return outputDir, reportPath, nil
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
