package meta

import "time"

// Index represents the global index of all exported channels
type Index struct {
	LastUpdated time.Time           `json:"last_updated"`
	Channels    map[string]*Channel `json:"channels"`
}

// Channel represents metadata for a single channel
type Channel struct {
	ID               string        `json:"id"`
	Name             string        `json:"name"`
	Path             string        `json:"path"` // Relative path to the markdown file
	MessageCount     int           `json:"message_count"`
	LastMessageAt    time.Time     `json:"last_message_at"`
	LastDownloadedAt time.Time     `json:"last_downloaded_at"`
	Analysis         *AnalysisMeta `json:"analysis,omitempty"`
}

// AnalysisMeta contains information about the last LLM analysis
type AnalysisMeta struct {
	LastAnalyzedAt time.Time `json:"last_analyzed_at"`
	Model          string    `json:"model"`
	Provider       string    `json:"provider"`
	InputTokens    int       `json:"input_tokens"`
	OutputTokens   int       `json:"output_tokens"`
	Cost           float64   `json:"cost"` // Estimated cost in USD
	Language       string    `json:"language"`
}

// NewIndex creates a new empty index
func NewIndex() *Index {
	return &Index{
		LastUpdated: time.Now(),
		Channels:    make(map[string]*Channel),
	}
}
