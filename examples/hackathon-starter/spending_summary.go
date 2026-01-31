package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/becomeliminal/nim-go-sdk/core"
	"github.com/becomeliminal/nim-go-sdk/tools"
)

type TransactionData struct {
	ID           string `json:"id"`
	Amount       string `json:"amount"`
	Currency     string `json:"currency"`
	Direction    string `json:"direction"`
	Counterparty string `json:"counterparty"`
	Note         string `json:"note"`
	Status       string `json:"status"`
	Type         string `json:"type"`
	CreatedAt    string `json:"createdAt"`
	USDValue     string `json:"usdValue"`
}

type SpendingCategory struct {
	Category string  `json:"category"`
	Count    int     `json:"count"`
	Total    float64 `json:"total"`
	Percent  float64 `json:"percent"`
}

func createSpendingSummaryTool(liminalExecutor core.ToolExecutor) core.Tool {
	return tools.New("summarize_spending").
		Description("Summarize and analyze spending patterns from transaction history. Can filter by time period (all, weekly, monthly) and provides detailed insights including spending by category, trends, and recommendations.").
		Schema(tools.ObjectSchema(map[string]interface{}{
			"period": tools.StringProperty("Time period to analyze: 'all' for all transactions, 'weekly' for last 7 days, 'monthly' for last 30 days (default: 'all')"),
		})).
		Handler(func(ctx context.Context, toolParams *core.ToolParams) (*core.ToolResult, error) {
			var params struct {
				Period string `json:"period"`
			}
			if err := json.Unmarshal(toolParams.Input, &params); err != nil {
				return &core.ToolResult{Success: false, Error: fmt.Sprintf("invalid input: %v", err)}, nil
			}
			if params.Period == "" {
				params.Period = "all"
			}
			if params.Period != "all" && params.Period != "weekly" && params.Period != "monthly" {
				return &core.ToolResult{Success: false, Error: "period must be 'all', 'weekly', or 'monthly'"}, nil
			}

			txRequest := map[string]interface{}{}
			txRequestJSON, _ := json.Marshal(txRequest)
			txResponse, err := liminalExecutor.Execute(ctx, &core.ExecuteRequest{
				UserID: toolParams.UserID, Tool: "get_transactions", Input: txRequestJSON, RequestID: toolParams.RequestID,
			})
			if err != nil {
				return &core.ToolResult{Success: false, Error: fmt.Sprintf("failed to fetch transactions: %v", err)}, nil
			}
			if !txResponse.Success {
				return &core.ToolResult{Success: false, Error: fmt.Sprintf("transaction fetch failed: %s", txResponse.Error)}, nil
			}

			var txData struct{ Transactions []TransactionData `json:"transactions"` }
			if err := json.Unmarshal(txResponse.Data, &txData); err != nil {
				return &core.ToolResult{Success: false, Error: fmt.Sprintf("failed to parse transactions: %v", err)}, nil
			}

			filteredTxs := filterTransactionsByPeriod(txData.Transactions, params.Period)
			analysis := analyzeSpending(filteredTxs, params.Period)
			
			// Compare with previous period
			comparison := comparePeriods(txData.Transactions, params.Period)
			
			result := map[string]interface{}{
				"period": params.Period, 
				"analysis": analysis, 
				"comparison": comparison,
				"generated_at": time.Now().Format(time.RFC3339),
			}
			return &core.ToolResult{Success: true, Data: result}, nil
		}).Build()
}

func filterTransactionsByPeriod(transactions []TransactionData, period string) []TransactionData {
	if period == "all" {
		return transactions
	}
	now := time.Now()
	var cutoff time.Time
	switch period {
	case "weekly":
		cutoff = now.AddDate(0, 0, -7)
	case "monthly":
		cutoff = now.AddDate(0, 0, -30)
	default:
		return transactions
	}
	var filtered []TransactionData
	for _, tx := range transactions {
		txTime, err := time.Parse(time.RFC3339, tx.CreatedAt)
		if err != nil {
			continue
		}
		if txTime.After(cutoff) {
			filtered = append(filtered, tx)
		}
	}
	return filtered
}

func analyzeSpending(transactions []TransactionData, period string) map[string]interface{} {
	if len(transactions) == 0 {
		return map[string]interface{}{"summary": "No transactions found in the specified period"}
	}
	var totalSpent, totalReceived float64
	var spendingTxs, receivingTxs []TransactionData
	categorySpending := make(map[string]float64)
	categoryCount := make(map[string]int)
	currencyBreakdown := make(map[string]float64)

	for _, tx := range transactions {
		if tx.Status != "confirmed" {
			continue
		}
		amount := parseAmountFloat(tx.Amount)
		if tx.Direction == "debit" {
			totalSpent += amount
			spendingTxs = append(spendingTxs, tx)
			category := categorizeTransaction(tx)
			categorySpending[category] += amount
			categoryCount[category]++
			currencyBreakdown[tx.Currency] += amount
		} else if tx.Direction == "credit" {
			totalReceived += amount
			receivingTxs = append(receivingTxs, tx)
		}
	}

	var categories []SpendingCategory
	for cat, total := range categorySpending {
		percent := 0.0
		if totalSpent > 0 {
			percent = (total / totalSpent) * 100
		}
		categories = append(categories, SpendingCategory{Category: cat, Count: categoryCount[cat], Total: total, Percent: percent})
	}
	sort.Slice(categories, func(i, j int) bool { return categories[i].Total > categories[j].Total })

	days := calculateDays(period, transactions)
	avgDailySpending := 0.0
	if days > 0 {
		avgDailySpending = totalSpent / float64(days)
	}
	insights := generateInsights(totalSpent, totalReceived, categories, avgDailySpending, period, len(spendingTxs))

	return map[string]interface{}{
		"summary": map[string]interface{}{
			"total_spent": fmt.Sprintf("%.2f", totalSpent), "total_received": fmt.Sprintf("%.2f", totalReceived),
			"net_cashflow": fmt.Sprintf("%.2f", totalReceived-totalSpent), "spending_count": len(spendingTxs),
			"receiving_count": len(receivingTxs), "avg_daily_spending": fmt.Sprintf("%.2f", avgDailySpending), "days_analyzed": days,
		},
		"categories": categories, "currency_breakdown": currencyBreakdown, "insights": insights, "top_expenses": getTopExpenses(spendingTxs, 5),
	}
}

func categorizeTransaction(tx TransactionData) string {
	note := strings.ToLower(tx.Note)
	if strings.Contains(note, "subscription") {
		return "Subscriptions"
	}
	if strings.Contains(note, "restaurant") || strings.Contains(note, "food") || strings.Contains(note, "coffee") || strings.Contains(note, "lunch") || strings.Contains(note, "dinner") {
		return "Food & Dining"
	}
	if strings.Contains(note, "shop") || strings.Contains(note, "store") || strings.Contains(note, "purchase") {
		return "Shopping"
	}
	if strings.Contains(note, "movie") || strings.Contains(note, "game") || strings.Contains(note, "entertainment") {
		return "Entertainment"
	}
	if strings.Contains(note, "uber") || strings.Contains(note, "lyft") || strings.Contains(note, "taxi") || strings.Contains(note, "transport") {
		return "Transportation"
	}
	if tx.Type == "deposit" || strings.Contains(note, "savings") || strings.Contains(note, "deposit") {
		return "Savings & Investment"
	}
	if strings.Contains(note, "bill") || strings.Contains(note, "utility") || strings.Contains(note, "payment") {
		return "Bills & Utilities"
	}
	if tx.Type == "p2p" {
		return "Transfers & Payments"
	}
	return "Other"
}

func calculateDays(period string, transactions []TransactionData) int {
	if period == "weekly" {
		return 7
	} else if period == "monthly" {
		return 30
	}
	if len(transactions) == 0 {
		return 0
	}
	var dates []time.Time
	for _, tx := range transactions {
		t, err := time.Parse(time.RFC3339, tx.CreatedAt)
		if err == nil {
			dates = append(dates, t)
		}
	}
	if len(dates) == 0 {
		return 0
	}
	sort.Slice(dates, func(i, j int) bool { return dates[i].Before(dates[j]) })
	days := int(dates[len(dates)-1].Sub(dates[0]).Hours() / 24)
	if days == 0 {
		return 1
	}
	return days
}

func getTopExpenses(transactions []TransactionData, limit int) []map[string]interface{} {
	sort.Slice(transactions, func(i, j int) bool {
		return parseAmountFloat(transactions[i].Amount) > parseAmountFloat(transactions[j].Amount)
	})
	var topExpenses []map[string]interface{}
	count := limit
	if len(transactions) < limit {
		count = len(transactions)
	}
	for i := 0; i < count; i++ {
		tx := transactions[i]
		topExpenses = append(topExpenses, map[string]interface{}{
			"amount": parseAmountFloat(tx.Amount), "currency": tx.Currency, "note": tx.Note,
			"date": formatDateShort(tx.CreatedAt), "category": categorizeTransaction(tx),
		})
	}
	return topExpenses
}

func generateInsights(totalSpent, totalReceived float64, categories []SpendingCategory, avgDaily float64, period string, txCount int) []string {
	var insights []string
	netFlow := totalReceived - totalSpent
	if netFlow > 0 {
		insights = append(insights, fmt.Sprintf("✅ Positive cashflow! You received $%.2f more than you spent.", netFlow))
	} else if netFlow < 0 {
		insights = append(insights, fmt.Sprintf("⚠️ Negative cashflow: You spent $%.2f more than you received.", math.Abs(netFlow)))
	}
	periodName := "period"
	if period == "weekly" {
		periodName = "week"
	} else if period == "monthly" {
		periodName = "month"
	}
	insights = append(insights, fmt.Sprintf("You made %d spending transactions this %s, averaging $%.2f per day.", txCount, periodName, avgDaily))
	if len(categories) > 0 {
		topCat := categories[0]
		insights = append(insights, fmt.Sprintf("💰 Your biggest spending category is '%s' at $%.2f (%.1f%% of total spending).", topCat.Category, topCat.Total, topCat.Percent))
	}
	for _, cat := range categories {
		if cat.Category == "Subscriptions" && cat.Total > 0 {
			monthlyEst := cat.Total
			if period == "weekly" {
				monthlyEst = cat.Total * 4.33
			}
			insights = append(insights, fmt.Sprintf("📱 You're spending $%.2f on subscriptions (estimated $%.2f/month).", cat.Total, monthlyEst))
			break
		}
	}
	if totalReceived > totalSpent {
		savingsOpportunity := (totalReceived - totalSpent) * 0.7
		insights = append(insights, fmt.Sprintf("💡 Consider saving $%.2f of your surplus into your savings account to earn interest!", savingsOpportunity))
	}
	return insights
}

func parseAmountFloat(amountStr string) float64 {
	var amount float64
	cleaned := strings.TrimPrefix(amountStr, "-")
	fmt.Sscanf(cleaned, "%f", &amount)
	return amount
}

func formatDateShort(dateStr string) string {
	t, err := time.Parse(time.RFC3339, dateStr)
	if err != nil {
		return dateStr
	}
	return t.Format("Jan 2")
}

func comparePeriods(allTransactions []TransactionData, period string) map[string]interface{} {
	if period == "all" {
		return map[string]interface{}{
			"message": "Period comparison not available for 'all' transactions view",
		}
	}
	
	now := time.Now()
	var currentStart, currentEnd, previousStart, previousEnd time.Time
	
	switch period {
	case "weekly":
		currentEnd = now
		currentStart = now.AddDate(0, 0, -7)
		previousEnd = currentStart
		previousStart = currentStart.AddDate(0, 0, -7)
	case "monthly":
		currentEnd = now
		currentStart = now.AddDate(0, 0, -30)
		previousEnd = currentStart
		previousStart = currentStart.AddDate(0, 0, -30)
	default:
		return map[string]interface{}{"message": "Unknown period"}
	}
	
	// Filter transactions for current period
	var currentTxs []TransactionData
	for _, tx := range allTransactions {
		txTime, err := time.Parse(time.RFC3339, tx.CreatedAt)
		if err != nil {
			continue
		}
		if txTime.After(currentStart) && txTime.Before(currentEnd.Add(time.Hour*24)) {
			currentTxs = append(currentTxs, tx)
		}
	}
	
	// Filter transactions for previous period
	var previousTxs []TransactionData
	for _, tx := range allTransactions {
		txTime, err := time.Parse(time.RFC3339, tx.CreatedAt)
		if err != nil {
			continue
		}
		if txTime.After(previousStart) && txTime.Before(previousEnd.Add(time.Hour*24)) {
			previousTxs = append(previousTxs, tx)
		}
	}
	
	// Calculate spending for both periods
	currentSpent, currentReceived := calculateTotals(currentTxs)
	previousSpent, previousReceived := calculateTotals(previousTxs)
	
	// Calculate changes
	spendingChange := currentSpent - previousSpent
	spendingChangePercent := 0.0
	if previousSpent > 0 {
		spendingChangePercent = (spendingChange / previousSpent) * 100
	}
	
	receivingChange := currentReceived - previousReceived
	receivingChangePercent := 0.0
	if previousReceived > 0 {
		receivingChangePercent = (receivingChange / previousReceived) * 100
	}
	
	currentSavings := currentReceived - currentSpent
	previousSavings := previousReceived - previousSpent
	savingsChange := currentSavings - previousSavings
	
	// Compare categories
	categoryComparison := compareCategorySpending(currentTxs, previousTxs)
	
	// Generate motivational insights
	insights := generateComparisonInsights(
		spendingChange, spendingChangePercent,
		receivingChange, receivingChangePercent,
		savingsChange, currentSavings, previousSavings,
		categoryComparison, period,
	)
	
	return map[string]interface{}{
		"current_period": map[string]interface{}{
			"spent":    fmt.Sprintf("%.2f", currentSpent),
			"received": fmt.Sprintf("%.2f", currentReceived),
			"savings":  fmt.Sprintf("%.2f", currentSavings),
		},
		"previous_period": map[string]interface{}{
			"spent":    fmt.Sprintf("%.2f", previousSpent),
			"received": fmt.Sprintf("%.2f", previousReceived),
			"savings":  fmt.Sprintf("%.2f", previousSavings),
		},
		"changes": map[string]interface{}{
			"spending_change":         fmt.Sprintf("%.2f", spendingChange),
			"spending_change_percent": fmt.Sprintf("%.1f%%", spendingChangePercent),
			"receiving_change":        fmt.Sprintf("%.2f", receivingChange),
			"receiving_change_percent": fmt.Sprintf("%.1f%%", receivingChangePercent),
			"savings_change":          fmt.Sprintf("%.2f", savingsChange),
		},
		"category_comparison": categoryComparison,
		"insights":            insights,
	}
}

func calculateTotals(transactions []TransactionData) (float64, float64) {
	var spent, received float64
	for _, tx := range transactions {
		if tx.Status != "confirmed" {
			continue
		}
		amount := parseAmountFloat(tx.Amount)
		if tx.Direction == "debit" {
			spent += amount
		} else if tx.Direction == "credit" {
			received += amount
		}
	}
	return spent, received
}

func compareCategorySpending(currentTxs, previousTxs []TransactionData) []map[string]interface{} {
	currentCategories := make(map[string]float64)
	previousCategories := make(map[string]float64)
	
	for _, tx := range currentTxs {
		if tx.Status == "confirmed" && tx.Direction == "debit" {
			category := categorizeTransaction(tx)
			currentCategories[category] += parseAmountFloat(tx.Amount)
		}
	}
	
	for _, tx := range previousTxs {
		if tx.Status == "confirmed" && tx.Direction == "debit" {
			category := categorizeTransaction(tx)
			previousCategories[category] += parseAmountFloat(tx.Amount)
		}
	}
	
	// Build comparison list
	allCategories := make(map[string]bool)
	for cat := range currentCategories {
		allCategories[cat] = true
	}
	for cat := range previousCategories {
		allCategories[cat] = true
	}
	
	var comparison []map[string]interface{}
	for cat := range allCategories {
		current := currentCategories[cat]
		previous := previousCategories[cat]
		change := current - previous
		changePercent := 0.0
		if previous > 0 {
			changePercent = (change / previous) * 100
		}
		
		if current > 0 || previous > 0 {
			comparison = append(comparison, map[string]interface{}{
				"category":       cat,
				"current":        fmt.Sprintf("%.2f", current),
				"previous":       fmt.Sprintf("%.2f", previous),
				"change":         fmt.Sprintf("%.2f", change),
				"change_percent": fmt.Sprintf("%.1f%%", changePercent),
			})
		}
	}
	
	// Sort by absolute change (biggest changes first)
	sort.Slice(comparison, func(i, j int) bool {
		changeI := 0.0
		changeJ := 0.0
		fmt.Sscanf(comparison[i]["change"].(string), "%f", &changeI)
		fmt.Sscanf(comparison[j]["change"].(string), "%f", &changeJ)
		return math.Abs(changeI) > math.Abs(changeJ)
	})
	
	return comparison
}

func generateComparisonInsights(spendingChange, spendingChangePercent, receivingChange, receivingChangePercent, savingsChange, currentSavings, previousSavings float64, categoryComparison []map[string]interface{}, period string) []string {
	var insights []string
	
	periodName := "this period"
	previousPeriodName := "last period"
	if period == "weekly" {
		periodName = "this week"
		previousPeriodName = "last week"
	} else if period == "monthly" {
		periodName = "this month"
		previousPeriodName = "last month"
	}
	
	// Spending comparison
	if spendingChange < 0 {
		insights = append(insights, fmt.Sprintf("🎉 Great job! You spent $%.2f (%.1f%%) less %s compared to %s!", math.Abs(spendingChange), math.Abs(spendingChangePercent), periodName, previousPeriodName))
	} else if spendingChange > 0 {
		insights = append(insights, fmt.Sprintf("⚠️ You spent $%.2f (%.1f%%) more %s compared to %s. Let's get back on track!", spendingChange, spendingChangePercent, periodName, previousPeriodName))
	} else {
		insights = append(insights, fmt.Sprintf("Your spending remained consistent between periods."))
	}
	
	// Savings comparison
	if savingsChange > 0 {
		insights = append(insights, fmt.Sprintf("💰 Excellent! Your savings improved by $%.2f compared to %s!", savingsChange, previousPeriodName))
	} else if savingsChange < 0 {
		insights = append(insights, fmt.Sprintf("📉 Your savings decreased by $%.2f. Consider reviewing your spending categories.", math.Abs(savingsChange)))
	}
	
	if currentSavings > 0 && previousSavings <= 0 {
		insights = append(insights, fmt.Sprintf("🌟 Amazing turnaround! You went from negative to positive cashflow!"))
	} else if currentSavings <= 0 && previousSavings > 0 {
		insights = append(insights, fmt.Sprintf("⚠️ You've moved into negative cashflow. Time to review your budget."))
	}
	
	// Income comparison
	if receivingChange > 0 {
		insights = append(insights, fmt.Sprintf("📈 Your income increased by $%.2f (%.1f%%) %s!", receivingChange, receivingChangePercent, periodName))
	} else if receivingChange < 0 {
		insights = append(insights, fmt.Sprintf("Your income decreased by $%.2f (%.1f%%) %s.", math.Abs(receivingChange), math.Abs(receivingChangePercent), periodName))
	}
	
	// Category insights (biggest changes)
	if len(categoryComparison) > 0 {
		topChange := categoryComparison[0]
		var changeAmount float64
		fmt.Sscanf(topChange["change"].(string), "%f", &changeAmount)
		
		if changeAmount > 0 {
			insights = append(insights, fmt.Sprintf("📊 Biggest spending increase: '%s' (+$%.2f)", topChange["category"], changeAmount))
		} else if changeAmount < 0 {
			insights = append(insights, fmt.Sprintf("✅ Biggest spending decrease: '%s' (-$%.2f)", topChange["category"], math.Abs(changeAmount)))
		}
	}
	
	// Motivational message based on overall trend
	if spendingChange < 0 && savingsChange > 0 {
		insights = append(insights, "🏆 You're on a winning streak! Keep up the great financial discipline!")
	} else if spendingChange > 0 && savingsChange < 0 {
		insights = append(insights, "💪 Don't worry! Small adjustments to your budget can get you back on track quickly.")
	}
	
	return insights
}
