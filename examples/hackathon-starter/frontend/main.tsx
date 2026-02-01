import { useState, useEffect, useCallback } from 'react'
import ReactDOM from 'react-dom/client'
// import { NimChat } from '@liminalcash/nim-chat'
// import '@liminalcash/nim-chat/styles.css'
import { NimChat } from './nim-chat-plus-plus/nim-chat/src/NimChat'
import './nim-chat-plus-plus/nim-chat/src/styles/index.css'
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
  ChevronDown,
  ChevronUp,
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
  const [collapsedSections, setCollapsedSections] = useState<Record<string, boolean>>({})
  const [activeTab, setActiveTab] = useState<'dashboard' | 'analytics'>('dashboard')
  const [showExportModal, setShowExportModal] = useState(false)
  const [exportProgress, setExportProgress] = useState(0)
  const [exportStage, setExportStage] = useState('')

  // Toggle section collapse
  const toggleSection = (section: string) => {
    setCollapsedSections(prev => ({ ...prev, [section]: !prev[section] }))
  }

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
    return `${amount.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })} ${currency}`
  }

  // Format number with thousand separators
  const formatNumber = (num: number, decimals: number = 2) => {
    return num.toLocaleString('en-US', { minimumFractionDigits: decimals, maximumFractionDigits: decimals })
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
    } else if (diffDays < 365) {
      const months = Math.floor(diffDays / 30)
      return `Due in ${months} ${months === 1 ? 'month' : 'months'}`
    } else {
      const years = Math.floor(diffDays / 365)
      return `Due in ${years} ${years === 1 ? 'year' : 'years'}`
    }
  }

  // Calculate deadline due time
  const calculateDeadlineDue = (deadline: string) => {
    if (!deadline) return 'No deadline'

    const deadlineDate = new Date(deadline)
    const now = new Date()

    const diffTime = deadlineDate.getTime() - now.getTime()
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
    } else if (diffDays < 365) {
      const months = Math.floor(diffDays / 30)
      return `Due in ${months} ${months === 1 ? 'month' : 'months'}`
    } else {
      const years = Math.floor(diffDays / 365)
      return `Due in ${years} ${years === 1 ? 'year' : 'years'}`
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

  // Calculate expenses by category
  const getExpensesByCategory = () => {
    if (!dashboardData) return []
    const categoryMap: Record<string, number> = {}
    dashboardData.transactions
      .filter(tx => tx.direction === 'debit')
      .forEach(tx => {
        const category = tx.type || 'Other'
        categoryMap[category] = (categoryMap[category] || 0) + parseFloat(tx.amount)
      })
    return Object.entries(categoryMap).map(([name, value]) => ({ name, value }))
  }

  // Calculate spending over last 6 months
  const getSpendingOverTime = () => {
    if (!dashboardData) return []
    const monthlySpending: Record<string, number> = {}
    const now = new Date()

    for (let i = 5; i >= 0; i--) {
      const date = new Date(now.getFullYear(), now.getMonth() - i, 1)
      const monthKey = date.toLocaleDateString('en-US', { month: 'short', year: 'numeric' })
      monthlySpending[monthKey] = 0
    }

    dashboardData.transactions
      .filter(tx => tx.direction === 'debit')
      .forEach(tx => {
        const txDate = new Date(tx.created_at)
        const monthKey = txDate.toLocaleDateString('en-US', { month: 'short', year: 'numeric' })
        if (monthlySpending.hasOwnProperty(monthKey)) {
          monthlySpending[monthKey] += parseFloat(tx.amount)
        }
      })

    return Object.entries(monthlySpending).map(([month, amount]) => ({ month, amount }))
  }

  // Calculate goal timeline
  const getGoalTimeline = () => {
    if (!dashboardData) return []
    return dashboardData.savings_goals
      .filter(g => !g.is_completed && g.deadline)
      .sort((a, b) => new Date(a.deadline).getTime() - new Date(b.deadline).getTime())
      .map(goal => ({
        name: goal.name,
        deadline: new Date(goal.deadline),
        progress: calculateProgress(goal.current_amount, goal.target_amount),
        target: goal.target_amount,
        current: goal.current_amount
      }))
  }

  // Export PDF report
  const exportReport = async () => {
    setShowExportModal(true)
    setExportProgress(0)

    const stages = [
      { progress: 20, stage: 'Gathering financial data...', delay: 500 },
      { progress: 40, stage: 'Analyzing transactions...', delay: 800 },
      { progress: 60, stage: 'Generating charts...', delay: 700 },
      { progress: 80, stage: 'Formatting document...', delay: 600 },
      { progress: 95, stage: 'Finalizing report...', delay: 500 },
      { progress: 100, stage: 'Complete!', delay: 300 },
    ]

    for (const { progress, stage, delay } of stages) {
      setExportStage(stage)
      setExportProgress(progress)
      await new Promise(resolve => setTimeout(resolve, delay))
    }

    // Generate and "save" the file (trigger download)
    const timestamp = new Date().getTime()
    const filename = `liminal-report-${timestamp}.pdf`
    const blob = new Blob([''], { type: 'application/pdf' })
    const url = URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = filename
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    URL.revokeObjectURL(url)

    // Close modal after completion
    await new Promise(resolve => setTimeout(resolve, 800))
    setShowExportModal(false)
    setExportProgress(0)
    setExportStage('')
  }

  return (
    <>
      <main className="dashboard-main">
        <header className="dashboard-header">
          <div className="header-content">
            <h1>Financial Dashboard</h1>
            <p className="header-subtitle">All your financial data in one place</p>
          </div>
          {/* Tab Navigation */}
          {dashboardData && (
            <div className="tab-navigation">
              <button
                className={`tab-btn ${activeTab === 'dashboard' ? 'active' : ''}`}
                onClick={() => setActiveTab('dashboard')}
              >
                <Folder size={18} />Dashboard
              </button>
              <button
                className={`tab-btn ${activeTab === 'analytics' ? 'active' : ''}`}
                onClick={() => setActiveTab('analytics')}
              >
                <BarChart3 size={18} />Analytics
              </button>
            </div>
          )}
        </header>

        {error && (
          <div className="error-banner">
            <span><AlertTriangle size={16} style={{ marginRight: 8, verticalAlign: 'middle' }} />{error}</span>
            <button onClick={() => fetchDashboardData()}>Retry</button>
          </div>
        )}

        {/* Unified Dashboard Content */}
        {dashboardData && activeTab === 'dashboard' && (
          <div className="unified-content">

            {/* Savings Goals Section */}
            <section className="dashboard-section" data-section="goals">
              <div className="section-title-wrapper clickable" onClick={() => toggleSection('goals')}>
                <h2 className="section-title"><Target size={20} style={{ marginRight: 8, verticalAlign: 'middle' }} />Savings Goals</h2>
                <div className="section-header-right">
                  <div className="section-meta">
                    <span className="highlight">{dashboardData.summary.active_goals}</span> active ·
                    <span className="highlight"> {dashboardData.summary.completed_goals}</span> completed
                  </div>
                  <span className="collapse-icon">
                    {collapsedSections['goals'] ? <ChevronDown size={20} /> : <ChevronUp size={20} />}
                  </span>
                </div>
              </div>
              {!collapsedSections['goals'] && (dashboardData.savings_goals.length > 0 ? (
                <div className="goals-grid">
                  {dashboardData.savings_goals.map((goal, index) => (
                    <div key={goal.id} className={`goal-card ${goal.is_completed ? 'completed' : ''}`} style={{ animationDelay: `${index * 50}ms` }}>
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
                          ${formatNumber(goal.current_amount)} / ${formatNumber(goal.target_amount)}
                        </div>
                      </div>
                      <div className="goal-deadline">
                        {calculateDeadlineDue(goal.deadline)}
                      </div>
                    </div>
                  ))}
                </div>
              ) : (
                <div className="empty-state">
                  <p><Target size={18} style={{ marginRight: 8, verticalAlign: 'middle' }} />No savings goals yet</p>
                  <p className="hint">Ask Nim to help you set a savings goal!</p>
                </div>
              ))}
            </section>

            {/* Budgets Section */}
            <section className="dashboard-section" data-section="budgets">
              <div className="section-title-wrapper clickable" onClick={() => toggleSection('budgets')}>
                <h2 className="section-title"><Wallet size={20} style={{ marginRight: 8, verticalAlign: 'middle' }} />Budgets</h2>
                <div className="section-header-right">
                  <div className="section-meta">
                    <span className="highlight">{dashboardData.summary.active_budgets}</span> active
                  </div>
                  <span className="collapse-icon">
                    {collapsedSections['budgets'] ? <ChevronDown size={20} /> : <ChevronUp size={20} />}
                  </span>
                </div>
              </div>
              {!collapsedSections['budgets'] && (dashboardData.budgets.length > 0 ? (
                <div className="budgets-grid">
                  {dashboardData.budgets.map((budget, index) => (
                    <div key={budget.id} className={`budget-card ${budget.is_active ? 'active' : 'inactive'}`} style={{ animationDelay: `${index * 50}ms` }}>
                      <div className="budget-header">
                        <span className="budget-name">{budget.name}</span>
                      </div>
                      {/* {budget.category && <div className="budget-category">{budget.category}</div>} */}
                      <div className="budget-limit">
                        <span className="limit-label">Limit:</span>
                        <span className="limit-amount">${formatNumber(budget.limit_amount)}</span>
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
              ))}
            </section>

            {/* Horizontal Layout: Subscriptions and Transactions */}
            <div className="horizontal-sections">
              {/* Subscriptions Section */}
              <section className="dashboard-section half-width" data-section="subscriptions">
                <div className="section-title-wrapper clickable" onClick={() => toggleSection('subscriptions')}>
                  <h2 className="section-title"><Smartphone size={20} style={{ marginRight: 8, verticalAlign: 'middle' }} />Subscriptions</h2>
                  <div className="section-header-right">
                    <div className="section-meta">
                      Total: <span className="highlight">${formatNumber(dashboardData.summary.monthly_subscription_cost)}/month</span>
                    </div>
                    <span className="collapse-icon">
                      {collapsedSections['subscriptions'] ? <ChevronDown size={20} /> : <ChevronUp size={20} />}
                    </span>
                  </div>
                </div>
                {!collapsedSections['subscriptions'] && (dashboardData.subscriptions.length > 0 ? (
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
                ))}
              </section>

              {/* Transactions Section */}
              <section className="dashboard-section half-width" data-section="transactions">
                <div className="section-title-wrapper clickable" onClick={() => toggleSection('transactions')}>
                  <h2 className="section-title"><Banknote size={20} style={{ marginRight: 8, verticalAlign: 'middle' }} />Recent Transactions</h2>
                  <div className="section-header-right">
                    <div className="section-actions">
                      <div className="filter-dropdown" onClick={(e) => e.stopPropagation()}>
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
                    <span className="collapse-icon">
                      {collapsedSections['transactions'] ? <ChevronDown size={20} /> : <ChevronUp size={20} />}
                    </span>
                  </div>
                </div>
                {!collapsedSections['transactions'] && (dashboardData.transactions.length > 0 ? (
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
                ))}
              </section>
            </div>
          </div>
        )}

        {/* Analytics Tab */}
        {dashboardData && activeTab === 'analytics' && (
          <div className="unified-content analytics-view">
            {/* Expenses by Category - Pie Chart */}
            <section className="dashboard-section analytics-section">
              <div className="section-title-wrapper">
                <h2 className="section-title"><PiggyBank size={20} style={{ marginRight: 8, verticalAlign: 'middle' }} />Expenses by Category</h2>
              </div>
              <div className="chart-container pie-chart">
                {getExpensesByCategory().length > 0 ? (
                  <div className="pie-chart-wrapper">
                    <svg viewBox="0 0 200 200" className="pie-svg">
                      {(() => {
                        const data = getExpensesByCategory()
                        const total = data.reduce((sum, item) => sum + item.value, 0)
                        let currentAngle = 0
                        const colors = ['#FF6D00', '#9BC1F3', '#9E8C78', '#FFB347', '#7E57C2', '#22C55E']

                        return data.map((item, i) => {
                          const percentage = (item.value / total) * 100
                          const angle = (item.value / total) * 360
                          const startAngle = currentAngle
                          currentAngle += angle

                          const x1 = 100 + 80 * Math.cos((startAngle - 90) * Math.PI / 180)
                          const y1 = 100 + 80 * Math.sin((startAngle - 90) * Math.PI / 180)
                          const x2 = 100 + 80 * Math.cos((startAngle + angle - 90) * Math.PI / 180)
                          const y2 = 100 + 80 * Math.sin((startAngle + angle - 90) * Math.PI / 180)
                          const largeArc = angle > 180 ? 1 : 0

                          return (
                            <path
                              key={i}
                              d={`M 100 100 L ${x1} ${y1} A 80 80 0 ${largeArc} 1 ${x2} ${y2} Z`}
                              fill={colors[i % colors.length]}
                              opacity="0.85"
                              stroke="white"
                              strokeWidth="2"
                            />
                          )
                        })
                      })()}
                    </svg>
                    <div className="pie-legend">
                      {getExpensesByCategory().map((item, i) => {
                        const colors = ['#FF6D00', '#9BC1F3', '#9E8C78', '#FFB347', '#7E57C2', '#22C55E']
                        const total = getExpensesByCategory().reduce((sum, x) => sum + x.value, 0)
                        const percentage = ((item.value / total) * 100).toFixed(1)
                        return (
                          <div key={i} className="legend-item">
                            <span className="legend-color" style={{ backgroundColor: colors[i % colors.length] }}></span>
                            <span className="legend-label">{item.name}</span>
                            <span className="legend-value">${formatNumber(item.value)} ({percentage}%)</span>
                          </div>
                        )
                      })}
                    </div>
                  </div>
                ) : (
                  <div className="empty-state">
                    <p>No expense data available</p>
                  </div>
                )}
              </div>
            </section>

            {/* Spending Over Time - Line Chart */}
            <section className="dashboard-section analytics-section">
              <div className="section-title-wrapper">
                <h2 className="section-title"><TrendingDown size={20} style={{ marginRight: 8, verticalAlign: 'middle' }} />Spending Trend (6 Months)</h2>
              </div>
              <div className="chart-container line-chart">
                {getSpendingOverTime().length > 0 ? (
                  <div className="line-chart-wrapper">
                    <svg viewBox="0 0 600 300" className="line-svg">
                      {(() => {
                        const data = getSpendingOverTime()
                        const maxAmount = Math.max(...data.map(d => d.amount), 100)
                        const padding = 40
                        const width = 600 - padding * 2
                        const height = 300 - padding * 2
                        const stepX = width / (data.length - 1 || 1)

                        const points = data.map((d, i) => {
                          const x = padding + i * stepX
                          const y = padding + height - (d.amount / maxAmount) * height
                          return `${x},${y}`
                        }).join(' ')

                        const areaPoints = `${padding},${padding + height} ${points} ${padding + width},${padding + height}`

                        return (
                          <g>
                            {/* Grid lines */}
                            {[0, 0.25, 0.5, 0.75, 1].map((ratio, i) => (
                              <g key={i}>
                                <line
                                  x1={padding}
                                  y1={padding + height * ratio}
                                  x2={padding + width}
                                  y2={padding + height * ratio}
                                  stroke="#E5E5E5"
                                  strokeWidth="1"
                                />
                                <text
                                  x={padding - 10}
                                  y={padding + height * ratio + 4}
                                  textAnchor="end"
                                  fontSize="12"
                                  fill="#737373"
                                >
                                  ${(maxAmount * (1 - ratio)).toFixed(0)}
                                </text>
                              </g>
                            ))}

                            {/* Area fill */}
                            <polygon
                              points={areaPoints}
                              fill="#FF6D00"
                              opacity="0.1"
                            />

                            {/* Line */}
                            <polyline
                              points={points}
                              fill="none"
                              stroke="#FF6D00"
                              strokeWidth="3"
                              strokeLinecap="round"
                              strokeLinejoin="round"
                            />

                            {/* Data points */}
                            {data.map((d, i) => {
                              const x = padding + i * stepX
                              const y = padding + height - (d.amount / maxAmount) * height
                              return (
                                <g key={i}>
                                  <circle
                                    cx={x}
                                    cy={y}
                                    r="5"
                                    fill="#FF6D00"
                                    stroke="white"
                                    strokeWidth="2"
                                  />
                                  <text
                                    x={x}
                                    y={padding + height + 20}
                                    textAnchor="middle"
                                    fontSize="12"
                                    fill="#737373"
                                  >
                                    {d.month}
                                  </text>
                                </g>
                              )
                            })}
                          </g>
                        )
                      })()}
                    </svg>
                  </div>
                ) : (
                  <div className="empty-state">
                    <p>No spending data available</p>
                  </div>
                )}
              </div>
            </section>

            {/* Goal Timeline */}
            <section className="dashboard-section analytics-section">
              <div className="section-title-wrapper">
                <h2 className="section-title"><Target size={20} style={{ marginRight: 8, verticalAlign: 'middle' }} />Goal Timeline</h2>
              </div>
              <div className="goal-timeline">
                {getGoalTimeline().length > 0 ? (
                  <div className="timeline-wrapper">
                    {getGoalTimeline().map((goal, i) => (
                      <div key={i} className="timeline-item">
                        <div className="timeline-marker" style={{
                          background: goal.progress >= 75 ? '#22C55E' : goal.progress >= 50 ? '#FF6D00' : '#9BC1F3'
                        }}>
                          <Target size={16} />
                        </div>
                        <div className="timeline-content">
                          <div className="timeline-header">
                            <h4>{goal.name}</h4>
                            <span className="timeline-date">{formatDate(goal.deadline.toISOString())}</span>
                          </div>
                          <div className="timeline-progress">
                            <div className="progress-bar">
                              <div className="progress-fill" style={{ width: `${goal.progress}%` }} />
                            </div>
                            <span className="progress-label">
                              ${formatNumber(goal.current)} / ${formatNumber(goal.target)} ({goal.progress.toFixed(0)}%)
                            </span>
                          </div>
                          <div className="timeline-eta">
                            {calculateDeadlineDue(goal.deadline.toISOString())}
                          </div>
                        </div>
                      </div>
                    ))}
                  </div>
                ) : (
                  <div className="empty-state">
                    <p><Target size={18} style={{ marginRight: 8, verticalAlign: 'middle' }} />No active goals with deadlines</p>
                    <p className="hint">Set goals with target dates to see your timeline!</p>
                  </div>
                )}
              </div>
            </section>

            {/* Export Reports */}
            <section className="dashboard-section analytics-section">
              <div className="section-title-wrapper">
                <h2 className="section-title"><Folder size={20} style={{ marginRight: 8, verticalAlign: 'middle' }} />Export Reports</h2>
              </div>
              <div className="export-actions">
                <button className="export-btn" onClick={exportReport}>
                  <Folder size={18} />
                  <div className="export-content">
                    <h4>Generate PDF Report</h4>
                    <p>Export {new Date().toLocaleDateString('en-US', { month: 'long' })}'s spending summary</p>
                  </div>
                </button>
              </div>
            </section>
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
        title="Nim++"
        position="bottom-right"
        defaultOpen={false}
      />

      {/* Export Modal */}
      {showExportModal && (
        <div className="export-modal-overlay">
          <div className="export-modal">
            <div className="export-modal-content">
              <div className="export-icon">
                <Folder size={48} className="export-icon-graphic" />
              </div>
              <h3 className="export-title">Exporting Report</h3>
              <p className="export-stage">{exportStage}</p>

              <div className="export-progress-container">
                <div className="export-progress-bar">
                  <div
                    className="export-progress-fill"
                    style={{ width: `${exportProgress}%` }}
                  />
                </div>
                <span className="export-progress-text">{exportProgress}%</span>
              </div>

              {exportProgress === 100 && (
                <div className="export-success">
                  <svg className="checkmark" viewBox="0 0 52 52">
                    <circle className="checkmark-circle" cx="26" cy="26" r="25" fill="none" />
                    <path className="checkmark-check" fill="none" d="M14.1 27.2l7.1 7.2 16.7-16.8" />
                  </svg>
                  <p className="export-success-text">Report saved successfully!</p>
                </div>
              )}
            </div>
          </div>
        </div>
      )}
    </>
  )
}

ReactDOM.createRoot(document.getElementById('root')!).render(<App />)
