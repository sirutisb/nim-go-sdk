package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/becomeliminal/nim-go-sdk/core"
	"github.com/becomeliminal/nim-go-sdk/tools"
)

// ============================================================================
// BUDGET DATA STRUCTURES
// ============================================================================
// Budget represents a spending limit for a category or general spending
type Budget struct {
	ID        int       `json:"id"`
	UserID    string    `json:"user_id"`
	Name      string    `json:"name"`
	Limit     float64   `json:"limit"`
	Category  string    `json:"category,omitempty"` // Optional category filter
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
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

			// Insert into database
			result, err := db.Exec(
				`INSERT INTO budgets (user_id, name, limit_amount, category, start_date, end_date, is_active) 
				 VALUES (?, ?, ?, ?, ?, ?, 1)`,
				toolParams.UserID, params.Name, params.Limit, params.Category,
				startDate.Format("2006-01-02"), endDate.Format("2006-01-02"),
			)
			if err != nil {
				return &core.ToolResult{
					Success: false,
					Error:   fmt.Sprintf("Failed to create budget: %v", err),
				}, nil
			}

			id, _ := result.LastInsertId()
			daysRemaining := int(endDate.Sub(startDate).Hours() / 24)

			responseData := map[string]interface{}{
				"success":        true,
				"message":        fmt.Sprintf("Budget '%s' created successfully!", params.Name),
				"budget_id":      id,
				"name":           params.Name,
				"limit":          fmt.Sprintf("$%.2f", params.Limit),
				"category":       params.Category,
				"start_date":     startDate.Format("January 2, 2006"),
				"end_date":       endDate.Format("January 2, 2006"),
				"days_remaining": daysRemaining,
				"note":           "Spending will be automatically tracked from your transaction history",
			}

			// Notify dashboard of update
			NotifyDashboardUpdate("budget", "created")

			return &core.ToolResult{
				Success: true,
				Data:    responseData,
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

			// Build query with optional category filter
			query := `SELECT id, user_id, name, limit_amount, category, start_date, end_date, is_active, created_at 
					  FROM budgets WHERE user_id = ?`
			args := []interface{}{toolParams.UserID}

			if params.Category != "" {
				query += " AND category = ?"
				args = append(args, params.Category)
			}

			rows, err := db.Query(query, args...)
			if err != nil {
				return &core.ToolResult{
					Success: false,
					Error:   fmt.Sprintf("Failed to fetch budgets: %v", err),
				}, nil
			}
			defer rows.Close()

			var userBudgets []Budget
			for rows.Next() {
				var budget Budget
				var startDateStr, endDateStr, createdAtStr string
				var isActiveInt int

				err := rows.Scan(&budget.ID, &budget.UserID, &budget.Name, &budget.Limit, &budget.Category,
					&startDateStr, &endDateStr, &isActiveInt, &createdAtStr)
				if err != nil {
					continue
				}

				budget.StartDate, _ = time.Parse("2006-01-02", startDateStr)
				budget.EndDate, _ = time.Parse("2006-01-02", endDateStr)
				budget.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAtStr)
				budget.IsActive = isActiveInt == 1

				userBudgets = append(userBudgets, budget)
			}

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

// ============================================================================
// CUSTOM TOOL: UPDATE BUDGET
// ============================================================================
// Allows users to update budget properties

func createUpdateBudgetTool() core.Tool {
	return tools.New("update_budget").
		Description("Update an existing budget's limit, name, category, or end date. Provide the budget ID or name to identify which budget to update.").
		Schema(tools.ObjectSchema(map[string]interface{}{
			"id":       tools.StringProperty("Budget ID (optional if name is provided)"),
			"name":     tools.StringProperty("Budget name to identify budget (optional if id is provided), or new name if updating"),
			"new_name": tools.StringProperty("New name for the budget (optional)"),
			"limit":    tools.NumberProperty("New spending limit in dollars (optional)"),
			"category": tools.StringProperty("New category (optional)"),
			"end_date": tools.StringProperty("New end date in YYYY-MM-DD format (optional)"),
		})).
		Handler(func(ctx context.Context, toolParams *core.ToolParams) (*core.ToolResult, error) {
			var params struct {
				ID       string  `json:"id"`
				Name     string  `json:"name"`
				NewName  string  `json:"new_name"`
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

			if params.ID == "" && params.Name == "" {
				return &core.ToolResult{
					Success: false,
					Error:   "Either 'id' or 'name' must be provided to identify the budget",
				}, nil
			}

			// Build update query dynamically
			updates := []string{}
			args := []interface{}{}

			if params.NewName != "" {
				updates = append(updates, "name = ?")
				args = append(args, params.NewName)
			}
			if params.Limit > 0 {
				updates = append(updates, "limit_amount = ?")
				args = append(args, params.Limit)
			}
			if params.Category != "" {
				updates = append(updates, "category = ?")
				args = append(args, params.Category)
			}
			if params.EndDate != "" {
				// Validate date format
				if _, err := time.Parse("2006-01-02", params.EndDate); err != nil {
					return &core.ToolResult{
						Success: false,
						Error:   "Invalid end_date format, use YYYY-MM-DD",
					}, nil
				}
				updates = append(updates, "end_date = ?")
				args = append(args, params.EndDate)
			}

			if len(updates) == 0 {
				return &core.ToolResult{
					Success: false,
					Error:   "No fields to update. Provide at least one of: new_name, limit, category, end_date",
				}, nil
			}

			// Add WHERE clause parameters
			args = append(args, toolParams.UserID)
			if params.ID != "" {
				args = append(args, params.ID)
			} else {
				args = append(args, params.Name)
			}

			// Build complete query
			query := fmt.Sprintf("UPDATE budgets SET %s WHERE user_id = ? AND ", joinStrings(updates, ", "))
			if params.ID != "" {
				query += "id = ?"
			} else {
				query += "name = ?"
			}

			result, err := db.Exec(query, args...)
			if err != nil {
				return &core.ToolResult{
					Success: false,
					Error:   fmt.Sprintf("Failed to update budget: %v", err),
				}, nil
			}

			rowsAffected, _ := result.RowsAffected()
			if rowsAffected == 0 {
				return &core.ToolResult{
					Success: false,
					Error:   "No budget found with the provided identifier. Use get_budgets to see your budgets.",
				}, nil
			}

			// Notify dashboard of update
			NotifyDashboardUpdate("budget", "updated")

			return &core.ToolResult{
				Success: true,
				Data: map[string]interface{}{
					"message":       "Budget updated successfully",
					"rows_affected": rowsAffected,
				},
			}, nil
		}).
		Build()
}

// ============================================================================
// CUSTOM TOOL: DELETE BUDGET
// ============================================================================
// Allows users to delete a budget

func createDeleteBudgetTool() core.Tool {
	return tools.New("delete_budget").
		Description("Delete a budget permanently. Provide either the budget ID or the exact name of the budget to remove it.").
		Schema(tools.ObjectSchema(map[string]interface{}{
			"id":   tools.StringProperty("Budget ID (optional if name is provided)"),
			"name": tools.StringProperty("Exact budget name (optional if id is provided)"),
		})).
		Handler(func(ctx context.Context, toolParams *core.ToolParams) (*core.ToolResult, error) {
			var params struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			}
			if err := json.Unmarshal(toolParams.Input, &params); err != nil {
				return &core.ToolResult{
					Success: false,
					Error:   fmt.Sprintf("invalid input: %v", err),
				}, nil
			}

			if params.ID == "" && params.Name == "" {
				return &core.ToolResult{
					Success: false,
					Error:   "Either 'id' or 'name' must be provided",
				}, nil
			}

			var result interface{}
			var err error

			if params.ID != "" {
				result, err = db.Exec("DELETE FROM budgets WHERE user_id = ? AND id = ?", toolParams.UserID, params.ID)
			} else {
				result, err = db.Exec("DELETE FROM budgets WHERE user_id = ? AND name = ?", toolParams.UserID, params.Name)
			}

			if err != nil {
				return &core.ToolResult{
					Success: false,
					Error:   fmt.Sprintf("Failed to delete budget: %v", err),
				}, nil
			}

			sqlResult, ok := result.(interface{ RowsAffected() (int64, error) })
			if !ok {
				return &core.ToolResult{
					Success: false,
					Error:   "Failed to get deletion result",
				}, nil
			}

			rowsAffected, _ := sqlResult.RowsAffected()
			if rowsAffected == 0 {
				return &core.ToolResult{
					Success: false,
					Error:   "No budget found with the provided identifier. Use get_budgets to see your budgets.",
				}, nil
			}

			// Notify dashboard of update
			NotifyDashboardUpdate("budget", "deleted")

			return &core.ToolResult{
				Success: true,
				Data: map[string]interface{}{
					"message":       "Budget deleted successfully",
					"rows_affected": rowsAffected,
				},
			}, nil
		}).
		Build()
}

// Helper function to join strings
func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
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
