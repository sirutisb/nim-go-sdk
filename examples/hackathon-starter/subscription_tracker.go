package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/becomeliminal/nim-go-sdk/core"
	"github.com/becomeliminal/nim-go-sdk/tools"
)

// Subscription represents a recurring payment
type Subscription struct {
	ID              int     `json:"id,omitempty"`
	Name            string  `json:"name"`
	Amount          float64 `json:"amount"`
	Currency        string  `json:"currency"`
	Frequency       string  `json:"frequency"`
	LastPaymentDate string  `json:"last_payment_date"`
	NextPaymentDate string  `json:"next_payment_date,omitempty"`
	DaysUntilNext   int     `json:"days_until_next,omitempty"`
}

// createAddSubscriptionTool creates a tool that adds a new subscription to the database
func createAddSubscriptionTool() core.Tool {
	return tools.New("add_subscription").
		Description("Add a new recurring subscription to track. Requires name, amount, currency (USDC or LIL), frequency (weekly/monthly/yearly), and last payment date (YYYY-MM-DD format).").
		Schema(tools.ObjectSchema(map[string]interface{}{
			"name":              tools.StringProperty("Name of the subscription (e.g., 'Netflix subscription', 'Gym membership')"),
			"amount":            tools.StringProperty("Monthly or recurring payment amount"),
			"currency":          tools.StringProperty("Currency type - USDC or LIL"),
			"frequency":         tools.StringProperty("Payment frequency - weekly, monthly, or yearly"),
			"last_payment_date": tools.StringProperty("Last payment date in YYYY-MM-DD format (e.g., '2026-01-15')"),
		}, "name", "amount", "currency", "frequency", "last_payment_date")).
		Handler(func(ctx context.Context, toolParams *core.ToolParams) (*core.ToolResult, error) {
			// Parse input parameters
			var params struct {
				Name            string `json:"name"`
				Amount          string `json:"amount"`
				Currency        string `json:"currency"`
				Frequency       string `json:"frequency"`
				LastPaymentDate string `json:"last_payment_date"`
			}
			if err := json.Unmarshal(toolParams.Input, &params); err != nil {
				return &core.ToolResult{
					Success: false,
					Error:   fmt.Sprintf("invalid input: %v", err),
				}, nil
			}

			// Parse amount
			var amount float64
			if _, err := fmt.Sscanf(params.Amount, "%f", &amount); err != nil {
				return &core.ToolResult{
					Success: false,
					Error:   fmt.Sprintf("invalid amount format: %v", err),
				}, nil
			}

			// Validate required fields
			if params.Name == "" || amount <= 0 || params.Currency == "" || params.Frequency == "" || params.LastPaymentDate == "" {
				return &core.ToolResult{
					Success: false,
					Error:   "All fields are required: name, amount, currency, frequency, last_payment_date",
				}, nil
			}

			// Validate date format
			if _, err := time.Parse("2006-01-02", params.LastPaymentDate); err != nil {
				return &core.ToolResult{
					Success: false,
					Error:   "Invalid date format. Use YYYY-MM-DD (e.g., '2026-01-15')",
				}, nil
			}

			// Insert into database
			result, err := db.Exec(
				"INSERT INTO subscriptions (name, amount, currency, frequency, last_payment_date) VALUES (?, ?, ?, ?, ?)",
				params.Name, amount, params.Currency, params.Frequency, params.LastPaymentDate,
			)
			if err != nil {
				return &core.ToolResult{
					Success: false,
					Error:   fmt.Sprintf("Failed to add subscription: %v", err),
				}, nil
			}

			id, _ := result.LastInsertId()

			return &core.ToolResult{
				Success: true,
				Data: json.RawMessage(fmt.Sprintf(`{
					"message": "Subscription added successfully",
					"subscription": {
						"id": %d,
						"name": "%s",
						"amount": %.2f,
						"currency": "%s",
						"frequency": "%s",
						"last_payment_date": "%s"
					}
				}`, id, params.Name, amount, params.Currency, params.Frequency, params.LastPaymentDate)),
			}, nil
		}).
		Build()
}

// createRemoveSubscriptionTool creates a tool that removes a subscription from the database
func createRemoveSubscriptionTool() core.Tool {
	return tools.New("remove_subscription").
		Description("Remove a subscription from tracking. Provide either the subscription ID or the exact name of the subscription.").
		Schema(tools.ObjectSchema(map[string]interface{}{
			"id":   tools.StringProperty("Subscription ID (optional if name is provided)"),
			"name": tools.StringProperty("Exact subscription name (optional if id is provided)"),
		})).
		Handler(func(ctx context.Context, toolParams *core.ToolParams) (*core.ToolResult, error) {
			// Parse input parameters
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

			var result sql.Result
			var err error

			if params.ID != "" {
				result, err = db.Exec("DELETE FROM subscriptions WHERE id = ?", params.ID)
			} else {
				result, err = db.Exec("DELETE FROM subscriptions WHERE name = ?", params.Name)
			}

			if err != nil {
				return &core.ToolResult{
					Success: false,
					Error:   fmt.Sprintf("Failed to remove subscription: %v", err),
				}, nil
			}

			rowsAffected, _ := result.RowsAffected()
			if rowsAffected == 0 {
				return &core.ToolResult{
					Success: false,
					Error:   "No subscription found with the provided identifier",
				}, nil
			}

			return &core.ToolResult{
				Success: true,
				Data: json.RawMessage(fmt.Sprintf(`{
					"message": "Subscription removed successfully",
					"rows_affected": %d
				}`, rowsAffected)),
			}, nil
		}).
		Build()
}

// createSubscriptionTrackerTool creates a tool that displays subscription data from the database
func createSubscriptionTrackerTool(liminalExecutor core.ToolExecutor) core.Tool {
	return tools.New("track_subscriptions").
		Description("Track all recurring subscription payments. Shows subscription details, payment frequency, and alerts for upcoming payments due within 7 days.").
		Schema(tools.ObjectSchema(map[string]interface{}{})).
		Handler(func(ctx context.Context, toolParams *core.ToolParams) (*core.ToolResult, error) {
			// Fetch subscriptions from database
			rows, err := db.Query("SELECT id, name, amount, currency, frequency, last_payment_date FROM subscriptions")
			if err != nil {
				return &core.ToolResult{
					Success: false,
					Error:   fmt.Sprintf("Failed to fetch subscriptions: %v", err),
				}, nil
			}
			defer rows.Close()

			var subscriptions []Subscription
			for rows.Next() {
				var sub Subscription
				err := rows.Scan(&sub.ID, &sub.Name, &sub.Amount, &sub.Currency, &sub.Frequency, &sub.LastPaymentDate)
				if err != nil {
					continue
				}
				subscriptions = append(subscriptions, sub)
			}

			// Calculate next payment dates and filter urgent ones
			now := time.Now()
			var upcomingSubscriptions []Subscription
			estimatedMonthlyCost := 0.0

			for i := range subscriptions {
				sub := &subscriptions[i]

				// Parse last payment date
				lastPayment, _ := time.Parse("2006-01-02", sub.LastPaymentDate)

				// Calculate next payment date based on frequency
				var nextPayment time.Time
				switch sub.Frequency {
				case "weekly":
					nextPayment = lastPayment.AddDate(0, 0, 7)
					estimatedMonthlyCost += sub.Amount * 4.33
				case "monthly":
					nextPayment = lastPayment.AddDate(0, 1, 0)
					estimatedMonthlyCost += sub.Amount
				case "yearly":
					nextPayment = lastPayment.AddDate(1, 0, 0)
					estimatedMonthlyCost += sub.Amount / 12
				}

				// Calculate days until next payment
				daysUntil := int(nextPayment.Sub(now).Hours() / 24)
				sub.NextPaymentDate = nextPayment.Format("January 2, 2006")
				sub.DaysUntilNext = daysUntil

				// Check if due within 7 days
				if daysUntil <= 7 && daysUntil >= 0 {
					upcomingSubscriptions = append(upcomingSubscriptions, *sub)
				}
			}

			// Build result
			resultData := map[string]interface{}{
				"subscriptions": subscriptions,
				"summary": map[string]interface{}{
					"total_subscriptions":    len(subscriptions),
					"estimated_monthly_cost": fmt.Sprintf("%.2f", estimatedMonthlyCost),
					"upcoming_count":         len(upcomingSubscriptions),
					"upcoming_subscriptions": upcomingSubscriptions,
				},
				"generated_at": time.Now().Format(time.RFC3339),
			}

			resultJSON, err := json.Marshal(resultData)
			if err != nil {
				return &core.ToolResult{
					Success: false,
					Error:   fmt.Sprintf("Failed to marshal result: %v", err),
				}, nil
			}

			return &core.ToolResult{
				Success: true,
				Data:    json.RawMessage(resultJSON),
			}, nil
		}).
		Build()
}
