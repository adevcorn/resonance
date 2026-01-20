package tool

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

// FetchURLTool allows agents to fetch content from URLs
type FetchURLTool struct {
	client *http.Client
}

// NewFetchURLTool creates a new fetch URL tool
func NewFetchURLTool() *FetchURLTool {
	return &FetchURLTool{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Name returns the tool name
func (f *FetchURLTool) Name() string {
	return "fetch_url"
}

// Description returns the tool description
func (f *FetchURLTool) Description() string {
	return "Fetch content from a URL. Use this to retrieve documentation, API responses, or web pages."
}

// Parameters returns the JSON Schema for fetch_url parameters
func (f *FetchURLTool) Parameters() json.RawMessage {
	schema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"url": map[string]interface{}{
				"type":        "string",
				"description": "The URL to fetch (must be http or https)",
			},
		},
		"required": []string{"url"},
	}

	data, _ := json.Marshal(schema)
	return data
}

// Execute fetches the URL
func (f *FetchURLTool) Execute(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
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

// ExecutionLocation returns where this tool executes
func (f *FetchURLTool) ExecutionLocation() protocol.ExecutionLocation {
	return protocol.ExecutionLocationServer
}
