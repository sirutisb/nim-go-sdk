import { useState, useEffect, useCallback } from 'react'
import ReactDOM from 'react-dom/client'
import { NimChat } from '@liminalcash/nim-chat'
import '@liminalcash/nim-chat/styles.css'
import './styles.css'

// Types for API responses
interface BalanceData {
  balance?: string;
  currency?: string;
  available?: string;
}

interface Transaction {
  id: string;
  type: string;
  amount: string | number;
  currency?: string;
  created_at?: string;
  description?: string;
}

interface TransactionsData {
  transactions?: Transaction[];
}

interface SpendingInsights {
  totalSpent: number;
  totalReceived: number;
  transactionCount: number;
  topCategories: string[];
}

function App() {
  const wsUrl = import.meta.env.VITE_WS_URL || 'ws://localhost:8080/ws'
  const liminalApiUrl = import.meta.env.VITE_API_URL || 'https://api.liminal.cash'

  // State for dashboard data
  const [balance, setBalance] = useState<BalanceData | null>(null)
  const [transactions, setTransactions] = useState<Transaction[]>([])
  const [spendingInsights, setSpendingInsights] = useState<SpendingInsights | null>(null)
  const [isLoading, setIsLoading] = useState(false)

  return (
    <>
      <main>
        <h1>Build financial autonomy for AI</h1>

        {/* Dashboard Section */}
        {true ? (
          <div className="dashboard-section">
            {/* Balance Card */}
            <div className="balance-card">
              <div className="balance-header">
                <span className="balance-label">Wallet Balance</span>
                {isLoading && <span className="loading-indicator">Updating...</span>}
              </div>
              <div className="balance-amount">
                {balance?.balance || balance?.available ? (
                  <>
                    <span className="currency">{balance.currency || 'USD'}</span>
                    <span className="amount">${balance.balance || balance.available}</span>
                  </>
                ) : (
                  <span className="amount placeholder">---.--</span>
                )}
              </div>
              <button className="refresh-btn" onClick={() => { fetchBalance(); fetchTransactions(); }}>
                ↻ Refresh
              </button>
            </div>

            {/* Spending Habits Placeholder */}
            <div className="spending-habits-section">
              <h2>Spending Habits</h2>
              <p className="section-subtitle">Based on your recent transaction activity</p>

              <div className="habits-grid">
                <div className="habit-card">
                  <div className="habit-icon">💸</div>
                  <div className="habit-label">Total Spent</div>
                  <div className="habit-value">
                    ${spendingInsights?.totalSpent.toFixed(2) || '0.00'}
                  </div>
                </div>

                <div className="habit-card">
                  <div className="habit-icon">💰</div>
                  <div className="habit-label">Total Received</div>
                  <div className="habit-value">
                    ${spendingInsights?.totalReceived.toFixed(2) || '0.00'}
                  </div>
                </div>

                <div className="habit-card">
                  <div className="habit-icon">📊</div>
                  <div className="habit-label">Transactions</div>
                  <div className="habit-value">
                    {spendingInsights?.transactionCount || 0}
                  </div>
                </div>

                <div className="habit-card coming-soon">
                  <div className="habit-icon">🎯</div>
                  <div className="habit-label">Top Categories</div>
                  <div className="habit-value small">
                    {spendingInsights?.topCategories.join(', ') || 'Coming soon'}
                  </div>
                </div>
              </div>

              <div className="habits-placeholder">
                <p>🚀 More spending insights coming soon!</p>
                <p className="placeholder-hint">Ask Nim to analyze your spending patterns for detailed insights.</p>
              </div>
            </div>
          </div>
        ) : (
          <div className="login-prompt">
            <p>👋 Log in with Nim to see your balance and spending habits</p>
            <p className="login-hint">Click the chat bubble in the bottom right to get started</p>
          </div>
        )}
      </main>

      <NimChat
        wsUrl={wsUrl}
        apiUrl={liminalApiUrl}
        title="Nim"
        position="bottom-right"
        defaultOpen={false}
      />
    </>
  )
}

ReactDOM.createRoot(document.getElementById('root')!).render(<App />)
