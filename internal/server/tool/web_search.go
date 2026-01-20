package tool

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/adevcorn/ensemble/internal/protocol"
)

// WebSearchInput is the input for web search
type WebSearchInput struct {
	Query string `json:"query"`
	Limit int    `json:"limit,omitempty"`
}

// SearchResult represents a single search result
type SearchResult struct {
	Title   string `json:"title"`
	URL     string `json:"url"`
	Snippet string `json:"snippet"`
}

// WebSearchOutput is the result of a web search
type WebSearchOutput struct {
	Query   string         `json:"query"`
	Results []SearchResult `json:"results"`
}

// WebSearchTool allows agents to search the web
type WebSearchTool struct {
}

// NewWebSearchTool creates a new web search tool
func NewWebSearchTool() *WebSearchTool {
	return &WebSearchTool{}
}

// Name returns the tool name
func (w *WebSearchTool) Name() string {
	return "web_search"
}

// Description returns the tool description
func (w *WebSearchTool) Description() string {
	return "Search the web for information. Returns a list of search results with titles, URLs, and snippets. Note: This is a placeholder implementation that returns mock results. To enable real search, integrate with a search API (e.g., Google Custom Search, Brave Search, DuckDuckGo)."
}

// Parameters returns the JSON Schema for web_search parameters
func (w *WebSearchTool) Parameters() json.RawMessage {
	schema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"query": map[string]interface{}{
				"type":        "string",
				"description": "The search query",
			},
			"limit": map[string]interface{}{
				"type":        "integer",
				"description": "Maximum number of results to return (default: 5)",
				"minimum":     1,
				"maximum":     10,
			},
		},
		"required": []string{"query"},
	}

	data, _ := json.Marshal(schema)
	return data
}

// Execute performs the web search
func (w *WebSearchTool) Execute(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
	var searchInput WebSearchInput
	if err := json.Unmarshal(input, &searchInput); err != nil {
		return nil, fmt.Errorf("invalid web_search input: %w", err)
	}

	if searchInput.Query == "" {
		return nil, fmt.Errorf("query cannot be empty")
	}

	if searchInput.Limit == 0 {
		searchInput.Limit = 5
	}

	// TODO: Implement real web search using a search API
	// For now, return a helpful error message
	output := WebSearchOutput{
		Query: searchInput.Query,
		Results: []SearchResult{
			{
				Title:   "Web Search Not Configured",
				URL:     "https://example.com",
				Snippet: fmt.Sprintf("Web search is not yet configured. To enable this feature, integrate with a search API provider (Google Custom Search, Brave Search, DuckDuckGo, etc.). Your query was: %q", searchInput.Query),
			},
		},
	}

	data, err := json.Marshal(output)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal output: %w", err)
	}

	return data, nil
}

// ExecutionLocation returns where this tool executes
func (w *WebSearchTool) ExecutionLocation() protocol.ExecutionLocation {
	return protocol.ExecutionLocationServer
}
