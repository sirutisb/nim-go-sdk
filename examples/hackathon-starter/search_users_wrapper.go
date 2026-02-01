package main

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/becomeliminal/nim-go-sdk/core"
	"github.com/becomeliminal/nim-go-sdk/tools"
)

// createSearchUsersWrapper creates a wrapper around the search_users tool
// that strips the "@" prefix from queries before calling the Liminal API
func createSearchUsersWrapper(liminalExecutor core.ToolExecutor) core.Tool {
	return tools.New("search_users").
		Description("Search for users by display tag or name. You can include @ in the query (e.g., '@alice') or omit it (e.g., 'alice').").
		Schema(tools.ObjectSchema(map[string]interface{}{
			"query": tools.StringProperty("Search query (display tag like @alice or name)"),
		}, "query")).
		Handler(func(ctx context.Context, toolParams *core.ToolParams) (*core.ToolResult, error) {
			// Parse input parameters
			var params struct {
				Query string `json:"query"`
			}
			if err := json.Unmarshal(toolParams.Input, &params); err != nil {
				return &core.ToolResult{
					Success: false,
					Error:   "invalid input: " + err.Error(),
				}, nil
			}

			// Strip @ prefix if present
			cleanQuery := strings.TrimPrefix(params.Query, "@")

			// Create new request with cleaned query
			cleanRequest := map[string]interface{}{
				"query": cleanQuery,
			}
			cleanJSON, _ := json.Marshal(cleanRequest)

			// Call the actual Liminal search_users tool
			response, err := liminalExecutor.Execute(ctx, &core.ExecuteRequest{
				UserID:    toolParams.UserID,
				Tool:      "search_users",
				Input:     cleanJSON,
				RequestID: toolParams.RequestID,
			})
			if err != nil {
				return &core.ToolResult{
					Success: false,
					Error:   "search failed: " + err.Error(),
				}, nil
			}

			// Convert ExecuteResponse to ToolResult
			return &core.ToolResult{
				Success: response.Success,
				Error:   response.Error,
				Data:    response.Data,
			}, nil
		}).
		Build()
}
