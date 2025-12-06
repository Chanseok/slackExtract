package llm

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// AnalysisResult holds the complete analysis of a channel
type AnalysisResult struct {
	ChannelName   string
	TotalMessages int
	DateRange     string
	Topics        []Topic
	Contributors  []Contributor
	Summary       string // Korean summary
}

// Topic represents an identified discussion topic
type Topic struct {
	Name        string
	Description string
	Importance  int      // 1-10 scale
	MessageIDs  []string // References to messages
	Summary     string
	Sentiments  Sentiments
	Keywords    []string
}

// Sentiments holds sentiment analysis results
type Sentiments struct {
	Positive int
	Negative int
	Neutral  int
	KeyPoints struct {
		Agreements    []string
		Disagreements []string
		Questions     []string
	}
}

// Contributor represents a person's participation stats
type Contributor struct {
	Name          string
	MessageCount  int
	TopicsInvolved []string
	KeyContributions []string
}

// ChannelAnalyzer performs LLM-based analysis on channel messages
type ChannelAnalyzer struct {
	client *Client
}

// NewChannelAnalyzer creates a new analyzer
func NewChannelAnalyzer(client *Client) *ChannelAnalyzer {
	return &ChannelAnalyzer{client: client}
}

// AnalyzeChannel performs comprehensive analysis on channel content
func (a *ChannelAnalyzer) AnalyzeChannel(channelName, content string) (*AnalysisResult, error) {
	result := &AnalysisResult{
		ChannelName: channelName,
	}

	// Step 1: Extract Topics
	topics, err := a.extractTopics(content)
	if err != nil {
		return nil, fmt.Errorf("failed to extract topics: %w", err)
	}
	result.Topics = topics

	// Step 2: Analyze Contributors
	contributors, err := a.analyzeContributors(content)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze contributors: %w", err)
	}
	result.Contributors = contributors

	// Step 3: Generate Korean Summary
	summary, err := a.generateKoreanSummary(channelName, content, topics)
	if err != nil {
		return nil, fmt.Errorf("failed to generate summary: %w", err)
	}
	result.Summary = summary

	return result, nil
}

// extractTopics identifies main discussion topics from the content
func (a *ChannelAnalyzer) extractTopics(content string) ([]Topic, error) {
	// Truncate content if too long (LLM context limit)
	truncatedContent := truncateForLLM(content, 15000)

	prompt := `Analyze the following Slack channel conversation and identify the main discussion topics.

For each topic, provide:
1. Topic name (short, descriptive)
2. Brief description (1-2 sentences)
3. Importance score (1-10, based on discussion length, participant count, urgency keywords)
4. Key keywords (3-5 words)
5. Sentiment breakdown (positive/negative/neutral message count estimate)

Format your response as JSON:
{
  "topics": [
    {
      "name": "Topic Name",
      "description": "Brief description",
      "importance": 8,
      "keywords": ["keyword1", "keyword2"],
      "sentiment": {"positive": 5, "negative": 2, "neutral": 10}
    }
  ]
}

Conversation:
` + truncatedContent

	messages := []ChatMessage{
		{Role: "system", Content: "You are an expert at analyzing team communications and identifying key discussion topics."},
		{Role: "user", Content: prompt},
	}

	// Increased max tokens to 16000 to support reasoning models like gemini-2.5-flash which use tokens for thinking
	response, err := a.client.Chat(messages, 0.2, 16000)
	if err != nil {
		return nil, err
	}

	// Parse JSON response
	topics := parseTopicsFromJSON(response)
	return topics, nil
}

// analyzeContributors identifies key contributors and their involvement
func (a *ChannelAnalyzer) analyzeContributors(content string) ([]Contributor, error) {
	truncatedContent := truncateForLLM(content, 15000)

	prompt := `Analyze the following Slack conversation and identify the key contributors.

For each significant contributor, provide:
1. Name
2. Approximate message count
3. Topics they're most involved in
4. Their key contributions or viewpoints (1-2 sentences each)

Format your response as JSON:
{
  "contributors": [
    {
      "name": "Person Name",
      "message_count": 15,
      "topics": ["Topic 1", "Topic 2"],
      "contributions": ["Led discussion on X", "Proposed solution for Y"]
    }
  ]
}

Conversation:
` + truncatedContent

	messages := []ChatMessage{
		{Role: "system", Content: "You are an expert at analyzing team dynamics and identifying key contributors in discussions."},
		{Role: "user", Content: prompt},
	}

	response, err := a.client.Chat(messages, 0.2, 16000)
	if err != nil {
		return nil, err
	}

	contributors := parseContributorsFromJSON(response)
	return contributors, nil
}

// generateKoreanSummary creates a comprehensive Korean summary
func (a *ChannelAnalyzer) generateKoreanSummary(channelName, content string, topics []Topic) (string, error) {
	truncatedContent := truncateForLLM(content, 12000)

	// Build topic context
	var topicList strings.Builder
	for i, t := range topics {
		topicList.WriteString(fmt.Sprintf("%d. %s (중요도: %d/10)\n", i+1, t.Name, t.Importance))
	}

	// Determine channel type and get specific prompt
	channelType := determineChannelType(channelName)
	specificPrompt := getPromptForChannelType(channelType)

	prompt := fmt.Sprintf(`다음은 Slack 채널 #%s의 대화 내용입니다.

주요 논의 주제:
%s

대화 내용을 분석하여 한국어로 종합 요약을 작성해주세요.

%s

대화 내용:
%s`, channelName, topicList.String(), specificPrompt, truncatedContent)

	messages := []ChatMessage{
		{Role: "system", Content: "당신은 팀 커뮤니케이션을 분석하고 핵심 내용을 명확하게 요약하는 전문가입니다. 항상 한국어로 응답합니다."},
		{Role: "user", Content: prompt},
	}

	return a.client.Chat(messages, 0.2, 16000)
}

// ProcessMultilingualContent handles translation of non-English messages
func (a *ChannelAnalyzer) ProcessMultilingualContent(content string) (string, error) {
	lines := strings.Split(content, "\n")
	var result strings.Builder

	for _, line := range lines {
		// Skip empty lines, headers, and metadata
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, ">") || strings.HasPrefix(trimmed, "---") {
			result.WriteString(line + "\n")
			continue
		}

		// Check if line contains actual message content (not timestamps, names)
		if !strings.Contains(line, " - ") && len(trimmed) > 50 {
			// Detect language
			lang, err := a.client.DetectLanguage(trimmed)
			if err != nil {
				result.WriteString(line + "\n")
				continue
			}

			// If not English, add translation
			if lang != "en" && lang != "" {
				translation, err := a.client.TranslateToEnglish(trimmed, lang)
				if err == nil && translation != "" {
					result.WriteString(fmt.Sprintf("[EN] %s\n", translation))
					result.WriteString(fmt.Sprintf("[%s] %s\n", strings.ToUpper(lang), trimmed))
					continue
				}
			}
		}

		result.WriteString(line + "\n")
	}

	return result.String(), nil
}

// Helper functions

func truncateForLLM(content string, maxChars int) string {
	if len(content) <= maxChars {
		return content
	}
	// Try to truncate at a reasonable boundary
	truncated := content[:maxChars]
	lastNewline := strings.LastIndex(truncated, "\n")
	if lastNewline > maxChars/2 {
		return truncated[:lastNewline] + "\n\n[... content truncated ...]"
	}
	return truncated + "\n\n[... content truncated ...]"
}

func parseTopicsFromJSON(response string) []Topic {
	// Extract JSON from response (it might be wrapped in markdown code blocks)
	jsonStr := extractJSON(response)
	
	var data struct {
		Topics []struct {
			Name        string   `json:"name"`
			Description string   `json:"description"`
			Importance  int      `json:"importance"`
			Keywords    []string `json:"keywords"`
			Sentiment   struct {
				Positive int `json:"positive"`
				Negative int `json:"negative"`
				Neutral  int `json:"neutral"`
			} `json:"sentiment"`
		} `json:"topics"`
	}

	if err := parseJSON(jsonStr, &data); err != nil {
		return nil
	}

	var topics []Topic
	for _, t := range data.Topics {
		topic := Topic{
			Name:        t.Name,
			Description: t.Description,
			Importance:  t.Importance,
			Keywords:    t.Keywords,
			Sentiments: Sentiments{
				Positive: t.Sentiment.Positive,
				Negative: t.Sentiment.Negative,
				Neutral:  t.Sentiment.Neutral,
			},
		}
		topics = append(topics, topic)
	}

	// Sort by importance
	sort.Slice(topics, func(i, j int) bool {
		return topics[i].Importance > topics[j].Importance
	})

	return topics
}

func parseContributorsFromJSON(response string) []Contributor {
	jsonStr := extractJSON(response)
	
	var data struct {
		Contributors []struct {
			Name          string   `json:"name"`
			MessageCount  int      `json:"message_count"`
			Topics        []string `json:"topics"`
			Contributions []string `json:"contributions"`
		} `json:"contributors"`
	}

	if err := parseJSON(jsonStr, &data); err != nil {
		return nil
	}

	var contributors []Contributor
	for _, c := range data.Contributors {
		contributor := Contributor{
			Name:             c.Name,
			MessageCount:     c.MessageCount,
			TopicsInvolved:   c.Topics,
			KeyContributions: c.Contributions,
		}
		contributors = append(contributors, contributor)
	}

	// Sort by message count
	sort.Slice(contributors, func(i, j int) bool {
		return contributors[i].MessageCount > contributors[j].MessageCount
	})

	return contributors
}

func extractJSON(s string) string {
	// Remove markdown code blocks if present
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "```json") {
		s = strings.TrimPrefix(s, "```json")
	} else if strings.HasPrefix(s, "```") {
		s = strings.TrimPrefix(s, "```")
	}
	if strings.HasSuffix(s, "```") {
		s = strings.TrimSuffix(s, "```")
	}
	return strings.TrimSpace(s)
}

func parseJSON(s string, v interface{}) error {
	return json.Unmarshal([]byte(s), v)
}

// determineChannelType guesses the channel type based on its name
func determineChannelType(channelName string) string {
	name := strings.ToLower(channelName)
	if strings.Contains(name, "project") {
		return "project"
	}
	if strings.Contains(name, "sales") {
		return "sales"
	}
	if strings.Contains(name, "marketing") {
		return "marketing"
	}
	return "general"
}

// getPromptForChannelType returns a specific prompt for the channel type
func getPromptForChannelType(channelType string) string {
	basePrompt := `요약에는 다음 내용을 포함해주세요:
1. 채널의 전반적인 목적과 분위기
2. 각 주요 주제에 대한 핵심 논의 내용 (2-3문장씩)
3. 주요 결정사항 또는 합의점
4. 미해결 이슈 또는 후속 조치가 필요한 사항
5. 특별히 주목할 만한 의견이나 아이디어`

	switch channelType {
	case "project":
		return `이 채널은 '프로젝트' 관련 채널입니다. 다음 관점에서 심층 분석해주세요:
1. 프로젝트 현황 및 진행 상황 (Key Status)
2. 주요 제안 및 기획 내용 (Key Proposals)
3. 기술적/기획적 논의 및 쟁점 (Key Debates)
4. 주요 의사결정 사항 (Decisions Made)
5. 향후 계획 및 액션 아이템 (Next Steps)
6. 리스크 또는 블로커 (Risks & Blockers)`
	case "sales":
		return `이 채널은 '영업(Sales)' 관련 채널입니다. 다음 관점에서 심층 분석해주세요:
1. 주요 거래 및 기회 (Key Deals & Opportunities)
2. 매출 및 성과 현황 (Revenue & Performance)
3. 고객 피드백 및 요구사항 (Client Feedback)
4. 경쟁사 동향 또는 시장 이슈 (Competition & Market)
5. 영업 활동의 주요 블로커 (Blockers)
6. 전략적 제안 또는 개선점`
	case "marketing":
		return `이 채널은 '마케팅(Marketing)' 관련 채널입니다. 다음 관점에서 심층 분석해주세요:
1. 진행 중인 캠페인 및 프로모션 (Active Campaigns)
2. 주요 성과 지표 및 분석 (KPIs & Metrics)
3. 채널별 성과 및 피드백 (Channel Performance)
4. 크리에이티브/콘텐츠 관련 논의 (Creative Feedback)
5. 예산 및 리소스 이슈 (Budget & Resources)
6. 향후 마케팅 전략 및 아이디어`
	default:
		return basePrompt
	}
}

