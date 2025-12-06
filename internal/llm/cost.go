package llm

import "strings"

// ModelPrice represents the cost per million tokens
type ModelPrice struct {
	InputPricePerMillion  float64
	OutputPricePerMillion float64
}

// ModelPrices holds the pricing for known models
// Prices are in USD
var ModelPrices = map[string]ModelPrice{
	// OpenAI
	"gpt-4o":          {5.00, 15.00},
	"gpt-4o-2024-05-13": {5.00, 15.00},
	"gpt-4o-mini":     {0.15, 0.60},
	"gpt-4-turbo":     {10.00, 30.00},
	"gpt-3.5-turbo":   {0.50, 1.50},

	// Gemini
	"gemini-1.5-flash": {0.075, 0.30},
	"gemini-1.5-pro":   {3.50, 10.50},
	"gemini-1.0-pro":   {0.50, 1.50},
}

// CalculateCost calculates the estimated cost for the given usage and model
func CalculateCost(model string, usage Usage) float64 {
	// Normalize model name
	model = strings.ToLower(model)
	
	// Handle versioned models (e.g., gemini-1.5-flash-001 -> gemini-1.5-flash)
	// This is a simple heuristic
	var price ModelPrice
	var found bool

	// Exact match
	if p, ok := ModelPrices[model]; ok {
		price = p
		found = true
	} else {
		// Prefix match
		for k, p := range ModelPrices {
			if strings.HasPrefix(model, k) {
				price = p
				found = true
				break
			}
		}
	}

	if !found {
		return 0.0
	}

	inputCost := (float64(usage.PromptTokens) / 1_000_000.0) * price.InputPricePerMillion
	outputCost := (float64(usage.CompletionTokens) / 1_000_000.0) * price.OutputPricePerMillion

	return inputCost + outputCost
}
