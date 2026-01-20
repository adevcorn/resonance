package capability

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/adevcorn/ensemble/internal/protocol"
)

// FetchURLInput is the input for fetching a URL
type FetchURLInput struct {
	URL string `json:"url"`
}

// FetchURLOutput is the result of fetching a URL
type FetchURLOutput struct {
	URL        string            `json:"url"`
	StatusCode int               `json:"status_code"`
	Headers    map[string]string `json:"headers"`
	Body       string            `json:"body"`
}

// FetchURLCapability allows fetching content from URLs
type FetchURLCapability struct {
	client *http.Client
}

// NewFetchURLCapability creates a new fetch URL capability
func NewFetchURLCapability() *FetchURLCapability {
	return &FetchURLCapability{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Name returns the capability name
func (f *FetchURLCapability) Name() string {
	return "fetch_url"
}

// Execute fetches the URL
func (f *FetchURLCapability) Execute(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
	var urlInput FetchURLInput
	if err := json.Unmarshal(input, &urlInput); err != nil {
		return nil, fmt.Errorf("invalid fetch_url input: %w", err)
	}

	if urlInput.URL == "" {
		return nil, fmt.Errorf("url cannot be empty")
	}

	// Create request with context
	req, err := http.NewRequestWithContext(ctx, "GET", urlInput.URL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set user agent
	req.Header.Set("User-Agent", "Ensemble-Agent/1.0")

	// Execute request
	resp, err := f.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch URL: %w", err)
	}
	defer resp.Body.Close()

	// Read response body (limit to 1MB)
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1024*1024))
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Convert headers to simple map
	headers := make(map[string]string)
	for key, values := range resp.Header {
		if len(values) > 0 {
			headers[key] = values[0]
		}
	}

	output := FetchURLOutput{
		URL:        urlInput.URL,
		StatusCode: resp.StatusCode,
		Headers:    headers,
		Body:       string(body),
	}

	data, err := json.Marshal(output)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal output: %w", err)
	}

	return data, nil
}

// ExecutionLocation returns where this capability executes
func (f *FetchURLCapability) ExecutionLocation() protocol.ExecutionLocation {
	return protocol.ExecutionLocationServer
}

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

// WebSearchCapability allows searching the web
type WebSearchCapability struct{}

// NewWebSearchCapability creates a new web search capability
func NewWebSearchCapability() *WebSearchCapability {
	return &WebSearchCapability{}
}

// Name returns the capability name
func (w *WebSearchCapability) Name() string {
	return "web_search"
}

// Execute performs the web search
func (w *WebSearchCapability) Execute(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
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

// ExecutionLocation returns where this capability executes
func (w *WebSearchCapability) ExecutionLocation() protocol.ExecutionLocation {
	return protocol.ExecutionLocationServer
}
