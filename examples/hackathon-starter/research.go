package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/becomeliminal/nim-go-sdk/core"
	"github.com/becomeliminal/nim-go-sdk/tools"
)

// ============================================================================
// CUSTOM TOOL: RESEARCH
// ============================================================================
// This tool uses Perplexity AI API to answer research questions

func createResearchTool() core.Tool {
	return tools.New("research").
		Description("Research any topic using Perplexity AI. Get accurate, up-to-date information with citations.").
		Schema(tools.ObjectSchema(map[string]interface{}{
			"query": tools.StringProperty("The research question or topic to investigate"),
		})).
		Handler(func(ctx context.Context, toolParams *core.ToolParams) (*core.ToolResult, error) {
			var params struct {
				Query string `json:"query"`
			}
			if err := json.Unmarshal(toolParams.Input, &params); err != nil {
				return &core.ToolResult{
					Success: false,
					Error:   fmt.Sprintf("invalid input: %v", err),
				}, nil
			}

			if params.Query == "" {
				return &core.ToolResult{
					Success: false,
					Error:   "query is required",
				}, nil
			}

			// Get Perplexity API key from environment
			apiKey := os.Getenv("PERPLEXITY_API_KEY")
			if apiKey == "" {
				return &core.ToolResult{
					Success: false,
					Error:   "PERPLEXITY_API_KEY environment variable not set",
				}, nil
			}

			// Call Perplexity API
			result, err := callPerplexityAPI(ctx, apiKey, params.Query)
			if err != nil {
				return &core.ToolResult{
					Success: false,
					Error:   fmt.Sprintf("Perplexity API error: %v", err),
				}, nil
			}

			return &core.ToolResult{
				Success: true,
				Data:    result,
			}, nil
		}).
		Build()
}

// PerplexityRequest represents the request to Perplexity API
type PerplexityRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

// Message represents a chat message
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// PerplexityResponse represents the response from Perplexity API
type PerplexityResponse struct {
	ID      string   `json:"id"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
	Usage   Usage    `json:"usage"`
}

type Choice struct {
	Index        int     `json:"index"`
	FinishReason string  `json:"finish_reason"`
	Message      Message `json:"message"`
}

type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// callPerplexityAPI makes a request to Perplexity AI
func callPerplexityAPI(ctx context.Context, apiKey, query string) (map[string]interface{}, error) {
	// Prepare request
	reqBody := PerplexityRequest{
		// Model: "llama-3.1-sonar-small-128k-online",
		Model: "sonar",
		Messages: []Message{
			{
				Role:    "user",
				Content: query,
			},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.perplexity.ai/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))

	// Make request
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var perplexityResp PerplexityResponse
	if err := json.Unmarshal(body, &perplexityResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Extract answer
	answer := ""
	if len(perplexityResp.Choices) > 0 {
		answer = perplexityResp.Choices[0].Message.Content
	}

	// Format result
	result := map[string]interface{}{
		"query":  query,
		"answer": answer,
		"model":  perplexityResp.Model,
		"usage": map[string]interface{}{
			"prompt_tokens":     perplexityResp.Usage.PromptTokens,
			"completion_tokens": perplexityResp.Usage.CompletionTokens,
			"total_tokens":      perplexityResp.Usage.TotalTokens,
		},
	}

	return result, nil
}
