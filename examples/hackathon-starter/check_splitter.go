package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/becomeliminal/nim-go-sdk/core"
	"github.com/becomeliminal/nim-go-sdk/tools"
)

// ============================================================================
// CUSTOM TOOL: CHECK SPLITTER
// ============================================================================
// Split bills with friends and calculate projected balance after collection

func createCheckSplitterTool(liminalExecutor core.ToolExecutor) core.Tool {
	return tools.New("split_check").
		Description("Split a bill or check between the user and their friends. Validates friend accounts, calculates splits (evenly or custom amounts), and shows projected balance after collection. Returns a summary with who owes what and asks for user confirmation that the correct accounts were found.").
		Schema(tools.ObjectSchema(map[string]interface{}{
			"total_amount": tools.StringProperty("Total bill amount to split (e.g., '100.50')"),
			"currency":     tools.StringProperty("Currency code (e.g., 'USD', 'USDC')"),
			"friends": tools.ArrayProperty("Array of friend usernames to split with (e.g., ['alice', 'bob'])",
				tools.StringProperty("Friend username")),
		}, "total_amount", "currency", "friends")).
		Handler(func(ctx context.Context, toolParams *core.ToolParams) (*core.ToolResult, error) {
			// Parse input parameters
			var params struct {
				TotalAmount  string            `json:"total_amount"`
				Currency     string            `json:"currency"`
				Friends      []string          `json:"friends"`
				CustomSplits map[string]string `json:"custom_splits"`
			}
			if err := json.Unmarshal(toolParams.Input, &params); err != nil {
				return &core.ToolResult{
					Success: false,
					Error:   fmt.Sprintf("invalid input: %v", err),
				}, nil
			}

			// Validate inputs
			if len(params.Friends) == 0 {
				return &core.ToolResult{
					Success: false,
					Error:   "at least one friend username is required",
				}, nil
			}

			// Normalize custom splits keys (strip '@')
			normalizedCustomSplits := make(map[string]string)
			for k, v := range params.CustomSplits {
				normalizedCustomSplits[strings.TrimPrefix(k, "@")] = v
			}
			params.CustomSplits = normalizedCustomSplits

			// Parse total amount
			var totalAmount float64
			if _, err := fmt.Sscanf(params.TotalAmount, "%f", &totalAmount); err != nil {
				return &core.ToolResult{
					Success: false,
					Error:   fmt.Sprintf("invalid total_amount format: %v", err),
				}, nil
			}

			// STEP 1: Search and validate all friends using the existing search_users tool
			type ValidatedFriend struct {
				DisplayTag string
				UserID     string
				Name       string
			}
			validatedFriends := make([]ValidatedFriend, 0, len(params.Friends))

			for _, rawUsername := range params.Friends {
				username := strings.TrimPrefix(rawUsername, "@")
				searchRequest := map[string]interface{}{
					"query": username,
				}
				searchJSON, _ := json.Marshal(searchRequest)

				// Call the existing search_users Liminal tool
				searchResponse, err := liminalExecutor.Execute(ctx, &core.ExecuteRequest{
					UserID:    toolParams.UserID,
					Tool:      "search_users",
					Input:     searchJSON,
					RequestID: toolParams.RequestID,
				})
				if err != nil {
					return &core.ToolResult{
						Success: false,
						Error:   fmt.Sprintf("failed to search for '%s': %v", username, err),
					}, nil
				}

				if !searchResponse.Success {
					return &core.ToolResult{
						Success: false,
						Error:   fmt.Sprintf("search failed for '%s': %s. Try using their exact username or display tag.", username, searchResponse.Error),
					}, nil
				}

				// Parse search results
				var searchData map[string]interface{}
				if err := json.Unmarshal(searchResponse.Data, &searchData); err != nil {
					return &core.ToolResult{
						Success: false,
						Error:   fmt.Sprintf("failed to parse search results for '%s': %v. Raw data: %s", username, err, string(searchResponse.Data)),
					}, nil
				}

				// Extract first result
				results, ok := searchData["users"].([]interface{})
				if !ok {
					return &core.ToolResult{
						Success: false,
						Error:   fmt.Sprintf("unexpected search response format for '%s'. Data: %v", username, searchData),
					}, nil
				}

				if len(results) == 0 {
					return &core.ToolResult{
						Success: false,
						Error:   fmt.Sprintf("no user found with username '%s'. Make sure you're using their exact username as it appears on their profile.", username),
					}, nil
				}

				firstResult, ok := results[0].(map[string]interface{})
				if !ok {
					return &core.ToolResult{
						Success: false,
						Error:   fmt.Sprintf("unexpected result format for '%s'", username),
					}, nil
				}

				displayTag, _ := firstResult["displayTag"].(string)
				displayTag = strings.TrimPrefix(displayTag, "@")
				userID, _ := firstResult["userId"].(string)
				name, _ := firstResult["name"].(string)

				validatedFriends = append(validatedFriends, ValidatedFriend{
					DisplayTag: displayTag,
					UserID:     userID,
					Name:       name,
				})
			}

			// STEP 2: Calculate splits
			type Split struct {
				DisplayTag string  `json:"display_tag"`
				Name       string  `json:"name"`
				Amount     float64 `json:"amount"`
				Formatted  string  `json:"formatted"`
			}
			splits := make([]Split, 0, len(validatedFriends))
			var totalOwed float64

			if len(params.CustomSplits) > 0 {
				// Custom splits
				for _, friend := range validatedFriends {
					amountStr, exists := params.CustomSplits[friend.DisplayTag]
					if !exists {
						return &core.ToolResult{
							Success: false,
							Error:   fmt.Sprintf("custom split amount not provided for %s", friend.DisplayTag),
						}, nil
					}
					var amount float64
					if _, err := fmt.Sscanf(amountStr, "%f", &amount); err != nil {
						return &core.ToolResult{
							Success: false,
							Error:   fmt.Sprintf("invalid amount for %s: %v", friend.DisplayTag, err),
						}, nil
					}
					splits = append(splits, Split{
						DisplayTag: friend.DisplayTag,
						Name:       friend.Name,
						Amount:     amount,
						Formatted:  fmt.Sprintf("%.2f %s", amount, params.Currency),
					})
					totalOwed += amount
				}
			} else {
				// Even split among user + friends
				numPeople := len(validatedFriends) + 1 // +1 for the user
				amountPerPerson := totalAmount / float64(numPeople)
				for _, friend := range validatedFriends {
					splits = append(splits, Split{
						DisplayTag: friend.DisplayTag,
						Name:       friend.Name,
						Amount:     amountPerPerson,
						Formatted:  fmt.Sprintf("%.2f %s", amountPerPerson, params.Currency),
					})
					totalOwed += amountPerPerson
				}
			}

			// Calculate user's share
			userShare := totalAmount - totalOwed

			// STEP 3: Get current balance using the existing get_balance tool
			balanceRequest := map[string]interface{}{}
			balanceJSON, _ := json.Marshal(balanceRequest)

			balanceResponse, err := liminalExecutor.Execute(ctx, &core.ExecuteRequest{
				UserID:    toolParams.UserID,
				Tool:      "get_balance",
				Input:     balanceJSON,
				RequestID: toolParams.RequestID,
			})

			var currentBalance float64
			var balanceFormatted string
			if err == nil && balanceResponse.Success {
				var balanceData map[string]interface{}
				if err := json.Unmarshal(balanceResponse.Data, &balanceData); err == nil {
					if balances, ok := balanceData["balances"].([]interface{}); ok && len(balances) > 0 {
						for _, bal := range balances {
							balMap := bal.(map[string]interface{})
							cur, _ := balMap["currency"].(string)
							if cur == params.Currency {
								amountStr, _ := balMap["amount"].(string)
								fmt.Sscanf(amountStr, "%f", &currentBalance)
								balanceFormatted = fmt.Sprintf("%.2f %s", currentBalance, params.Currency)
								break
							}
						}
					}
				}
			}

			// STEP 4: Calculate projected balance (current + money to collect)
			projectedBalance := currentBalance + totalOwed

			// STEP 5: Return split summary
			result := map[string]interface{}{
				"total_bill":        fmt.Sprintf("%.2f %s", totalAmount, params.Currency),
				"your_share":        fmt.Sprintf("%.2f %s", userShare, params.Currency),
				"collecting_from":   splits,
				"total_to_collect":  fmt.Sprintf("%.2f %s", totalOwed, params.Currency),
				"current_balance":   balanceFormatted,
				"projected_balance": fmt.Sprintf("%.2f %s", projectedBalance, params.Currency),
				"status":            "pending_confirmation",
				"message":           "Please confirm these are the correct accounts before I help you collect the money.",
			}

			return &core.ToolResult{
				Success: true,
				Data:    result,
			}, nil
		}).
		Build()
}
