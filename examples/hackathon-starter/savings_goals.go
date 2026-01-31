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
// SAVINGS GOALS DATA STRUCTURES
// ============================================================================
// SavingsGoal represents a user's savings or spending goal with optional category
type SavingsGoal struct {
	ID            string    `json:"id"`
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

// Global storage for savings goals (keyed by userID -> slice of goals)
var (
	savingsGoals    = make(map[string][]SavingsGoal)
	savingsGoalsMu  sync.RWMutex
	goalIDCounter   int
	goalIDCounterMu sync.Mutex
)

// generateGoalID creates a unique goal ID
func generateGoalID() string {
	goalIDCounterMu.Lock()
	defer goalIDCounterMu.Unlock()
	goalIDCounter++
	return fmt.Sprintf("goal_%d", goalIDCounter)
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

			// Create the goal
			goal := SavingsGoal{
				ID:            generateGoalID(),
				UserID:        toolParams.UserID,
				Name:          params.Name,
				TargetAmount:  params.TargetAmount,
				CurrentAmount: 0,
				Category:      params.Category,
				GoalType:      params.GoalType,
				Deadline:      time.Now().AddDate(0, 0, params.DeadlineDays),
				CreatedAt:     time.Now(),
				IsCompleted:   false,
			}

			// Store the goal
			savingsGoalsMu.Lock()
			savingsGoals[toolParams.UserID] = append(savingsGoals[toolParams.UserID], goal)
			savingsGoalsMu.Unlock()

			result := map[string]interface{}{
				"success":        true,
				"message":        fmt.Sprintf("Goal '%s' created successfully!", goal.Name),
				"goal_id":        goal.ID,
				"name":           goal.Name,
				"target_amount":  fmt.Sprintf("$%.2f", goal.TargetAmount),
				"goal_type":      goal.GoalType,
				"category":       goal.Category,
				"deadline":       goal.Deadline.Format("January 2, 2006"),
				"days_remaining": params.DeadlineDays,
			}

			return &core.ToolResult{
				Success: true,
				Data:    result,
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

			// Retrieve user's goals
			savingsGoalsMu.RLock()
			userGoals := savingsGoals[toolParams.UserID]
			savingsGoalsMu.RUnlock()

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

			savingsGoalsMu.Lock()
			defer savingsGoalsMu.Unlock()

			userGoals := savingsGoals[toolParams.UserID]
			var foundGoal *SavingsGoal
			var foundIndex int

			// Simple search for goal by name
			for i := range userGoals {
				if userGoals[i].Name == params.GoalName ||
					(len(userGoals[i].Name) > 3 && len(params.GoalName) > 3 &&
						(contains(userGoals[i].Name, params.GoalName) || contains(params.GoalName, userGoals[i].Name))) {
					foundGoal = &userGoals[i]
					foundIndex = i
					break
				}
			}

			if foundGoal == nil {
				return &core.ToolResult{
					Success: false,
					Error:   fmt.Sprintf("Goal '%s' not found. Please verify the goal name from your list of goals.", params.GoalName),
				}, nil
			}

			// Update amount
			foundGoal.CurrentAmount += params.Amount

			// Check for completion
			justCompleted := false
			if foundGoal.GoalType == "savings" && foundGoal.CurrentAmount >= foundGoal.TargetAmount && !foundGoal.IsCompleted {
				foundGoal.IsCompleted = true
				justCompleted = true
			}

			// Save back to slice
			savingsGoals[toolParams.UserID][foundIndex] = *foundGoal

			message := fmt.Sprintf("Updated '%s'. New balance: $%.2f / $%.2f",
				foundGoal.Name, foundGoal.CurrentAmount, foundGoal.TargetAmount)

			if justCompleted {
				message += " 🎉 CONGRATULATIONS! You've reached your goal! 🎉"
			} else if foundGoal.GoalType == "spending_limit" && foundGoal.CurrentAmount > foundGoal.TargetAmount {
				message += " ⚠️ Alert: You have exceeded your spending limit!"
			}

			return &core.ToolResult{
				Success: true,
				Data: map[string]interface{}{
					"goal_id":       foundGoal.ID,
					"name":          foundGoal.Name,
					"new_amount":    foundGoal.CurrentAmount,
					"target_amount": foundGoal.TargetAmount,
					"progress":      fmt.Sprintf("%.1f%%", (foundGoal.CurrentAmount/foundGoal.TargetAmount)*100),
					"message":       message,
					"is_completed":  foundGoal.IsCompleted,
				},
			}, nil
		}).
		Build()
}

// Helper for simple string containment (case-insensitive would be better in production)
func contains(s1, s2 string) bool {
	return len(s1) >= len(s2) && s1[0:len(s2)] == s2 // Very naive check, but sufficient for mock
}
