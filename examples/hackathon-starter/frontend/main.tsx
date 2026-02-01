import { useState, useEffect, useCallback } from 'react'
import ReactDOM from 'react-dom/client'
import { NimChat } from '@liminalcash/nim-chat'
import '@liminalcash/nim-chat/styles.css'
import './styles.css'
import {
  AlertTriangle,
  BarChart3,
  CreditCard,
  TrendingUp,
  TrendingDown,
  Target,
  Wallet,
  PiggyBank,
  Folder,
  Smartphone,
  ArrowDownLeft,
  ArrowUpRight,
  ArrowDown,
  ArrowUp,
  Banknote,
  Filter,
} from 'lucide-react'

// Types for API responses
interface SubscriptionDTO {
  id: number;
  name: string;
  amount: number;
  currency: string;
  frequency: string;
  last_payment_date: string;
  created_at: string;
}

interface TransactionDTO {
  id: string;
  user_id: string;
  amount: string;
  counterparty: string;
  created_at: string;
  currency: string;
  direction: string;
  note: string;
  status: string;
  tx_hash: string;
  type: string;
  usd_value: string;
}

interface SavingsGoalDTO {
  id: number;
  user_id: string;
  name: string;
  target_amount: number;
  current_amount: number;
  category: string;
  goal_type: string;
  deadline: string;
  created_at: string;
  is_completed: boolean;
}

interface BudgetDTO {
  id: number;
  user_id: string;
  name: string;
  limit_amount: number;
  category: string;
  start_date: string;
  end_date: string;
  is_active: boolean;
  created_at: string;
}

interface DashboardSummary {
  total_subscriptions: number;
  monthly_subscription_cost: number;
  total_transactions: number;
  total_spent: number;
  total_received: number;
  active_goals: number;
  completed_goals: number;
  active_budgets: number;
}

interface DashboardData {
  subscriptions: SubscriptionDTO[];
  transactions: TransactionDTO[];
  savings_goals: SavingsGoalDTO[];
  budgets: BudgetDTO[];
  summary: DashboardSummary;
}

function App() {
  const wsUrl = (import.meta as any).env?.VITE_WS_URL || 'ws://localhost:8080/ws'
  const liminalApiUrl = (import.meta as any).env?.VITE_API_URL || 'https://api.liminal.cash'
  const backendUrl = (import.meta as any).env?.VITE_BACKEND_URL || 'http://localhost:8080'

  // State for dashboard data
  const [dashboardData, setDashboardData] = useState<DashboardData | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [isRefreshing, setIsRefreshing] = useState(false)
  const [txFilter, setTxFilter] = useState<'all' | 'credit' | 'debit'>('all')
  const [showTxFilter, setShowTxFilter] = useState(false)

  // Fetch dashboard data
  const fetchDashboardData = useCallback(async (showRefreshIndicator = false) => {
    if (showRefreshIndicator) {
      setIsRefreshing(true)
    } else {
      setIsLoading(true)
    }
    setError(null)
    try {
      const response = await fetch(`${backendUrl}/api/dashboard`)
      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`)
      }
      const data = await response.json()
      setDashboardData(data)
    } catch (err) {
      console.error('Failed to fetch dashboard data:', err)
      setError('Failed to load dashboard data. Is the backend running?')
    } finally {
      setIsLoading(false)
      setIsRefreshing(false)
    }
  }, [backendUrl])

  // Initial fetch
  useEffect(() => {
    fetchDashboardData()
  }, [fetchDashboardData])

  // Subscribe to SSE for real-time updates
  useEffect(() => {
    const eventSource = new EventSource(`${backendUrl}/api/dashboard/events`)

    eventSource.addEventListener('connected', () => {
      console.log('[SSE] Connected to dashboard events')
    })

    eventSource.addEventListener('update', (event) => {
      const data = JSON.parse(event.data)
      console.log('[SSE] Received update:', data)
      // Refresh dashboard data when an update is received
      fetchDashboardData(true)
    })

    eventSource.onerror = (error) => {
      console.error('[SSE] Connection error:', error)
    }

    return () => {
      eventSource.close()
    }
  }, [backendUrl, fetchDashboardData])

  // Format currency
  const formatCurrency = (amount: number, currency: string = 'USD') => {
    return `${amount.toFixed(2)} ${currency}`
  }

  // Format date
  const formatDate = (dateStr: string) => {
    if (!dateStr) return 'N/A'
    const date = new Date(dateStr)
    return date.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' })
  }

  // Get frequency label
  const getFrequencyLabel = (frequency: string) => {
    switch (frequency) {
      case 'weekly': return '/week'
      case 'monthly': return '/month'
      case 'yearly': return '/year'
      default: return ''
    }
  }

  // Get frequency tag color class
  const getFrequencyTagClass = (frequency: string) => {
    switch (frequency) {
      case 'weekly': return 'freq-tag-weekly'
      case 'monthly': return 'freq-tag-monthly'
      case 'yearly':
      case 'annually': return 'freq-tag-yearly'
      default: return 'freq-tag-monthly'
    }
  }

  // Calculate due in time
  const calculateDueIn = (lastPaymentDate: string, frequency: string) => {
    if (!lastPaymentDate) return 'Due soon'
    
    const lastPayment = new Date(lastPaymentDate)
    const now = new Date()
    let nextDue: Date
    
    switch (frequency) {
      case 'weekly':
        nextDue = new Date(lastPayment)
        nextDue.setDate(nextDue.getDate() + 7)
        break
      case 'monthly':
        nextDue = new Date(lastPayment)
        nextDue.setMonth(nextDue.getMonth() + 1)
        break
      case 'yearly':
      case 'annually':
        nextDue = new Date(lastPayment)
        nextDue.setFullYear(nextDue.getFullYear() + 1)
        break
      default:
        nextDue = new Date(lastPayment)
        nextDue.setMonth(nextDue.getMonth() + 1)
    }
    
    const diffTime = nextDue.getTime() - now.getTime()
    const diffDays = Math.ceil(diffTime / (1000 * 60 * 60 * 24))
    
    if (diffDays < 0) {
      return 'Overdue'
    } else if (diffDays === 0) {
      return 'Due today'
    } else if (diffDays === 1) {
      return 'Due tomorrow'
    } else if (diffDays < 7) {
      return `Due in ${diffDays} days`
    } else if (diffDays < 30) {
      const weeks = Math.floor(diffDays / 7)
      return `Due in ${weeks} ${weeks === 1 ? 'week' : 'weeks'}`
    } else {
      const months = Math.floor(diffDays / 30)
      return `Due in ${months} ${months === 1 ? 'month' : 'months'}`
    }
  }

  // Get status color
  const getStatusColor = (status: string) => {
    switch (status) {
      case 'confirmed': return 'status-success'
      case 'pending': return 'status-pending'
      case 'failed': return 'status-failed'
      default: return ''
    }
  }

  // Get direction icon
  const getDirectionIcon = (direction: string) => {
    return direction === 'credit' ? <ArrowDown size={16} /> : <ArrowUp size={16} />
  }

  // Calculate goal progress
  const calculateProgress = (current: number, target: number) => {
    return Math.min((current / target) * 100, 100)
  }

  return (
    <>
      <main className="dashboard-main">
        <header className="dashboard-header">
          <div className="header-content">
            <h1>Financial Dashboard</h1>
            <p className="header-subtitle">All your financial data in one place</p>
          </div>
        </header>

        {error && (
          <div className="error-banner">
            <span><AlertTriangle size={16} style={{ marginRight: 8, verticalAlign: 'middle' }} />{error}</span>
            <button onClick={() => fetchDashboardData()}>Retry</button>
          </div>
        )}

        {/* Unified Dashboard Content */}
        {dashboardData && (
          <div className="unified-content">
            {/* Summary Cards Section */}
            <section className="dashboard-section summary-section" data-section="summary">
              <div className="section-title-wrapper">
                <h2 className="section-title"><BarChart3 size={20} style={{ marginRight: 8, verticalAlign: 'middle' }} />Overview</h2>
                <div className="section-divider"></div>
              </div>
              <div className="summary-grid">
                <div className="summary-card orange">
                  <div className="summary-icon"><CreditCard size={24} /></div>
                  <div className="summary-content">
                    <div className="summary-label">Subscriptions</div>
                    <div className="summary-value">{dashboardData.summary.total_subscriptions}</div>
                    <div className="summary-subtitle">
                      ${dashboardData.summary.monthly_subscription_cost.toFixed(2)}/month
                    </div>
                  </div>
                </div>

                <div className="summary-card blue">
                  <div className="summary-icon"><TrendingUp size={24} /></div>
                  <div className="summary-content">
                    <div className="summary-label">Total Received</div>
                    <div className="summary-value">${dashboardData.summary.total_received.toFixed(2)}</div>
                    <div className="summary-subtitle">
                      {dashboardData.summary.total_transactions} transactions
                    </div>
                  </div>
                </div>

                <div className="summary-card brown">
                  <div className="summary-icon"><TrendingDown size={24} /></div>
                  <div className="summary-content">
                    <div className="summary-label">Total Spent</div>
                    <div className="summary-value">${dashboardData.summary.total_spent.toFixed(2)}</div>
                    <div className="summary-subtitle">This period</div>
                  </div>
                </div>

                <div className="summary-card beige">
                  <div className="summary-icon"><Target size={24} /></div>
                  <div className="summary-content">
                    <div className="summary-label">Active Goals</div>
                    <div className="summary-value">{dashboardData.summary.active_goals}</div>
                    <div className="summary-subtitle">
                      {dashboardData.summary.completed_goals} completed
                    </div>
                  </div>
                </div>
              </div>
            </section>

            {/* Savings Goals Section */}
            <section className="dashboard-section" data-section="goals">
              <div className="section-title-wrapper">
                <h2 className="section-title"><Target size={20} style={{ marginRight: 8, verticalAlign: 'middle' }} />Savings Goals</h2>
                <div className="section-meta">
                  <span className="highlight">{dashboardData.summary.active_goals}</span> active ·
                  <span className="highlight"> {dashboardData.summary.completed_goals}</span> completed
                </div>
              </div>
              {dashboardData.savings_goals.length > 0 ? (
                <div className="goals-grid">
                  {dashboardData.savings_goals.map((goal, index) => (
                    <div key={goal.id} className={`goal-card ${goal.is_completed ? 'completed' : ''}`} style={{ animationDelay: `${index * 50}ms` }}>
                      <div className="goal-header">
                        <span className="goal-icon">{goal.goal_type === 'savings' ? <PiggyBank size={18} /> : <BarChart3 size={18} />}</span>
                        <span className={`goal-type ${goal.goal_type}`}>{goal.goal_type}</span>
                      </div>
                      <div className="goal-name">{goal.name}</div>
                      {goal.category && <div className="goal-category">{goal.category}</div>}
                      <div className="goal-progress">
                        <div className="progress-bar">
                          <div
                            className="progress-fill"
                            style={{ width: `${calculateProgress(goal.current_amount, goal.target_amount)}%` }}
                          />
                        </div>
                        <div className="progress-text">
                          ${goal.current_amount.toFixed(2)} / ${goal.target_amount.toFixed(2)}
                        </div>
                      </div>
                      <div className="goal-deadline">
                        Deadline: {formatDate(goal.deadline)}
                      </div>
                    </div>
                  ))}
                </div>
              ) : (
                <div className="empty-state">
                  <p><Target size={18} style={{ marginRight: 8, verticalAlign: 'middle' }} />No savings goals yet</p>
                  <p className="hint">Ask Nim to help you set a savings goal!</p>
                </div>
              )}
            </section>

            {/* Budgets Section */}
            <section className="dashboard-section" data-section="budgets">
              <div className="section-title-wrapper">
                <h2 className="section-title"><Wallet size={20} style={{ marginRight: 8, verticalAlign: 'middle' }} />Budgets</h2>
                <div className="section-meta">
                  <span className="highlight">{dashboardData.summary.active_budgets}</span> active
                </div>
              </div>
              {dashboardData.budgets.length > 0 ? (
                <div className="budgets-grid">
                  {dashboardData.budgets.map((budget, index) => (
                    <div key={budget.id} className={`budget-card ${budget.is_active ? 'active' : 'inactive'}`} style={{ animationDelay: `${index * 50}ms` }}>
                      <div className="budget-header">
                        <span className="budget-name">{budget.name}</span>
                      </div>
                      {/* {budget.category && <div className="budget-category">{budget.category}</div>} */}
                      <div className="budget-limit">
                        <span className="limit-label">Limit:</span>
                        <span className="limit-amount">${budget.limit_amount.toFixed(2)}</span>
                      </div>
                      <div className="budget-period">
                        {formatDate(budget.start_date)} - {formatDate(budget.end_date)}
                      </div>
                    </div>
                  ))}
                </div>
              ) : (
                <div className="empty-state">
                  <p><Wallet size={18} style={{ marginRight: 8, verticalAlign: 'middle' }} />No budgets set</p>
                  <p className="hint">Ask Nim to help you create a budget!</p>
                </div>
              )}
            </section>

            {/* Horizontal Layout: Subscriptions and Transactions */}
            <div className="horizontal-sections">
              {/* Subscriptions Section */}
              <section className="dashboard-section half-width" data-section="subscriptions">
                <div className="section-title-wrapper">
                  <h2 className="section-title"><Smartphone size={20} style={{ marginRight: 8, verticalAlign: 'middle' }} />Subscriptions</h2>
                  <div className="section-meta">
                    Total: <span className="highlight">${dashboardData.summary.monthly_subscription_cost.toFixed(2)}/month</span>
                  </div>
                </div>
                {dashboardData.subscriptions.length > 0 ? (
                  <div className="subscriptions-list">
                    {dashboardData.subscriptions.map((sub, index) => (
                      <div key={sub.id} className="subscription-card" style={{ animationDelay: `${index * 50}ms` }}>
                        <div className="sub-details">
                          <div className="sub-name-row">
                            <span className="sub-name">{sub.name.replace(' subscription', '')}</span>
                            <span className={`freq-tag ${getFrequencyTagClass(sub.frequency)}`}>{sub.frequency}</span>
                          </div>
                          <div className="sub-meta">
                            <span className="sub-due">{calculateDueIn(sub.last_payment_date, sub.frequency)}</span>
                          </div>
                        </div>
                        <div className="sub-amount">
                          <span className="amount">{formatCurrency(sub.amount, sub.currency)}</span>
                          <span className="frequency">{getFrequencyLabel(sub.frequency)}</span>
                        </div>
                      </div>
                    ))}
                  </div>
                ) : (
                  <div className="empty-state">
                    <p><Smartphone size={18} style={{ marginRight: 8, verticalAlign: 'middle' }} />No subscriptions tracked yet</p>
                    <p className="hint">Ask Nim to help you track a subscription!</p>
                  </div>
                )}
              </section>

              {/* Transactions Section */}
              <section className="dashboard-section half-width" data-section="transactions">
                <div className="section-title-wrapper">
                  <h2 className="section-title"><Banknote size={20} style={{ marginRight: 8, verticalAlign: 'middle' }} />Recent Transactions</h2>
                  <div className="section-actions">
                    <div className="filter-dropdown">
                      <button 
                        className={`filter-trigger ${txFilter !== 'all' ? 'has-filter' : ''}`}
                        onClick={() => setShowTxFilter(!showTxFilter)}
                      >
                        <Filter size={16} />
                        {txFilter !== 'all' && <span className="filter-badge">{txFilter === 'credit' ? 'In' : 'Out'}</span>}
                      </button>
                      {showTxFilter && (
                        <div className="filter-menu">
                          <button 
                            className={`filter-option ${txFilter === 'all' ? 'active' : ''}`}
                            onClick={() => { setTxFilter('all'); setShowTxFilter(false); }}
                          >
                            All transactions
                          </button>
                          <button 
                            className={`filter-option ${txFilter === 'credit' ? 'active' : ''}`}
                            onClick={() => { setTxFilter('credit'); setShowTxFilter(false); }}
                          >
                            <ArrowDown size={14} /> Received
                          </button>
                          <button 
                            className={`filter-option ${txFilter === 'debit' ? 'active' : ''}`}
                            onClick={() => { setTxFilter('debit'); setShowTxFilter(false); }}
                          >
                            <ArrowUp size={14} /> Sent
                          </button>
                        </div>
                      )}
                    </div>
                  </div>
                </div>
                {dashboardData.transactions.length > 0 ? (
                  <div className="transactions-list">
                    {dashboardData.transactions
                      .filter(tx => txFilter === 'all' || tx.direction === txFilter)
                      .slice(0, 10)
                      .map((tx, index) => (
                      <div key={tx.id} className="transaction-item" style={{ animationDelay: `${index * 30}ms` }}>
                        <div className={`tx-direction ${tx.direction}`}>
                          {getDirectionIcon(tx.direction)}
                        </div>
                        <div className="tx-details">
                          <div className="tx-note">{tx.note || `${tx.type} transaction`}</div>
                          <div className="tx-meta">
                            <span className="tx-date">{formatDate(tx.created_at)}</span>
                          </div>
                        </div>
                        <div className={`tx-amount ${tx.direction}`}>
                          {tx.direction === 'credit' ? '+' : ''}{tx.amount} {tx.currency}
                        </div>
                        <div className={`tx-status ${getStatusColor(tx.status)}`}>
                          {tx.status}
                        </div>
                      </div>
                    ))}
                  </div>
                ) : (
                  <div className="empty-state">
                    <p><Banknote size={18} style={{ marginRight: 8, verticalAlign: 'middle' }} />No transactions yet</p>
                    <p className="hint">Your transaction history will appear here</p>
                  </div>
                )}
              </section>
            </div>
          </div>
        )}

        {/* Loading State */}
        {isLoading && !dashboardData && (
          <div className="loading-state">
            <div className="spinner"></div>
            <p>Loading dashboard...</p>
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
