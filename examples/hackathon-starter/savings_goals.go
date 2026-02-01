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
// SAVINGS GOALS DATA STRUCTURES
// ============================================================================
// SavingsGoal represents a user's savings or spending goal with optional category
type SavingsGoal struct {
	ID            int       `json:"id"`
	UserID        string    `json:"user_id"`
	Name          string    `json:"name"`
	TargetAmount  float64   `json:"target_amount"`
	CurrentAmount float64   `json:"current_amount"`
	Category      string    `json:"category,omitempty"` // Optional category for future extensibility
	GoalType      string    `json:"goal_type"`          // "savings" or "spending_limit"
	Deadline      time.Time `json:"deadline"`
	CreatedAt     time.Time `json:"created_at"`
	IsCompleted   bool      `json:"is_completed"`
}

// ============================================================================
// CUSTOM TOOL: SET SAVINGS GOAL
// ============================================================================
// Allows users to set savings or spending goals with deadlines and optional categories

func createSetSavingsGoalTool() core.Tool {
	return tools.New("set_savings_goal").
		Description("Set a savings or spending goal for the user. The goal can have an optional category for organization. Supports monthly spending limits or savings targets with deadlines.").
		Schema(tools.ObjectSchema(map[string]interface{}{
			"name":          tools.StringProperty("Name or description of the goal (e.g., 'Monthly spending limit', 'Emergency fund')"),
			"target_amount": tools.NumberProperty("Target amount in dollars for the goal"),
			"goal_type":     tools.StringProperty("Type of goal: 'savings' for saving toward a target, 'spending_limit' for limiting spending"),
			"deadline_days": tools.IntegerProperty("Number of days until the goal deadline (default: 30 for monthly goals)"),
			"category":      tools.StringProperty("Optional category for the goal (e.g., 'groceries', 'entertainment', 'emergency')"),
		})).
		Handler(func(ctx context.Context, toolParams *core.ToolParams) (*core.ToolResult, error) {
			var params struct {
				Name         string  `json:"name"`
				TargetAmount float64 `json:"target_amount"`
				GoalType     string  `json:"goal_type"`
				DeadlineDays int     `json:"deadline_days"`
				Category     string  `json:"category"`
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
			if params.TargetAmount <= 0 {
				return &core.ToolResult{
					Success: false,
					Error:   "target_amount must be greater than 0",
				}, nil
			}

			// Default values
			if params.GoalType == "" {
				params.GoalType = "savings"
			}
			if params.DeadlineDays == 0 {
				params.DeadlineDays = 30 // Default to monthly
			}

			deadline := time.Now().AddDate(0, 0, params.DeadlineDays)

			// Insert into database
			result, err := db.Exec(
				`INSERT INTO savings_goals (user_id, name, target_amount, current_amount, category, goal_type, deadline, is_completed) 
				 VALUES (?, ?, ?, 0, ?, ?, ?, 0)`,
				toolParams.UserID, params.Name, params.TargetAmount, params.Category, params.GoalType, deadline.Format("2006-01-02"),
			)
			if err != nil {
				return &core.ToolResult{
					Success: false,
					Error:   fmt.Sprintf("Failed to create goal: %v", err),
				}, nil
			}

			id, _ := result.LastInsertId()

			responseData := map[string]interface{}{
				"success":        true,
				"message":        fmt.Sprintf("Goal '%s' created successfully!", params.Name),
				"goal_id":        id,
				"name":           params.Name,
				"target_amount":  fmt.Sprintf("$%.2f", params.TargetAmount),
				"goal_type":      params.GoalType,
				"category":       params.Category,
				"deadline":       deadline.Format("January 2, 2006"),
				"days_remaining": params.DeadlineDays,
			}

			return &core.ToolResult{
				Success: true,
				Data:    responseData,
			}, nil
		}).
		Build()
}

// ============================================================================
// CUSTOM TOOL: GET SAVINGS GOALS
// ============================================================================
// Allows users to query their savings goals with progress tracking

func createGetSavingsGoalsTool() core.Tool {
	return tools.New("get_savings_goals").
		Description("Fetch all savings and spending goals for the user. Shows progress, time remaining, and whether goals are on track. Can optionally filter by category.").
		Schema(tools.ObjectSchema(map[string]interface{}{
			"category": tools.StringProperty("Optional category to filter goals by"),
		})).
		Handler(func(ctx context.Context, toolParams *core.ToolParams) (*core.ToolResult, error) {
			var params struct {
				Category string `json:"category"`
			}
			_ = json.Unmarshal(toolParams.Input, &params)

			// Build query with optional category filter
			query := `SELECT id, user_id, name, target_amount, current_amount, category, goal_type, deadline, created_at, is_completed 
					  FROM savings_goals WHERE user_id = ?`
			args := []interface{}{toolParams.UserID}

			if params.Category != "" {
				query += " AND category = ?"
				args = append(args, params.Category)
			}

			rows, err := db.Query(query, args...)
			if err != nil {
				return &core.ToolResult{
					Success: false,
					Error:   fmt.Sprintf("Failed to fetch goals: %v", err),
				}, nil
			}
			defer rows.Close()

			var userGoals []SavingsGoal
			for rows.Next() {
				var goal SavingsGoal
				var deadlineStr, createdAtStr string
				var isCompletedInt int

				err := rows.Scan(&goal.ID, &goal.UserID, &goal.Name, &goal.TargetAmount, &goal.CurrentAmount,
					&goal.Category, &goal.GoalType, &deadlineStr, &createdAtStr, &isCompletedInt)
				if err != nil {
					continue
				}

				goal.Deadline, _ = time.Parse("2006-01-02", deadlineStr)
				goal.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAtStr)
				goal.IsCompleted = isCompletedInt == 1

				userGoals = append(userGoals, goal)
			}

			if len(userGoals) == 0 {
				return &core.ToolResult{
					Success: true,
					Data: map[string]interface{}{
						"message":     "No savings goals found. Use set_savings_goal to create one!",
						"goals":       []interface{}{},
						"total_goals": 0,
					},
				}, nil
			}

			// Filter and format goals
			var formattedGoals []map[string]interface{}
			now := time.Now()

			for _, goal := range userGoals {
				// Filter by category if specified
				if params.Category != "" && goal.Category != params.Category {
					continue
				}

				// Calculate progress and status
				daysRemaining := int(goal.Deadline.Sub(now).Hours() / 24)
				if daysRemaining < 0 {
					daysRemaining = 0
				}

				totalDays := int(goal.Deadline.Sub(goal.CreatedAt).Hours() / 24)
				daysElapsed := totalDays - daysRemaining
				if daysElapsed < 0 {
					daysElapsed = 0
				}

				progressPercent := 0.0
				if goal.TargetAmount > 0 {
					progressPercent = (goal.CurrentAmount / goal.TargetAmount) * 100
				}

				// Determine if on track
				expectedProgress := 0.0
				if totalDays > 0 {
					expectedProgress = (float64(daysElapsed) / float64(totalDays)) * 100
				}

				status := "on_track"
				statusMessage := "You're on track! 🎯"
				if goal.IsCompleted {
					status = "completed"
					statusMessage = "Goal completed! 🎉"
				} else if daysRemaining == 0 {
					status = "expired"
					statusMessage = "Goal deadline has passed ⏰"
				} else if goal.GoalType == "savings" && progressPercent < expectedProgress-10 {
					status = "behind"
					statusMessage = "You're falling behind, consider adding more! 💪"
				} else if goal.GoalType == "spending_limit" && progressPercent > expectedProgress+10 {
					status = "over_budget"
					statusMessage = "You're spending faster than planned! ⚠️"
				}

				formattedGoal := map[string]interface{}{
					"id":               goal.ID,
					"name":             goal.Name,
					"goal_type":        goal.GoalType,
					"category":         goal.Category,
					"target_amount":    fmt.Sprintf("$%.2f", goal.TargetAmount),
					"current_amount":   fmt.Sprintf("$%.2f", goal.CurrentAmount),
					"remaining_amount": fmt.Sprintf("$%.2f", goal.TargetAmount-goal.CurrentAmount),
					"progress_percent": fmt.Sprintf("%.1f%%", progressPercent),
					"deadline":         goal.Deadline.Format("January 2, 2006"),
					"days_remaining":   daysRemaining,
					"status":           status,
					"status_message":   statusMessage,
					"is_completed":     goal.IsCompleted,
				}
				formattedGoals = append(formattedGoals, formattedGoal)
			}

			result := map[string]interface{}{
				"goals":        formattedGoals,
				"total_goals":  len(formattedGoals),
				"retrieved_at": now.Format(time.RFC3339),
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
// CUSTOM TOOL: UPDATE GOAL PROGRESS
// ============================================================================
// Allows users to update the progress of a savings goal or spending limit

func createUpdateGoalProgressTool() core.Tool {
	return tools.New("update_goal_progress").
		Description("Update the current amount of a savings goal or spending limit. Use this when the user saves money or spends from a budget. Positive amounts increase the current balance.").
		Schema(tools.ObjectSchema(map[string]interface{}{
			"goal_name": tools.StringProperty("The name of the goal to update (fuzzy match)"),
			"amount":    tools.NumberProperty("The amount to add to the current progress (use negative values to subtract)"),
		})).
		Handler(func(ctx context.Context, toolParams *core.ToolParams) (*core.ToolResult, error) {
			var params struct {
				GoalName string  `json:"goal_name"`
				Amount   float64 `json:"amount"`
			}
			if err := json.Unmarshal(toolParams.Input, &params); err != nil {
				return &core.ToolResult{
					Success: false,
					Error:   fmt.Sprintf("invalid input: %v", err),
				}, nil
			}

			if params.GoalName == "" {
				return &core.ToolResult{
					Success: false,
					Error:   "goal_name is required",
				}, nil
			}

			// Find the goal by name (with fuzzy matching)
			query := `SELECT id, name, target_amount, current_amount, goal_type, is_completed 
					  FROM savings_goals WHERE user_id = ? AND (name = ? OR name LIKE ?)`
			row := db.QueryRow(query, toolParams.UserID, params.GoalName, "%"+params.GoalName+"%")

			var goal SavingsGoal
			var isCompletedInt int
			err := row.Scan(&goal.ID, &goal.Name, &goal.TargetAmount, &goal.CurrentAmount, &goal.GoalType, &isCompletedInt)
			if err != nil {
				return &core.ToolResult{
					Success: false,
					Error:   fmt.Sprintf("Goal '%s' not found. Please verify the goal name from your list of goals.", params.GoalName),
				}, nil
			}
			goal.IsCompleted = isCompletedInt == 1

			// Calculate new amount
			newAmount := goal.CurrentAmount + params.Amount

			// Check for completion
			justCompleted := false
			newIsCompleted := goal.IsCompleted
			if goal.GoalType == "savings" && newAmount >= goal.TargetAmount && !goal.IsCompleted {
				newIsCompleted = true
				justCompleted = true
			}

			// Update in database
			_, err = db.Exec(`UPDATE savings_goals SET current_amount = ?, is_completed = ? WHERE id = ?`,
				newAmount, newIsCompleted, goal.ID)
			if err != nil {
				return &core.ToolResult{
					Success: false,
					Error:   fmt.Sprintf("Failed to update goal: %v", err),
				}, nil
			}

			message := fmt.Sprintf("Updated '%s'. New balance: $%.2f / $%.2f",
				goal.Name, newAmount, goal.TargetAmount)

			if justCompleted {
				message += " 🎉 CONGRATULATIONS! You've reached your goal! 🎉"
			} else if goal.GoalType == "spending_limit" && newAmount > goal.TargetAmount {
				message += " ⚠️ Alert: You have exceeded your spending limit!"
			}

			return &core.ToolResult{
				Success: true,
				Data: map[string]interface{}{
					"goal_id":       goal.ID,
					"name":          goal.Name,
					"new_amount":    newAmount,
					"target_amount": goal.TargetAmount,
					"progress":      fmt.Sprintf("%.1f%%", (newAmount/goal.TargetAmount)*100),
					"message":       message,
					"is_completed":  newIsCompleted,
				},
			}, nil
		}).
		Build()
}

// ============================================================================
// CUSTOM TOOL: DELETE SAVINGS GOAL
// ============================================================================
// Allows users to delete a savings goal by ID or name

func createDeleteSavingsGoalTool() core.Tool {
	return tools.New("delete_savings_goal").
		Description("Delete a savings or spending goal. Provide either the goal ID or the exact name of the goal to remove it permanently.").
		Schema(tools.ObjectSchema(map[string]interface{}{
			"id":   tools.StringProperty("Goal ID (optional if name is provided)"),
			"name": tools.StringProperty("Exact goal name (optional if id is provided)"),
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
				result, err = db.Exec("DELETE FROM savings_goals WHERE user_id = ? AND id = ?", toolParams.UserID, params.ID)
			} else {
				result, err = db.Exec("DELETE FROM savings_goals WHERE user_id = ? AND name = ?", toolParams.UserID, params.Name)
			}

			if err != nil {
				return &core.ToolResult{
					Success: false,
					Error:   fmt.Sprintf("Failed to delete goal: %v", err),
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
					Error:   "No goal found with the provided identifier. Use get_savings_goals to see your goals.",
				}, nil
			}

			return &core.ToolResult{
				Success: true,
				Data: map[string]interface{}{
					"message":       "Goal deleted successfully",
					"rows_affected": rowsAffected,
				},
			}, nil
		}).
		Build()
}

// Helper for simple string containment (case-insensitive would be better in production)
func contains(s1, s2 string) bool {
	return len(s1) >= len(s2) && s1[0:len(s2)] == s2 // Very naive check, but sufficient for mock
}
