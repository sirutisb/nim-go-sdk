import { useState, useEffect, useCallback } from 'react'
import ReactDOM from 'react-dom/client'
import { NimChat } from '@liminalcash/nim-chat'
import '@liminalcash/nim-chat/styles.css'
import './styles.css'
import { motion, AnimatePresence } from 'framer-motion'
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
  const [selectedCurrency, setSelectedCurrency] = useState<'USD' | 'LIL'>('USD')
  const [hoveredSlice, setHoveredSlice] = useState<{name: string, value: number, percentage: string} | null>(null)

  // Mock balance data
  const balances = {
    USD: 2.00,
    LIL: 11.49
  }

  // Mock analytics data0
  const mockExpensesByCategory = [
    { name: 'Food & Dining', value: 450 },
    { name: 'Transportation', value: 220 },
    { name: 'Entertainment', value: 180 },
    { name: 'Shopping', value: 320 },
    { name: 'Bills & Utilities', value: 280 },
  ]

  const mockSpendingOverTime = [
    { month: 'Aug 2025', amount: 1200 },
    { month: 'Sep 2025', amount: 1450 },
    { month: 'Oct 2025', amount: 980 },
    { month: 'Nov 2025', amount: 1650 },
    { month: 'Dec 2025', amount: 1320 },
    { month: 'Jan 2026', amount: 1450 },
  ]

  // Mock data: Income vs Expenses by month
  const mockIncomeVsExpenses = [
    { month: 'Sep', income: 4200, expenses: 3100 },
    { month: 'Oct', income: 4500, expenses: 3400 },
    { month: 'Nov', income: 4100, expenses: 2900 },
    { month: 'Dec', income: 5200, expenses: 4100 },
    { month: 'Jan', income: 4800, expenses: 3200 },
    { month: 'Feb', income: 4600, expenses: 2800 },
  ]

  // Mock data: Weekly activity (transactions per day)
  const mockWeeklyActivity = [
    { day: 'Mon', transactions: 5, amount: 120 },
    { day: 'Tue', transactions: 3, amount: 85 },
    { day: 'Wed', transactions: 8, amount: 240 },
    { day: 'Thu', transactions: 4, amount: 95 },
    { day: 'Fri', transactions: 12, amount: 380 },
    { day: 'Sat', transactions: 15, amount: 520 },
    { day: 'Sun', transactions: 6, amount: 150 },
  ]

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
    const now = new Date()
    const monthYear = now.toLocaleDateString('en-US', { month: 'long', year: 'numeric' })
    alert(`Generating PDF report for ${monthYear}...\n\nThis would export a comprehensive spending report with:\n- Total spending breakdown\n- Category analysis\n- Transaction history\n- Budget comparison`)
  }

  // Animation variants
  const fadeInUp = {
    initial: { opacity: 0, y: 20 },
    animate: { opacity: 1, y: 0 },
    exit: { opacity: 0, y: -20 }
  }

  const staggerContainer = {
    animate: {
      transition: {
        staggerChildren: 0.1
      }
    }
  }

  const scaleIn = {
    initial: { opacity: 0, scale: 0.9 },
    animate: { opacity: 1, scale: 1 },
    exit: { opacity: 0, scale: 0.9 }
  }

  const slideIn = {
    initial: { opacity: 0, x: -20 },
    animate: { opacity: 1, x: 0 },
    exit: { opacity: 0, x: 20 }
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
              <motion.button 
                className={`tab-btn ${activeTab === 'dashboard' ? 'active' : ''}`}
                onClick={() => setActiveTab('dashboard')}
                whileHover={{ y: -2 }}
                whileTap={{ scale: 0.95 }}
                transition={{ type: "spring", stiffness: 400, damping: 17 }}
              >
                <Folder size={18} />Dashboard
              </motion.button>
              <motion.button 
                className={`tab-btn ${activeTab === 'analytics' ? 'active' : ''}`}
                onClick={() => setActiveTab('analytics')}
                whileHover={{ y: -2 }}
                whileTap={{ scale: 0.95 }}
                transition={{ type: "spring", stiffness: 400, damping: 17 }}
              >
                <BarChart3 size={18} />Analytics
              </motion.button>
            </div>
          )}
        </header>

        <AnimatePresence mode="wait">
          {error && (
            <motion.div 
              className="error-banner"
              initial={{ opacity: 0, height: 0 }}
              animate={{ opacity: 1, height: 'auto' }}
              exit={{ opacity: 0, height: 0 }}
              transition={{ duration: 0.3 }}
            >
              <span><AlertTriangle size={16} style={{ marginRight: 8, verticalAlign: 'middle' }} />{error}</span>
              <button onClick={() => fetchDashboardData()}>Retry</button>
            </motion.div>
          )}
        </AnimatePresence>

        {/* Unified Dashboard Content */}
        <AnimatePresence mode="wait">
          {dashboardData && activeTab === 'dashboard' && (
            <motion.div 
              className="unified-content"
              key="dashboard"
              initial="initial"
              animate="animate"
              exit="exit"
              variants={fadeInUp}
              transition={{ duration: 0.4 }}
            >
              {/* Balance Display */}
              <motion.div 
                className="balance-section"
                variants={scaleIn}
                initial="initial"
                animate="animate"
                transition={{ duration: 0.5, delay: 0.1 }}
              >
                <motion.p 
                  className="page-title"
                  initial={{ opacity: 0 }}
                  animate={{ opacity: 1 }}
                  transition={{ delay: 0.2 }}
                >
                  Your Balance
                </motion.p>
                <motion.h1 
                  className="balance-amount"
                  key={selectedCurrency}
                  initial={{ opacity: 0, scale: 0.8 }}
                  animate={{ opacity: 1, scale: 1 }}
                  transition={{ type: "spring", stiffness: 200, damping: 15 }}
                >
                  {selectedCurrency === 'USD' ? '$' : ''}{balances[selectedCurrency].toFixed(2)} {selectedCurrency === 'LIL' ? 'LIL' : ''}
                </motion.h1>
                <div className="currency-pills">
                  <motion.button 
                    className={`currency-pill ${selectedCurrency === 'USD' ? 'active' : ''}`}
                    onClick={() => setSelectedCurrency('USD')}
                    whileHover={{ scale: 1.05 }}
                    whileTap={{ scale: 0.95 }}
                    transition={{ type: "spring", stiffness: 400, damping: 17 }}
                  >
                    USD
                  </motion.button>
                  <motion.button 
                    className={`currency-pill ${selectedCurrency === 'LIL' ? 'active' : ''}`}
                    onClick={() => setSelectedCurrency('LIL')}
                    whileHover={{ scale: 1.05 }}
                    whileTap={{ scale: 0.95 }}
                    transition={{ type: "spring", stiffness: 400, damping: 17 }}
                  >
                    LIL
                  </motion.button>
                </div>
              </motion.div>

            {/* Savings Goals Section */}
            <motion.section 
              className="dashboard-section" 
              data-section="goals"
              variants={fadeInUp}
              initial="initial"
              animate="animate"
              transition={{ delay: 0.2 }}
            >
              <motion.div 
                className="section-title-wrapper clickable" 
                onClick={() => toggleSection('goals')}
                whileHover={{ x: 5 }}
                transition={{ type: "spring", stiffness: 300 }}
              >
                <h2 className="section-title"><Target size={20} style={{ marginRight: 8, verticalAlign: 'middle' }} />Savings Goals</h2>
                <div className="section-header-right">
                  <div className="section-meta">
                    <span className="highlight">{dashboardData.summary.active_goals}</span> active ·
                    <span className="highlight"> {dashboardData.summary.completed_goals}</span> completed
                  </div>
                  <motion.span 
                    className="collapse-icon"
                    animate={{ rotate: collapsedSections['goals'] ? 0 : 180 }}
                    transition={{ duration: 0.3 }}
                  >
                    <ChevronDown size={20} />
                  </motion.span>
                </div>
              </motion.div>
              <AnimatePresence>
                {!collapsedSections['goals'] && (dashboardData.savings_goals.length > 0 ? (
                  <motion.div 
                    className="goals-grid"
                    initial={{ opacity: 0, height: 0 }}
                    animate={{ opacity: 1, height: 'auto' }}
                    exit={{ opacity: 0, height: 0 }}
                    transition={{ duration: 0.3 }}
                  >
                    {dashboardData.savings_goals.map((goal, index) => (
                      <motion.div 
                        key={goal.id} 
                        className={`goal-card ${goal.is_completed ? 'completed' : ''}`}
                        initial={{ opacity: 0, y: 20 }}
                        animate={{ opacity: 1, y: 0 }}
                        transition={{ delay: index * 0.1 }}
                        whileHover={{ scale: 1.02, y: -5 }}
                        whileTap={{ scale: 0.98 }}
                      >
                      <div className="goal-name">{goal.name}</div>
                      {goal.category && <div className="goal-category">{goal.category}</div>}
                      <div className="goal-progress">
                        <div className="progress-bar">
                          <motion.div
                            className="progress-fill"
                            initial={{ width: 0 }}
                            animate={{ width: `${calculateProgress(goal.current_amount, goal.target_amount)}%` }}
                            transition={{ duration: 1, delay: index * 0.1 + 0.3, ease: "easeOut" }}
                          />
                        </div>
                        <div className="progress-text">
                          ${formatNumber(goal.current_amount)} / ${formatNumber(goal.target_amount)}
                        </div>
                      </div>
                      <div className="goal-deadline">
                        {calculateDeadlineDue(goal.deadline)}
                      </div>
                    </motion.div>
                  ))}
                  </motion.div>
                ) : (
                  <motion.div 
                    className="empty-state"
                    initial={{ opacity: 0 }}
                    animate={{ opacity: 1 }}
                    exit={{ opacity: 0 }}
                  >
                    <p><Target size={18} style={{ marginRight: 8, verticalAlign: 'middle' }} />No savings goals yet</p>
                    <p className="hint">Ask Nim to help you set a savings goal!</p>
                  </motion.div>
                ))}
              </AnimatePresence>
            </motion.section>

            {/* Budgets Section */}
            <motion.section 
              className="dashboard-section" 
              data-section="budgets"
              variants={fadeInUp}
              initial="initial"
              animate="animate"
              transition={{ delay: 0.3 }}
            >
              <motion.div 
                className="section-title-wrapper clickable" 
                onClick={() => toggleSection('budgets')}
                whileHover={{ x: 5 }}
                transition={{ type: "spring", stiffness: 300 }}
              >
                <h2 className="section-title"><Wallet size={20} style={{ marginRight: 8, verticalAlign: 'middle' }} />Budgets</h2>
                <div className="section-header-right">
                  <div className="section-meta">
                    <span className="highlight">{dashboardData.summary.active_budgets}</span> active
                  </div>
                  <motion.span 
                    className="collapse-icon"
                    animate={{ rotate: collapsedSections['budgets'] ? 0 : 180 }}
                    transition={{ duration: 0.3 }}
                  >
                    <ChevronDown size={20} />
                  </motion.span>
                </div>
              </motion.div>
              <AnimatePresence>
                {!collapsedSections['budgets'] && (dashboardData.budgets.length > 0 ? (
                  <motion.div 
                    className="budgets-grid"
                    initial={{ opacity: 0, height: 0 }}
                    animate={{ opacity: 1, height: 'auto' }}
                    exit={{ opacity: 0, height: 0 }}
                    transition={{ duration: 0.3 }}
                  >
                    {dashboardData.budgets.map((budget, index) => (
                      <motion.div 
                        key={budget.id} 
                        className={`budget-card ${budget.is_active ? 'active' : 'inactive'}`}
                        initial={{ opacity: 0, y: 20 }}
                        animate={{ opacity: 1, y: 0 }}
                        transition={{ delay: index * 0.1 }}
                        whileHover={{ scale: 1.02, y: -5 }}
                        whileTap={{ scale: 0.98 }}
                      >
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
                    </motion.div>
                  ))}
                  </motion.div>
                ) : (
                  <motion.div 
                    className="empty-state"
                    initial={{ opacity: 0 }}
                    animate={{ opacity: 1 }}
                    exit={{ opacity: 0 }}
                  >
                    <p><Wallet size={18} style={{ marginRight: 8, verticalAlign: 'middle' }} />No budgets set</p>
                    <p className="hint">Ask Nim to help you create a budget!</p>
                  </motion.div>
                ))}
              </AnimatePresence>
            </motion.section>

            {/* Horizontal Layout: Subscriptions and Transactions */}
            <motion.div 
              className="horizontal-sections"
              variants={staggerContainer}
              initial="initial"
              animate="animate"
            >
              {/* Subscriptions Section */}
              <motion.section 
                className="dashboard-section half-width" 
                data-section="subscriptions"
                variants={slideIn}
                whileHover={{ y: -5 }}
                transition={{ type: "spring", stiffness: 300 }}
              >
                <motion.div 
                  className="section-title-wrapper clickable" 
                  onClick={() => toggleSection('subscriptions')}
                  whileHover={{ x: 5 }}
                >
                  <h2 className="section-title"><Smartphone size={20} style={{ marginRight: 8, verticalAlign: 'middle' }} />Subscriptions</h2>
                  <div className="section-header-right">
                    <div className="section-meta">
                      Total: <span className="highlight">${formatNumber(dashboardData.summary.monthly_subscription_cost)}/month</span>
                    </div>
                    <motion.span 
                      className="collapse-icon"
                      animate={{ rotate: collapsedSections['subscriptions'] ? 0 : 180 }}
                      transition={{ duration: 0.3 }}
                    >
                      <ChevronDown size={20} />
                    </motion.span>
                  </div>
                </motion.div>
                <AnimatePresence>
                  {!collapsedSections['subscriptions'] && (dashboardData.subscriptions.length > 0 ? (
                    <motion.div 
                      className="subscriptions-list"
                      initial={{ opacity: 0, height: 0 }}
                      animate={{ opacity: 1, height: 'auto' }}
                      exit={{ opacity: 0, height: 0 }}
                      transition={{ duration: 0.3 }}
                    >
                      {dashboardData.subscriptions.map((sub, index) => (
                        <motion.div 
                          key={sub.id} 
                          className="subscription-card"
                          initial={{ opacity: 0, x: -20 }}
                          animate={{ opacity: 1, x: 0 }}
                          transition={{ delay: index * 0.05 }}
                          whileHover={{ scale: 1.02, x: 5 }}
                        >
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
                      </motion.div>
                    ))}
                    </motion.div>
                  ) : (
                    <motion.div 
                      className="empty-state"
                      initial={{ opacity: 0 }}
                      animate={{ opacity: 1 }}
                      exit={{ opacity: 0 }}
                    >
                      <p><Smartphone size={18} style={{ marginRight: 8, verticalAlign: 'middle' }} />No subscriptions tracked yet</p>
                      <p className="hint">Ask Nim to help you track a subscription!</p>
                    </motion.div>
                  ))}
                </AnimatePresence>
              </motion.section>

              {/* Transactions Section */}
              <motion.section 
                className="dashboard-section half-width" 
                data-section="transactions"
                variants={slideIn}
                whileHover={{ y: -5 }}
                transition={{ type: "spring", stiffness: 300 }}
              >
                <motion.div 
                  className="section-title-wrapper clickable" 
                  onClick={() => toggleSection('transactions')}
                  whileHover={{ x: 5 }}
                >
                  <h2 className="section-title"><Banknote size={20} style={{ marginRight: 8, verticalAlign: 'middle' }} />Recent Transactions</h2>
                  <div className="section-header-right">
                    <div className="section-actions">
                      <div className="filter-dropdown" onClick={(e) => e.stopPropagation()}>
                        <motion.button 
                          className={`filter-trigger ${txFilter !== 'all' ? 'has-filter' : ''}`}
                          onClick={() => setShowTxFilter(!showTxFilter)}
                          whileHover={{ scale: 1.1 }}
                          whileTap={{ scale: 0.9 }}
                        >
                          <Filter size={16} />
                          {txFilter !== 'all' && <span className="filter-badge">{txFilter === 'credit' ? 'In' : 'Out'}</span>}
                        </motion.button>
                        <AnimatePresence>
                          {showTxFilter && (
                            <motion.div 
                              className="filter-menu"
                              initial={{ opacity: 0, scale: 0.95, y: -10 }}
                              animate={{ opacity: 1, scale: 1, y: 0 }}
                              exit={{ opacity: 0, scale: 0.95, y: -10 }}
                              transition={{ duration: 0.2 }}
                            >
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
                            </motion.div>
                          )}
                        </AnimatePresence>
                      </div>
                    </div>
                    <motion.span 
                      className="collapse-icon"
                      animate={{ rotate: collapsedSections['transactions'] ? 0 : 180 }}
                      transition={{ duration: 0.3 }}
                    >
                      <ChevronDown size={20} />
                    </motion.span>
                  </div>
                </motion.div>
                <AnimatePresence>
                  {!collapsedSections['transactions'] && (dashboardData.transactions.length > 0 ? (
                    <motion.div 
                      className="transactions-list"
                      initial={{ opacity: 0, height: 0 }}
                      animate={{ opacity: 1, height: 'auto' }}
                      exit={{ opacity: 0, height: 0 }}
                      transition={{ duration: 0.3 }}
                    >
                    {dashboardData.transactions
                      .filter(tx => txFilter === 'all' || tx.direction === txFilter)
                      .slice(0, 10)
                      .map((tx, index) => (
                      <motion.div 
                        key={tx.id} 
                        className="transaction-item"
                        initial={{ opacity: 0, x: -20 }}
                        animate={{ opacity: 1, x: 0 }}
                        transition={{ delay: index * 0.05 }}
                        whileHover={{ scale: 1.02, x: 5 }}
                        layout
                      >
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
                      </motion.div>
                    ))}
                    </motion.div>
                  ) : (
                    <motion.div 
                      className="empty-state"
                      initial={{ opacity: 0 }}
                      animate={{ opacity: 1 }}
                      exit={{ opacity: 0 }}
                    >
                      <p><Banknote size={18} style={{ marginRight: 8, verticalAlign: 'middle' }} />No transactions yet</p>
                      <p className="hint">Your transaction history will appear here</p>
                    </motion.div>
                  ))}
                </AnimatePresence>
              </motion.section>
            </motion.div>
          </motion.div>
        )}
        </AnimatePresence>

        {/* Analytics Tab */}
        <AnimatePresence mode="wait">
          {dashboardData && activeTab === 'analytics' && (
            <motion.div 
              className="unified-content analytics-view"
              key="analytics"
              initial="initial"
              animate="animate"
              exit="exit"
              variants={fadeInUp}
              transition={{ duration: 0.4 }}
            >
              <div className="analytics-charts-row">
              {/* Expenses by Category - Pie Chart */}
              <motion.section 
                className="dashboard-section analytics-section chart-section"
                variants={scaleIn}
                initial="initial"
                animate="animate"
                transition={{ delay: 0.2 }}
              >
                <div className="section-title-wrapper">
                  <h2 className="section-title"><PiggyBank size={20} style={{ marginRight: 8, verticalAlign: 'middle' }} />Expenses by Category</h2>
                </div>
                <div className="chart-container pie-chart">
                  <div className="pie-chart-wrapper" style={{ position: 'relative' }}>
                    <svg viewBox="0 0 200 200" className="pie-svg">
                      {(() => {
                        const data = mockExpensesByCategory
                        const total = data.reduce((sum, item) => sum + item.value, 0)
                        let currentAngle = 0
                        const colors = ['#FF6D00', '#9BC1F3', '#9E8C78', '#FFB347', '#7E57C2', '#22C55E']
                        
                        return data.map((item, i) => {
                          const angle = (item.value / total) * 360
                          const startAngle = currentAngle
                          currentAngle += angle
                          
                          const x1 = 100 + 80 * Math.cos((startAngle - 90) * Math.PI / 180)
                          const y1 = 100 + 80 * Math.sin((startAngle - 90) * Math.PI / 180)
                          const x2 = 100 + 80 * Math.cos((startAngle + angle - 90) * Math.PI / 180)
                          const y2 = 100 + 80 * Math.sin((startAngle + angle - 90) * Math.PI / 180)
                          const largeArc = angle > 180 ? 1 : 0
                          const percentage = ((item.value / total) * 100).toFixed(1)
                          
                          return (
                            <motion.path
                              key={i}
                              d={`M 100 100 L ${x1} ${y1} A 80 80 0 ${largeArc} 1 ${x2} ${y2} Z`}
                              fill={colors[i % colors.length]}
                              opacity={hoveredSlice?.name === item.name ? 1 : 0.85}
                              stroke="white"
                              strokeWidth="2"
                              initial={{ scale: 0, opacity: 0 }}
                              animate={{ scale: hoveredSlice?.name === item.name ? 1.05 : 1, opacity: hoveredSlice?.name === item.name ? 1 : 0.85 }}
                              transition={{ delay: i * 0.1 + 0.3, duration: 0.5 }}
                              onMouseEnter={() => setHoveredSlice({ name: item.name, value: item.value, percentage })}
                              onMouseLeave={() => setHoveredSlice(null)}
                              style={{ cursor: 'pointer' }}
                            />
                          )
                        })
                      })()}
                    </svg>
                    {hoveredSlice && (
                      <motion.div
                        className="pie-tooltip"
                        initial={{ opacity: 0, scale: 0.8 }}
                        animate={{ opacity: 1, scale: 1 }}
                        exit={{ opacity: 0, scale: 0.8 }}
                      >
                        <div className="tooltip-name">{hoveredSlice.name}</div>
                        <div className="tooltip-value">${formatNumber(hoveredSlice.value)}</div>
                        <div className="tooltip-percentage">{hoveredSlice.percentage}%</div>
                      </motion.div>
                    )}
                  </div>
              </div>
            </motion.section>

            {/* Spending Over Time - Line Chart */}
            <motion.section 
              className="dashboard-section analytics-section chart-section"
              variants={fadeInUp}
              initial="initial"
              animate="animate"
              transition={{ delay: 0.3 }}
            >
              <div className="section-title-wrapper" style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                <h2 className="section-title" style={{ marginBottom: 0 }}><TrendingDown size={20} style={{ marginRight: 8, verticalAlign: 'middle' }} />Spending Trend</h2>
                <motion.button 
                  className="export-btn-small" 
                  onClick={exportReport}
                  whileHover={{ scale: 1.05 }}
                  whileTap={{ scale: 0.95 }}
                  transition={{ type: "spring", stiffness: 400, damping: 17 }}
                >
                  <Folder size={16} />
                  <span>Export</span>
                </motion.button>
              </div>
              <div className="chart-container line-chart">
                  <div className="line-chart-wrapper">
                    <svg viewBox="0 0 600 300" className="line-svg">
                      {(() => {
                        const data = mockSpendingOverTime
                        const maxAmount = Math.max(...data.map(d => d.amount))
                        const padding = 40
                        const width = 600 - padding * 2
                        const height = 300 - padding * 2
                        const stepX = width / (data.length - 1)
                        
                        const points = data.map((d, i) => {
                          const x = padding + i * stepX
                          const y = padding + height - (d.amount / maxAmount) * height
                          return { x, y, amount: d.amount, month: d.month }
                        })
                        
                        const pathD = points.map((p, i) => `${i === 0 ? 'M' : 'L'} ${p.x} ${p.y}`).join(' ')
                        
                        return (
                          <g>
                            {/* Grid lines */}
                            {[0, 1, 2, 3, 4].map((i) => (
                              <line
                                key={`grid-${i}`}
                                x1={padding}
                                y1={padding + (height / 4) * i}
                                x2={padding + width}
                                y2={padding + (height / 4) * i}
                                stroke="#E5E5E5"
                                strokeWidth="1"
                              />
                            ))}
                            
                            {/* Line path */}
                            <motion.path
                              d={pathD}
                              fill="none"
                              stroke="#FF6D00"
                              strokeWidth="3"
                              strokeLinecap="round"
                              initial={{ pathLength: 0 }}
                              animate={{ pathLength: 1 }}
                              transition={{ duration: 1.5, ease: "easeInOut" }}
                            />
                            
                            {/* Area under curve */}
                            <motion.path
                              d={`${pathD} L ${padding + width} ${padding + height} L ${padding} ${padding + height} Z`}
                              fill="url(#gradient)"
                              opacity="0.2"
                              initial={{ opacity: 0 }}
                              animate={{ opacity: 0.2 }}
                              transition={{ duration: 1, delay: 0.5 }}
                            />
                            
                            <defs>
                              <linearGradient id="gradient" x1="0" y1="0" x2="0" y2="1">
                                <stop offset="0%" stopColor="#FF6D00" />
                                <stop offset="100%" stopColor="#FF6D00" stopOpacity="0" />
                              </linearGradient>
                            </defs>
                            
                            {/* Data points */}
                            {points.map((p, i) => (
                              <g key={i}>
                                <motion.circle
                                  cx={p.x}
                                  cy={p.y}
                                  r="5"
                                  fill="#FF6D00"
                                  stroke="white"
                                  strokeWidth="2"
                                  initial={{ scale: 0 }}
                                  animate={{ scale: 1 }}
                                  transition={{ delay: i * 0.1 + 0.5, type: "spring" }}
                                  whileHover={{ scale: 1.5 }}
                                />
                                <text
                                  x={p.x}
                                  y={p.y - 15}
                                  textAnchor="middle"
                                  fontSize="12"
                                  fill="#737373"
                                  fontWeight="500"
                                >
                                  ${(p.amount / 1000).toFixed(1)}k
                                </text>
                                <text
                                  x={p.x}
                                  y={padding + height + 20}
                                  textAnchor="middle"
                                  fontSize="12"
                                  fill="#737373"
                                >
                                  {p.month}
                                </text>
                              </g>
                            ))}
                          </g>
                        )
                      })()}
                    </svg>
                  </div>
              </div>
            </motion.section>
            </div>

            {/* Second row of charts */}
            <div className="analytics-charts-row">
              {/* Income vs Expenses - Bar Chart */}
              <motion.section 
                className="dashboard-section analytics-section chart-section"
                variants={fadeInUp}
                initial="initial"
                animate="animate"
                transition={{ delay: 0.4 }}
              >
                <div className="section-title-wrapper">
                  <h2 className="section-title"><BarChart3 size={20} style={{ marginRight: 8, verticalAlign: 'middle' }} />Income vs Expenses</h2>
                </div>
                <div className="chart-container bar-chart">
                  <div className="bar-chart-wrapper">
                    <div className="income-expense-chart">
                      <div className="chart-y-axis">
                        {['$5k', '$4k', '$3k', '$2k', '$1k', '$0'].map((label, i) => (
                          <span key={i} className="y-label">{label}</span>
                        ))}
                      </div>
                      <div className="chart-bars-area">
                        {mockIncomeVsExpenses.map((d, i) => {
                          const maxAmount = Math.max(...mockIncomeVsExpenses.flatMap(x => [x.income, x.expenses]))
                          const incomePercent = (d.income / maxAmount) * 100
                          const expensePercent = (d.expenses / maxAmount) * 100
                          
                          return (
                            <div key={i} className="bar-group">
                              <div className="bars-wrapper">
                                <motion.div 
                                  className="bar income-bar"
                                  initial={{ height: 0 }}
                                  animate={{ height: `${incomePercent}%` }}
                                  transition={{ duration: 0.8, delay: i * 0.1 }}
                                />
                                <motion.div 
                                  className="bar expense-bar"
                                  initial={{ height: 0 }}
                                  animate={{ height: `${expensePercent}%` }}
                                  transition={{ duration: 0.8, delay: i * 0.1 + 0.05 }}
                                />
                              </div>
                              <span className="month-label">{d.month}</span>
                            </div>
                          )
                        })}
                      </div>
                    </div>
                    <div className="bar-legend">
                      <div className="legend-item">
                        <span className="legend-color" style={{ backgroundColor: '#22C55E' }}></span>
                        <span className="legend-label">Income</span>
                      </div>
                      <div className="legend-item">
                        <span className="legend-color" style={{ backgroundColor: '#FF6D00' }}></span>
                        <span className="legend-label">Expenses</span>
                      </div>
                    </div>
                  </div>
                </div>
              </motion.section>

              {/* Weekly Activity - Vertical Bar Chart */}
              <motion.section 
                className="dashboard-section analytics-section chart-section"
                variants={fadeInUp}
                initial="initial"
                animate="animate"
                transition={{ delay: 0.5 }}
              >
                <div className="section-title-wrapper">
                  <h2 className="section-title"><CreditCard size={20} style={{ marginRight: 8, verticalAlign: 'middle' }} />Weekly Activity</h2>
                </div>
                <div className="chart-container activity-chart">
                  <div className="activity-chart-wrapper">
                    <div className="activity-bars-container">
                      {mockWeeklyActivity.map((d, i) => {
                        const maxTransactions = Math.max(...mockWeeklyActivity.map(x => x.transactions))
                        const heightPercent = (d.transactions / maxTransactions) * 100
                        const intensity = d.transactions / maxTransactions
                        const color = `rgba(255, 109, 0, ${0.4 + intensity * 0.6})`
                        
                        return (
                          <div key={i} className="activity-bar-group">
                            <span className="bar-value">{d.transactions}</span>
                            <motion.div 
                              className="activity-bar"
                              style={{ backgroundColor: color }}
                              initial={{ height: 0 }}
                              animate={{ height: `${heightPercent}%` }}
                              transition={{ duration: 0.6, delay: i * 0.08 }}
                            />
                            <span className="bar-day">{d.day}</span>
                            <span className="bar-amount">${d.amount}</span>
                          </div>
                        )
                      })}
                    </div>
                  </div>
                </div>
              </motion.section>
            </div>
          </motion.div>
        )}
        </AnimatePresence>

        {/* Loading State */}
        <AnimatePresence>
          {isLoading && !dashboardData && (
            <motion.div 
              className="loading-state"
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              exit={{ opacity: 0 }}
            >
              <motion.div 
                className="spinner"
                animate={{ rotate: 360 }}
                transition={{ duration: 1, repeat: Infinity, ease: "linear" }}
              ></motion.div>
              <p>Loading dashboard...</p>
            </motion.div>
          )}
        </AnimatePresence>
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
