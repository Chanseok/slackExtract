package llm

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// SaveAnalysisReport saves the analysis result as a Markdown report
func SaveAnalysisReport(result *AnalysisResult, outputDir string) error {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	filename := filepath.Join(outputDir, fmt.Sprintf("%s_analysis.md", result.ChannelName))
	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer f.Close()

	// Header
	fmt.Fprintf(f, "# 📊 채널 분석 보고서: #%s\n\n", result.ChannelName)
	fmt.Fprintf(f, "> **분석 일시:** %s  \n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Fprintf(f, "> **총 메시지 수:** %d  \n\n", result.TotalMessages)
	fmt.Fprintln(f, "---")
	fmt.Fprintln(f, "")

	// Korean Summary
	fmt.Fprintln(f, "## 📝 종합 요약")
	fmt.Fprintln(f, "")
	fmt.Fprintln(f, result.Summary)
	fmt.Fprintln(f, "")
	fmt.Fprintln(f, "---")
	fmt.Fprintln(f, "")

	// Topics
	fmt.Fprintln(f, "## 🎯 주요 토픽")
	fmt.Fprintln(f, "")
	for i, topic := range result.Topics {
		importance := strings.Repeat("⭐", min(topic.Importance, 10)/2)
		if topic.Importance%2 == 1 {
			importance += "☆"
		}
		
		fmt.Fprintf(f, "### %d. %s %s\n\n", i+1, topic.Name, importance)
		fmt.Fprintf(f, "**설명:** %s\n\n", topic.Description)
		
		if len(topic.Keywords) > 0 {
			fmt.Fprintf(f, "**키워드:** `%s`\n\n", strings.Join(topic.Keywords, "`, `"))
		}
		
		// Sentiment
		total := topic.Sentiments.Positive + topic.Sentiments.Negative + topic.Sentiments.Neutral
		if total > 0 {
			fmt.Fprintln(f, "**감정 분석:**")
			fmt.Fprintf(f, "- 긍정 😊: %d건 (%.0f%%)\n", topic.Sentiments.Positive, float64(topic.Sentiments.Positive)/float64(total)*100)
			fmt.Fprintf(f, "- 부정 😟: %d건 (%.0f%%)\n", topic.Sentiments.Negative, float64(topic.Sentiments.Negative)/float64(total)*100)
			fmt.Fprintf(f, "- 중립 😐: %d건 (%.0f%%)\n\n", topic.Sentiments.Neutral, float64(topic.Sentiments.Neutral)/float64(total)*100)
		}

		if topic.Summary != "" {
			fmt.Fprintf(f, "**요약:** %s\n\n", topic.Summary)
		}
	}
	fmt.Fprintln(f, "---")
	fmt.Fprintln(f, "")

	// Contributors
	fmt.Fprintln(f, "## 👥 주요 기여자")
	fmt.Fprintln(f, "")
	fmt.Fprintln(f, "| 이름 | 메시지 수 | 참여 토픽 |")
	fmt.Fprintln(f, "|------|----------|----------|")
	for _, c := range result.Contributors {
		topics := strings.Join(c.TopicsInvolved, ", ")
		if len(topics) > 50 {
			topics = topics[:47] + "..."
		}
		fmt.Fprintf(f, "| %s | %d | %s |\n", c.Name, c.MessageCount, topics)
	}
	fmt.Fprintln(f, "")

	// Detailed contributions
	if len(result.Contributors) > 0 {
		fmt.Fprintln(f, "### 주요 기여 내용")
		fmt.Fprintln(f, "")
		for _, c := range result.Contributors {
			if len(c.KeyContributions) > 0 {
				fmt.Fprintf(f, "**%s:**\n", c.Name)
				for _, contrib := range c.KeyContributions {
					fmt.Fprintf(f, "- %s\n", contrib)
				}
				fmt.Fprintln(f, "")
			}
		}
	}

	fmt.Println("  -> Analysis saved to:", filename)
	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
