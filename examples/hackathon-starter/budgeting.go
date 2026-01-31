package main

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/becomeliminal/nim-go-sdk/core"
	"github.com/becomeliminal/nim-go-sdk/tools"
)

// ============================================================================
// BUDGET DATA STRUCTURES
// ============================================================================
// Budget represents a spending limit for a category or general spending
type Budget struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Name      string    `json:"name"`
	Limit     float64   `json:"limit"`
	Category  string    `json:"category,omitempty"` // Optional category filter
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}

// Global storage for budgets (keyed by userID -> slice of budgets)
var (
	budgets           = make(map[string][]Budget)
	budgetsMu         sync.RWMutex
	budgetIDCounter   int
	budgetIDCounterMu sync.Mutex
)

// generateBudgetID creates a unique budget ID
func generateBudgetID() string {
	budgetIDCounterMu.Lock()
	defer budgetIDCounterMu.Unlock()
	budgetIDCounter++
	return fmt.Sprintf("budget_%d", budgetIDCounter)
}

// ============================================================================
// CUSTOM TOOL: CREATE BUDGET
// ============================================================================
// Allows users to create spending budgets with automatic transaction tracking

func createBudgetTool() core.Tool {
	return tools.New("create_budget").
		Description("Create a spending budget with a limit. The budget automatically tracks spending from transaction history. Defaults to current month if no dates specified.").
		Schema(tools.ObjectSchema(map[string]interface{}{
			"name":     tools.StringProperty("Name or description of the budget (e.g., 'Monthly spending', 'Food budget')"),
			"limit":    tools.NumberProperty("Maximum spending limit in dollars"),
			"category": tools.StringProperty("Optional category to filter transactions (e.g., 'food', 'entertainment', 'groceries')"),
			"end_date": tools.StringProperty("Optional end date in YYYY-MM-DD format (defaults to end of current month)"),
		})).
		Handler(func(ctx context.Context, toolParams *core.ToolParams) (*core.ToolResult, error) {
			var params struct {
				Name     string  `json:"name"`
				Limit    float64 `json:"limit"`
				Category string  `json:"category"`
				EndDate  string  `json:"end_date"`
			}
			if err := json.Unmarshal(toolParams.Input, &params); err != nil {
				return &core.ToolResult{
					Success: false,
					Error:   fmt.Sprintf("invalid input: %v", err),
				}, nil
			}

			// Validate required fields
			if params.Name == "" {
				return &core.ToolResult{
					Success: false,
					Error:   "name is required",
				}, nil
			}
			if params.Limit <= 0 {
				return &core.ToolResult{
					Success: false,
					Error:   "limit must be greater than 0",
				}, nil
			}

			// Set start date to now
			startDate := time.Now()

			// Parse or default end date
			var endDate time.Time
			if params.EndDate != "" {
				parsed, err := time.Parse("2006-01-02", params.EndDate)
				if err != nil {
					return &core.ToolResult{
						Success: false,
						Error:   fmt.Sprintf("invalid end_date format, use YYYY-MM-DD: %v", err),
					}, nil
				}
				endDate = parsed
			} else {
				// Default to end of current month
				year, month, _ := startDate.Date()
				endDate = time.Date(year, month+1, 0, 23, 59, 59, 0, startDate.Location())
			}

			// Create the budget
			budget := Budget{
				ID:        generateBudgetID(),
				UserID:    toolParams.UserID,
				Name:      params.Name,
				Limit:     params.Limit,
				Category:  params.Category,
				StartDate: startDate,
				EndDate:   endDate,
				IsActive:  true,
				CreatedAt: time.Now(),
			}

			// Store the budget
			budgetsMu.Lock()
			budgets[toolParams.UserID] = append(budgets[toolParams.UserID], budget)
			budgetsMu.Unlock()

			daysRemaining := int(endDate.Sub(startDate).Hours() / 24)

			result := map[string]interface{}{
				"success":        true,
				"message":        fmt.Sprintf("Budget '%s' created successfully!", budget.Name),
				"budget_id":      budget.ID,
				"name":           budget.Name,
				"limit":          fmt.Sprintf("$%.2f", budget.Limit),
				"category":       budget.Category,
				"start_date":     budget.StartDate.Format("January 2, 2006"),
				"end_date":       budget.EndDate.Format("January 2, 2006"),
				"days_remaining": daysRemaining,
				"note":           "Spending will be automatically tracked from your transaction history",
			}

			return &core.ToolResult{
				Success: true,
				Data:    result,
			}, nil
		}).
		Build()
}

// ============================================================================
// CUSTOM TOOL: GET BUDGETS
// ============================================================================
// Fetches budgets and calculates spending from transaction history

func createGetBudgetsTool(liminalExecutor core.ToolExecutor) core.Tool {
	return tools.New("get_budgets").
		Description("View all budgets with current spending calculated from transaction history. Shows spending status, percentage used, and remaining balance.").
		Schema(tools.ObjectSchema(map[string]interface{}{
			"category": tools.StringProperty("Optional category to filter budgets by"),
		})).
		Handler(func(ctx context.Context, toolParams *core.ToolParams) (*core.ToolResult, error) {
			var params struct {
				Category string `json:"category"`
			}
			_ = json.Unmarshal(toolParams.Input, &params)

			// Retrieve user's budgets
			budgetsMu.RLock()
			userBudgets := budgets[toolParams.UserID]
			budgetsMu.RUnlock()

			if len(userBudgets) == 0 {
				return &core.ToolResult{
					Success: true,
					Data: map[string]interface{}{
						"message":       "No budgets found. Use create_budget to create one!",
						"budgets":       []interface{}{},
						"total_budgets": 0,
					},
				}, nil
			}

			// Fetch transaction history
			txRequest := map[string]interface{}{
				"limit": 100,
			}
			txRequestJSON, _ := json.Marshal(txRequest)

			txResponse, err := liminalExecutor.Execute(ctx, &core.ExecuteRequest{
				UserID:    toolParams.UserID,
				Tool:      "get_transactions",
				Input:     txRequestJSON,
				RequestID: toolParams.RequestID,
			})

			var transactions []map[string]interface{}
			if err == nil && txResponse.Success {
				var txData map[string]interface{}
				if err := json.Unmarshal(txResponse.Data, &txData); err == nil {
					if txArray, ok := txData["transactions"].([]interface{}); ok {
						for _, tx := range txArray {
							if txMap, ok := tx.(map[string]interface{}); ok {
								transactions = append(transactions, txMap)
							}
						}
					}
				}
			}

			// Format budgets with calculated spending
			var formattedBudgets []map[string]interface{}
			now := time.Now()

			for _, budget := range userBudgets {
				// Filter by category if specified
				if params.Category != "" && budget.Category != params.Category {
					continue
				}

				// Calculate spending from transactions
				currentSpent := calculateSpendingForBudget(budget, transactions)

				// Calculate progress
				daysRemaining := int(budget.EndDate.Sub(now).Hours() / 24)
				if daysRemaining < 0 {
					daysRemaining = 0
				}

				totalDays := int(budget.EndDate.Sub(budget.StartDate).Hours() / 24)
				daysElapsed := totalDays - daysRemaining
				if daysElapsed < 0 {
					daysElapsed = 0
				}

				percentUsed := 0.0
				if budget.Limit > 0 {
					percentUsed = (currentSpent / budget.Limit) * 100
				}

				// Determine status
				status := "under_budget"
				statusMessage := "You're doing great! 💚"
				if percentUsed >= 100 {
					status = "over_budget"
					statusMessage = "⚠️ Budget exceeded! You've spent more than your limit."
				} else if percentUsed >= 80 {
					status = "approaching_limit"
					statusMessage = "⚡ Approaching your limit! Spend carefully."
				} else if percentUsed >= 50 {
					status = "halfway"
					statusMessage = "Halfway through your budget 📊"
				}

				remaining := budget.Limit - currentSpent
				if remaining < 0 {
					remaining = 0
				}

				formattedBudget := map[string]interface{}{
					"id":             budget.ID,
					"name":           budget.Name,
					"category":       budget.Category,
					"limit":          fmt.Sprintf("$%.2f", budget.Limit),
					"current_spent":  fmt.Sprintf("$%.2f", currentSpent),
					"remaining":      fmt.Sprintf("$%.2f", remaining),
					"percent_used":   fmt.Sprintf("%.1f%%", percentUsed),
					"start_date":     budget.StartDate.Format("January 2, 2006"),
					"end_date":       budget.EndDate.Format("January 2, 2006"),
					"days_remaining": daysRemaining,
					"status":         status,
					"status_message": statusMessage,
					"is_active":      budget.IsActive,
				}
				formattedBudgets = append(formattedBudgets, formattedBudget)
			}

			result := map[string]interface{}{
				"budgets":       formattedBudgets,
				"total_budgets": len(formattedBudgets),
				"retrieved_at":  now.Format(time.RFC3339),
				"note":          "Spending is calculated from your transaction history",
			}

			if params.Category != "" {
				result["filtered_by_category"] = params.Category
			}

			return &core.ToolResult{
				Success: true,
				Data:    result,
			}, nil
		}).
		Build()
}

// calculateSpendingForBudget calculates total spending for a budget from transactions
func calculateSpendingForBudget(budget Budget, transactions []map[string]interface{}) float64 {
	var totalSpent float64

	for _, tx := range transactions {
		// Parse transaction timestamp
		txTimeStr, _ := tx["timestamp"].(string)
		txTime, err := time.Parse(time.RFC3339, txTimeStr)
		if err != nil {
			continue
		}

		// Check if transaction is within budget period
		if txTime.Before(budget.StartDate) || txTime.After(budget.EndDate) {
			continue
		}

		// Only count outgoing transactions (sends)
		txType, _ := tx["type"].(string)
		if txType != "send" {
			continue
		}

		// If budget has a category, filter by category
		if budget.Category != "" {
			txCategory, _ := tx["category"].(string)
			// Simple category matching (case-insensitive contains)
			if !containsIgnoreCase(txCategory, budget.Category) {
				// Also check description/memo
				description, _ := tx["description"].(string)
				memo, _ := tx["memo"].(string)
				if !containsIgnoreCase(description, budget.Category) && !containsIgnoreCase(memo, budget.Category) {
					continue
				}
			}
		}

		// Add amount to total
		amount, _ := tx["amount"].(float64)
		totalSpent += amount
	}

	return totalSpent
}

// Helper function for case-insensitive string matching
func containsIgnoreCase(s, substr string) bool {
	if s == "" || substr == "" {
		return false
	}
	// Simple lowercase comparison
	sLower := toLower(s)
	substrLower := toLower(substr)
	return contains(sLower, substrLower)
}

// Simple toLower implementation
func toLower(s string) string {
	result := make([]rune, len(s))
	for i, r := range s {
		if r >= 'A' && r <= 'Z' {
			result[i] = r + 32
		} else {
			result[i] = r
		}
	}
	return string(result)
}
