package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
)

// ============================================================================
// DASHBOARD API ENDPOINTS
// ============================================================================
// REST API endpoints for the dashboard to fetch data from the database

// SubscriptionDTO represents a subscription record for API response
type SubscriptionDTO struct {
	ID              int     `json:"id"`
	Name            string  `json:"name"`
	Amount          float64 `json:"amount"`
	Currency        string  `json:"currency"`
	Frequency       string  `json:"frequency"`
	LastPaymentDate string  `json:"last_payment_date"`
	CreatedAt       string  `json:"created_at"`
}

// Transaction represents a transaction record
type Transaction struct {
	ID           string `json:"id"`
	UserID       string `json:"user_id"`
	Amount       string `json:"amount"`
	Counterparty string `json:"counterparty"`
	CreatedAt    string `json:"created_at"`
	Currency     string `json:"currency"`
	Direction    string `json:"direction"`
	Note         string `json:"note"`
	Status       string `json:"status"`
	TxHash       string `json:"tx_hash"`
	Type         string `json:"type"`
	UsdValue     string `json:"usd_value"`
}

// SavingsGoalDB represents a savings goal from the database
type SavingsGoalDB struct {
	ID            int     `json:"id"`
	UserID        string  `json:"user_id"`
	Name          string  `json:"name"`
	TargetAmount  float64 `json:"target_amount"`
	CurrentAmount float64 `json:"current_amount"`
	Category      string  `json:"category"`
	GoalType      string  `json:"goal_type"`
	Deadline      string  `json:"deadline"`
	CreatedAt     string  `json:"created_at"`
	IsCompleted   bool    `json:"is_completed"`
}

// BudgetDB represents a budget from the database
type BudgetDB struct {
	ID          int     `json:"id"`
	UserID      string  `json:"user_id"`
	Name        string  `json:"name"`
	LimitAmount float64 `json:"limit_amount"`
	Category    string  `json:"category"`
	StartDate   string  `json:"start_date"`
	EndDate     string  `json:"end_date"`
	IsActive    bool    `json:"is_active"`
	CreatedAt   string  `json:"created_at"`
}

// DashboardData represents all dashboard data
type DashboardData struct {
	Subscriptions []SubscriptionDTO `json:"subscriptions"`
	Transactions  []Transaction     `json:"transactions"`
	SavingsGoals  []SavingsGoalDB   `json:"savings_goals"`
	Budgets       []BudgetDB        `json:"budgets"`
	Summary       DashboardSummary  `json:"summary"`
}

// DashboardSummary represents summary statistics
type DashboardSummary struct {
	TotalSubscriptions      int     `json:"total_subscriptions"`
	MonthlySubscriptionCost float64 `json:"monthly_subscription_cost"`
	TotalTransactions       int     `json:"total_transactions"`
	TotalSpent              float64 `json:"total_spent"`
	TotalReceived           float64 `json:"total_received"`
	ActiveGoals             int     `json:"active_goals"`
	CompletedGoals          int     `json:"completed_goals"`
	ActiveBudgets           int     `json:"active_budgets"`
}

// RegisterDashboardRoutes registers the dashboard API routes
func RegisterDashboardRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/dashboard", corsMiddleware(handleDashboard))
	mux.HandleFunc("/api/subscriptions", corsMiddleware(handleSubscriptions))
	mux.HandleFunc("/api/transactions", corsMiddleware(handleTransactions))
	mux.HandleFunc("/api/savings-goals", corsMiddleware(handleSavingsGoals))
	mux.HandleFunc("/api/budgets", corsMiddleware(handleBudgets))
}

// corsMiddleware adds CORS headers
func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}

// handleDashboard returns all dashboard data
func handleDashboard(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	subscriptions, err := getSubscriptions()
	if err != nil {
		log.Printf("Error fetching subscriptions: %v", err)
		subscriptions = []SubscriptionDTO{}
	}

	transactions, err := getTransactions(50)
	if err != nil {
		log.Printf("Error fetching transactions: %v", err)
		transactions = []Transaction{}
	}

	savingsGoals, err := getSavingsGoalsFromDB("")
	if err != nil {
		log.Printf("Error fetching savings goals: %v", err)
		savingsGoals = []SavingsGoalDB{}
	}

	budgets, err := getBudgetsFromDB("")
	if err != nil {
		log.Printf("Error fetching budgets: %v", err)
		budgets = []BudgetDB{}
	}

	summary := calculateSummary(subscriptions, transactions, savingsGoals, budgets)

	data := DashboardData{
		Subscriptions: subscriptions,
		Transactions:  transactions,
		SavingsGoals:  savingsGoals,
		Budgets:       budgets,
		Summary:       summary,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

// handleSubscriptions returns subscriptions
func handleSubscriptions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	subscriptions, err := getSubscriptions()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(subscriptions)
}

// handleTransactions returns transactions
func handleTransactions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	limit := 50
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil {
			limit = parsed
		}
	}

	transactions, err := getTransactions(limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(transactions)
}

// handleSavingsGoals returns savings goals
func handleSavingsGoals(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := r.URL.Query().Get("user_id")
	goals, err := getSavingsGoalsFromDB(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(goals)
}

// handleBudgets returns budgets
func handleBudgets(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := r.URL.Query().Get("user_id")
	budgets, err := getBudgetsFromDB(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(budgets)
}

// Database query functions
func getSubscriptions() ([]SubscriptionDTO, error) {
	rows, err := db.Query(`
		SELECT id, name, amount, currency, frequency, last_payment_date, 
		       COALESCE(created_at, '') as created_at
		FROM subscriptions 
		ORDER BY amount DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subscriptions []SubscriptionDTO
	for rows.Next() {
		var s SubscriptionDTO
		if err := rows.Scan(&s.ID, &s.Name, &s.Amount, &s.Currency, &s.Frequency, &s.LastPaymentDate, &s.CreatedAt); err != nil {
			return nil, err
		}
		subscriptions = append(subscriptions, s)
	}

	if subscriptions == nil {
		subscriptions = []SubscriptionDTO{}
	}
	return subscriptions, nil
}

func getTransactions(limit int) ([]Transaction, error) {
	rows, err := db.Query(`
		SELECT id, user_id, amount, counterparty, created_at, currency, 
		       direction, COALESCE(note, '') as note, status, 
		       COALESCE(tx_hash, '') as tx_hash, type, usd_value
		FROM transactions 
		ORDER BY created_at DESC 
		LIMIT ?
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []Transaction
	for rows.Next() {
		var t Transaction
		if err := rows.Scan(&t.ID, &t.UserID, &t.Amount, &t.Counterparty, &t.CreatedAt,
			&t.Currency, &t.Direction, &t.Note, &t.Status, &t.TxHash, &t.Type, &t.UsdValue); err != nil {
			return nil, err
		}
		transactions = append(transactions, t)
	}

	if transactions == nil {
		transactions = []Transaction{}
	}
	return transactions, nil
}

func getSavingsGoalsFromDB(userID string) ([]SavingsGoalDB, error) {
	query := `
		SELECT id, user_id, name, target_amount, current_amount, 
		       COALESCE(category, '') as category, goal_type, deadline, 
		       COALESCE(created_at, '') as created_at, is_completed
		FROM savings_goals
	`
	if userID != "" {
		query += " WHERE user_id = ?"
	}
	query += " ORDER BY created_at DESC"

	var rows *sql.Rows
	var err error
	if userID != "" {
		rows, err = db.Query(query, userID)
	} else {
		rows, err = db.Query(query)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var goals []SavingsGoalDB
	for rows.Next() {
		var g SavingsGoalDB
		var isCompleted int
		if err := rows.Scan(&g.ID, &g.UserID, &g.Name, &g.TargetAmount, &g.CurrentAmount,
			&g.Category, &g.GoalType, &g.Deadline, &g.CreatedAt, &isCompleted); err != nil {
			return nil, err
		}
		g.IsCompleted = isCompleted == 1
		goals = append(goals, g)
	}

	if goals == nil {
		goals = []SavingsGoalDB{}
	}
	return goals, nil
}

func getBudgetsFromDB(userID string) ([]BudgetDB, error) {
	query := `
		SELECT id, user_id, name, limit_amount, COALESCE(category, '') as category,
		       start_date, end_date, is_active, COALESCE(created_at, '') as created_at
		FROM budgets
	`
	if userID != "" {
		query += " WHERE user_id = ?"
	}
	query += " ORDER BY created_at DESC"

	var rows *sql.Rows
	var err error
	if userID != "" {
		rows, err = db.Query(query, userID)
	} else {
		rows, err = db.Query(query)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var budgets []BudgetDB
	for rows.Next() {
		var b BudgetDB
		var isActive int
		if err := rows.Scan(&b.ID, &b.UserID, &b.Name, &b.LimitAmount, &b.Category,
			&b.StartDate, &b.EndDate, &isActive, &b.CreatedAt); err != nil {
			return nil, err
		}
		b.IsActive = isActive == 1
		budgets = append(budgets, b)
	}

	if budgets == nil {
		budgets = []BudgetDB{}
	}
	return budgets, nil
}

func calculateSummary(subs []SubscriptionDTO, txs []Transaction, goals []SavingsGoalDB, budgets []BudgetDB) DashboardSummary {
	var monthlySubCost float64
	for _, s := range subs {
		switch s.Frequency {
		case "weekly":
			monthlySubCost += s.Amount * 4.33
		case "monthly":
			monthlySubCost += s.Amount
		case "yearly":
			monthlySubCost += s.Amount / 12
		}
	}

	var totalSpent, totalReceived float64
	for _, t := range txs {
		amount, _ := strconv.ParseFloat(t.Amount, 64)
		if t.Direction == "debit" {
			totalSpent += -amount // debit amounts are negative
		} else {
			totalReceived += amount
		}
	}

	var activeGoals, completedGoals int
	for _, g := range goals {
		if g.IsCompleted {
			completedGoals++
		} else {
			activeGoals++
		}
	}

	var activeBudgets int
	for _, b := range budgets {
		if b.IsActive {
			activeBudgets++
		}
	}

	return DashboardSummary{
		TotalSubscriptions:      len(subs),
		MonthlySubscriptionCost: monthlySubCost,
		TotalTransactions:       len(txs),
		TotalSpent:              totalSpent,
		TotalReceived:           totalReceived,
		ActiveGoals:             activeGoals,
		CompletedGoals:          completedGoals,
		ActiveBudgets:           activeBudgets,
	}
}
