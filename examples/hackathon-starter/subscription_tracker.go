package main

import (
	"context"
	"fmt"
	"time"

	"github.com/becomeliminal/nim-go-sdk/core"
	"github.com/becomeliminal/nim-go-sdk/tools"
)

// Subscription represents a recurring payment
type Subscription struct {
	Name            string  `json:"name"`
	Amount          float64 `json:"amount"`
	Currency        string  `json:"currency"`
	Frequency       string  `json:"frequency"`
	LastPaymentDate string  `json:"last_payment_date"`
	NextPaymentDate string  `json:"next_payment_date"`
	DaysUntilNext   int     `json:"days_until_next"`
}

// createSubscriptionTrackerTool creates a tool that displays mock subscription data
func createSubscriptionTrackerTool(liminalExecutor core.ToolExecutor) core.Tool {
	return tools.New("track_subscriptions").
		Description("Track all recurring subscription payments. Shows subscription details, payment frequency, and alerts for upcoming payments due within 7 days.").
		Schema(tools.ObjectSchema(map[string]interface{}{})).
		Handler(func(ctx context.Context, toolParams *core.ToolParams) (*core.ToolResult, error) {
			// Get mock subscription data
			subscriptions := getMockSubscriptions()

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
			result := map[string]interface{}{
				"subscriptions": subscriptions,
				"summary": map[string]interface{}{
					"total_subscriptions":     len(subscriptions),
					"estimated_monthly_cost":  fmt.Sprintf("%.2f", estimatedMonthlyCost),
					"upcoming_count":          len(upcomingSubscriptions),
					"upcoming_subscriptions":  upcomingSubscriptions,
				},
				"generated_at": time.Now().Format(time.RFC3339),
			}

			return &core.ToolResult{
				Success: true,
				Data:    result,
			}, nil
		}).
		Build()
}

// getMockSubscriptions returns mock subscription data with varied frequencies and amounts
func getMockSubscriptions() []Subscription {
	return []Subscription{
		{
			Name:            "Netflix subscription",
			Amount:          15.99,
			Currency:        "USDC",
			Frequency:       "monthly",
			LastPaymentDate: "2026-01-05",
		},
		{
			Name:            "Spotify subscription",
			Amount:          9.99,
			Currency:        "USDC",
			Frequency:       "monthly",
			LastPaymentDate: "2026-01-28",
		},
		{
			Name:            "Adobe Creative Cloud subscription",
			Amount:          52.99,
			Currency:        "USDC",
			Frequency:       "monthly",
			LastPaymentDate: "2026-01-15",
		},
		{
			Name:            "GitHub Pro subscription",
			Amount:          4.00,
			Currency:        "USDC",
			Frequency:       "monthly",
			LastPaymentDate: "2026-01-10",
		},
		{
			Name:            "The New York Times subscription",
			Amount:          4.25,
			Currency:        "USDC",
			Frequency:       "weekly",
			LastPaymentDate: "2026-01-27",
		},
		{
			Name:            "Gym membership subscription",
			Amount:          45.00,
			Currency:        "USDC",
			Frequency:       "monthly",
			LastPaymentDate: "2026-01-01",
		},
		{
			Name:            "Amazon Prime subscription",
			Amount:          139.00,
			Currency:        "USDC",
			Frequency:       "yearly",
			LastPaymentDate: "2025-03-15",
		},
		{
			Name:            "LinkedIn Premium subscription",
			Amount:          29.99,
			Currency:        "LIL",
			Frequency:       "monthly",
			LastPaymentDate: "2026-01-20",
		},
		{
			Name:            "Notion subscription",
			Amount:          8.00,
			Currency:        "USDC",
			Frequency:       "monthly",
			LastPaymentDate: "2026-01-25",
		},
		{
			Name:            "Cloud storage subscription",
			Amount:          1.99,
			Currency:        "LIL",
			Frequency:       "weekly",
			LastPaymentDate: "2026-01-30",
		},
	}
}
