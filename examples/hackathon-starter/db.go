package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

// InitDB initializes the SQLite database and creates tables
func InitDB(dbPath string) error {
	var err error
	db, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	// Create subscriptions table
	createSubscriptionsTableSQL := `
	CREATE TABLE IF NOT EXISTS subscriptions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		amount REAL NOT NULL,
		currency TEXT NOT NULL DEFAULT 'USDC',
		frequency TEXT NOT NULL CHECK(frequency IN ('weekly', 'monthly', 'yearly')),
		last_payment_date TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	if _, err := db.Exec(createSubscriptionsTableSQL); err != nil {
		return fmt.Errorf("failed to create subscriptions table: %w", err)
	}

	// Create savings_goals table
	createSavingsGoalsTableSQL := `
	CREATE TABLE IF NOT EXISTS savings_goals (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id TEXT NOT NULL,
		name TEXT NOT NULL,
		target_amount REAL NOT NULL,
		current_amount REAL NOT NULL DEFAULT 0,
		category TEXT,
		goal_type TEXT NOT NULL CHECK(goal_type IN ('savings', 'spending_limit')),
		deadline TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		is_completed INTEGER NOT NULL DEFAULT 0
	);`

	if _, err := db.Exec(createSavingsGoalsTableSQL); err != nil {
		return fmt.Errorf("failed to create savings_goals table: %w", err)
	}

	// Create index on user_id for faster queries
	createIndexSQL := `CREATE INDEX IF NOT EXISTS idx_savings_goals_user_id ON savings_goals(user_id);`
	if _, err := db.Exec(createIndexSQL); err != nil {
		return fmt.Errorf("failed to create index: %w", err)
	}

	// Create budgets table
	createBudgetsTableSQL := `
	CREATE TABLE IF NOT EXISTS budgets (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id TEXT NOT NULL,
		name TEXT NOT NULL,
		limit_amount REAL NOT NULL,
		category TEXT,
		start_date TEXT NOT NULL,
		end_date TEXT NOT NULL,
		is_active INTEGER NOT NULL DEFAULT 1,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	if _, err := db.Exec(createBudgetsTableSQL); err != nil {
		return fmt.Errorf("failed to create budgets table: %w", err)
	}

	// Create index on user_id for budgets
	createBudgetIndexSQL := `CREATE INDEX IF NOT EXISTS idx_budgets_user_id ON budgets(user_id);`
	if _, err := db.Exec(createBudgetIndexSQL); err != nil {
		return fmt.Errorf("failed to create budgets index: %w", err)
	}

	// Create transactions table
	createTransactionsTableSQL := `
	CREATE TABLE IF NOT EXISTS transactions (
		id TEXT PRIMARY KEY,
		user_id TEXT NOT NULL,
		amount TEXT NOT NULL,
		counterparty TEXT NOT NULL,
		created_at TEXT NOT NULL,
		currency TEXT NOT NULL,
		direction TEXT NOT NULL CHECK(direction IN ('credit', 'debit')),
		note TEXT,
		status TEXT NOT NULL CHECK(status IN ('confirmed', 'failed', 'pending')),
		tx_hash TEXT,
		type TEXT NOT NULL,
		usd_value TEXT NOT NULL
	);`

	if _, err := db.Exec(createTransactionsTableSQL); err != nil {
		return fmt.Errorf("failed to create transactions table: %w", err)
	}

	// Create index on user_id for transactions
	createTxIndexSQL := `CREATE INDEX IF NOT EXISTS idx_transactions_user_id ON transactions(user_id);`
	if _, err := db.Exec(createTxIndexSQL); err != nil {
		return fmt.Errorf("failed to create transactions index: %w", err)
	}

	// Create index on created_at for sorting
	createTxDateIndexSQL := `CREATE INDEX IF NOT EXISTS idx_transactions_created_at ON transactions(created_at);`
	if _, err := db.Exec(createTxDateIndexSQL); err != nil {
		return fmt.Errorf("failed to create transactions date index: %w", err)
	}

	log.Println("✅ Database initialized successfully")
	return nil
}

// CloseDB closes the database connection
func CloseDB() {
	if db != nil {
		db.Close()
	}
}

// SeedSubscriptions adds initial subscription data (optional)
func SeedSubscriptions() error {
	// Check if we already have data
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM subscriptions").Scan(&count)
	if err != nil {
		return err
	}

	if count > 0 {
		log.Println("📊 Database already contains subscriptions, skipping seed")
		return nil
	}

	log.Println("🌱 Seeding initial subscription data...")

	subscriptions := []struct {
		Name            string
		Amount          float64
		Currency        string
		Frequency       string
		LastPaymentDate string
	}{
		{"Netflix subscription", 15.99, "USDC", "monthly", "2026-01-05"},
		{"Spotify subscription", 9.99, "USDC", "monthly", "2026-01-28"},
		{"Adobe Creative Cloud subscription", 52.99, "USDC", "monthly", "2026-01-15"},
		{"GitHub Pro subscription", 4.00, "USDC", "monthly", "2026-01-10"},
		{"The New York Times subscription", 4.25, "USDC", "weekly", "2026-01-27"},
		{"Gym membership subscription", 45.00, "USDC", "monthly", "2026-01-01"},
		{"Amazon Prime subscription", 139.00, "USDC", "yearly", "2025-03-15"},
		{"LinkedIn Premium subscription", 29.99, "LIL", "monthly", "2026-01-20"},
		{"Notion subscription", 8.00, "USDC", "monthly", "2026-01-25"},
		{"Cloud storage subscription", 1.99, "LIL", "weekly", "2026-01-30"},
	}

	for _, sub := range subscriptions {
		_, err := db.Exec(
			"INSERT INTO subscriptions (name, amount, currency, frequency, last_payment_date) VALUES (?, ?, ?, ?, ?)",
			sub.Name, sub.Amount, sub.Currency, sub.Frequency, sub.LastPaymentDate,
		)
		if err != nil {
			return fmt.Errorf("failed to seed subscription %s: %w", sub.Name, err)
		}
	}

	log.Println("✅ Seeded initial subscription data")
	return nil
}

// SeedTransactions adds initial transaction data (optional)
func SeedTransactions() error {
	// Check if we already have data
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM transactions").Scan(&count)
	if err != nil {
		return err
	}

	if count > 0 {
		log.Println("📊 Database already contains transactions, skipping seed")
		return nil
	}

	log.Println("🌱 Seeding initial transaction data...")

	// Default user_id for demo purposes
	userID := "demo_user"

	transactions := []struct {
		ID           string
		Amount       string
		Counterparty string
		CreatedAt    string
		Currency     string
		Direction    string
		Note         string
		Status       string
		TxHash       string
		Type         string
		UsdValue     string
	}{
		{"019c1653-6519-7e54-aaf2-887d915dc9d0", "0.5", "019c14bb-480d-7f34-b2f7-0ee3dbe478cd", "2026-01-31T23:11:30Z", "LIL", "credit", "", "confirmed", "", "p2p", "0.5"},
		{"019c15ef-d250-7dff-858f-d7e664914b58", "1", "019c14bb-480d-7f34-b2f7-0ee3dbe478cd", "2026-01-31T21:22:44Z", "LIL", "credit", "Optional reason for request", "confirmed", "", "p2p", "1"},
		{"019c15ef-a521-7be0-9842-f394387770c2", "1", "019c14bb-480d-7f34-b2f7-0ee3dbe478cd", "2026-01-31T21:22:32Z", "USD", "credit", "Optional reason for request", "failed", "", "p2p", "1"},
		{"019c15e9-a0f2-7f28-8955-8e9baac8b594", "-4.52", "019c14bb-480d-7f34-b2f7-0ee3dbe478cd", "2026-01-31T21:15:58Z", "LIL", "debit", "dinner bill", "confirmed", "", "p2p", "-4.52"},
		{"019c15e6-841d-74f1-a5d7-d5e808bf0dff", "0.10", "019c14bb-480d-7f34-b2f7-0ee3dbe478cd", "2026-01-31T21:12:34Z", "LIL", "credit", "", "confirmed", "", "p2p", "0.10"},
		{"019c15e5-989e-7a5f-83cf-fd553c903046", "0.20", "019c14bb-480d-7f34-b2f7-0ee3dbe478cd", "2026-01-31T21:11:34Z", "USD", "credit", "", "failed", "", "p2p", "0.20"},
		{"019c15e5-0e0c-7472-a4c1-a0704f3b9b88", "1.00", "019c14bb-480d-7f34-b2f7-0ee3dbe478cd", "2026-01-31T21:10:58Z", "LIL", "credit", "", "confirmed", "", "p2p", "1.00"},
		{"019c1561-2e48-7186-8b52-f1418decd999", "1", "d78133db-b0a9-48a3-b404-089d8efe4149", "2026-01-31T18:46:56Z", "LIL", "credit", "", "confirmed", "", "p2p", "1"},
		{"019c14ec-de2a-789f-8499-4a3597b712a3", "0.31", "019c14bb-480d-7f34-b2f7-0ee3dbe478cd", "2026-01-31T16:39:53Z", "LIL", "credit", "", "confirmed", "", "p2p", "0.31"},
		{"019c14ec-b940-77a7-84ea-6170475b0e79", "0.10", "019c14bb-480d-7f34-b2f7-0ee3dbe478cd", "2026-01-31T16:39:44Z", "LIL", "credit", "", "confirmed", "", "p2p", "0.10"},
	}

	for _, tx := range transactions {
		_, err := db.Exec(
			`INSERT INTO transactions (id, user_id, amount, counterparty, created_at, currency, direction, note, status, tx_hash, type, usd_value) 
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			tx.ID, userID, tx.Amount, tx.Counterparty, tx.CreatedAt, tx.Currency, tx.Direction, tx.Note, tx.Status, tx.TxHash, tx.Type, tx.UsdValue,
		)
		if err != nil {
			return fmt.Errorf("failed to seed transaction %s: %w", tx.ID, err)
		}
	}

	log.Println("✅ Seeded initial transaction data")
	return nil
}
