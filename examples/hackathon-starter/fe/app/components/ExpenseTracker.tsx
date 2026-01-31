'use client'

import { useEffect, useState } from 'react'
import { getBalance, getTransactions, getJWTToken, type Transaction } from '@/lib/liminal-api'

interface Balance {
  balance: string
  currency: string
}

export default function ExpenseTracker() {
  const [balance, setBalance] = useState<Balance | null>(null)
  const [transactions, setTransactions] = useState<Transaction[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [monthlySpent, setMonthlySpent] = useState<number>(0)
  const [savingsBalance, setSavingsBalance] = useState<number>(0)

  useEffect(() => {
    const fetchData = async () => {
      try {
        const jwtToken = getJWTToken()
        
        if (!jwtToken) {
          setError('Please log in via Nim Chat first')
          setLoading(false)
          return
        }

        // Fetch balance and transactions using the same endpoints the AI uses
        const [balanceData, transactionsData] = await Promise.all([
          getBalance(jwtToken),
          getTransactions(jwtToken, 100) // Get more transactions to calculate monthly stats
        ])

        // Set balance (use first currency)
        if (balanceData.balance && balanceData.balance.length > 0) {
          const primaryBalance = balanceData.balance[0]
          setBalance({
            balance: parseFloat(primaryBalance.amount).toLocaleString('en-US', {
              minimumFractionDigits: 2,
              maximumFractionDigits: 2
            }),
            currency: primaryBalance.currency
          })
        }

        // Set transactions
        const txns = transactionsData.transactions || []
        setTransactions(txns)
        
        // Calculate monthly spent (outgoing transactions this month)
        const now = new Date()
        const firstDayOfMonth = new Date(now.getFullYear(), now.getMonth(), 1)
        
        const thisMonthSpent = txns
          .filter(tx => {
            const txDate = new Date(tx.timestamp)
            return txDate >= firstDayOfMonth && (tx.type === 'send' || tx.type === 'withdraw')
          })
          .reduce((sum, tx) => sum + Math.abs(parseFloat(tx.amount)), 0)
        
        setMonthlySpent(thisMonthSpent)
        
        // Note: Savings balance would come from get_savings_balance endpoint
        // For now, calculate from deposit/withdraw transactions
        const netSavings = txns
          .filter(tx => tx.type === 'deposit' || tx.type === 'withdraw')
          .reduce((sum, tx) => {
            const amount = parseFloat(tx.amount)
            return tx.type === 'deposit' ? sum + amount : sum - amount
          }, 0)
        
        setSavingsBalance(Math.max(0, netSavings))
        
        setLoading(false)
      } catch (err) {
        console.error('Failed to fetch data:', err)
        setError(err instanceof Error ? err.message : 'Failed to load data')
        setLoading(false)
      }
    }

    fetchData()
    
    // Refresh data every 30 seconds
    const interval = setInterval(fetchData, 30000)
    return () => clearInterval(interval)
  }, [])

  return (
    <div className="min-h-screen bg-[#FFE951] p-4 md:p-8">
      <div className="max-w-6xl mx-auto">
        {/* Header */}
        <div className="mb-8">
          <h1 className="text-5xl md:text-7xl font-black mb-2 text-black border-8 border-black p-6 bg-white inline-block shadow-[8px_8px_0px_0px_rgba(0,0,0,1)]">
            EXPENSE TRACKER
          </h1>
        </div>

        {/* Balance Card */}
        <div className="mb-8">
          <div className="bg-[#FF6B6B] border-8 border-black p-8 shadow-[12px_12px_0px_0px_rgba(0,0,0,1)]">
            <div className="mb-4">
              <span className="text-xl font-black text-black bg-white px-4 py-2 border-4 border-black inline-block">
                CURRENT BALANCE
              </span>
            </div>
            {loading ? (
              <div className="text-6xl font-black text-white animate-pulse">
                LOADING...
              </div>
            ) : error ? (
              <div className="text-4xl font-black text-white">
                {error}
              </div>
            ) : (
              <div className="text-7xl md:text-8xl font-black text-white">
                ${balance?.balance}
              </div>
            )}
            <div className="mt-4 text-xl font-bold text-white">
              {balance?.currency}
            </div>
          </div>
        </div>

        {/* Quick Stats Grid */}
        <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
          {/* This Month */}
          <div className="bg-[#4ECDC4] border-8 border-black p-6 shadow-[8px_8px_0px_0px_rgba(0,0,0,1)]">
            <div className="text-sm font-black text-black mb-2 bg-white px-3 py-1 border-4 border-black inline-block">
              THIS MONTH
            </div>
            <div className="text-4xl font-black text-black mt-4">
              {loading ? '...' : `$${monthlySpent.toLocaleString('en-US', {
                minimumFractionDigits: 2,
                maximumFractionDigits: 2
              })}`}
            </div>
            <div className="text-lg font-bold text-black mt-2">SPENT</div>
          </div>

          {/* Savings */}
          <div className="bg-[#95E1D3] border-8 border-black p-6 shadow-[8px_8px_0px_0px_rgba(0,0,0,1)]">
            <div className="text-sm font-black text-black mb-2 bg-white px-3 py-1 border-4 border-black inline-block">
              SAVINGS
            </div>
            <div className="text-4xl font-black text-black mt-4">
              {loading ? '...' : `$${savingsBalance.toLocaleString('en-US', {
                minimumFractionDigits: 2,
                maximumFractionDigits: 2
              })}`}
            </div>
            <div className="text-lg font-bold text-black mt-2">SAVED</div>
          </div>

          {/* Total Transactions */}
          <div className="bg-[#F38181] border-8 border-black p-6 shadow-[8px_8px_0px_0px_rgba(0,0,0,1)]">
            <div className="text-sm font-black text-black mb-2 bg-white px-3 py-1 border-4 border-black inline-block">
              TRANSACTIONS
            </div>
            <div className="text-4xl font-black text-black mt-4">
              {loading ? '...' : transactions.length}
            </div>
            <div className="text-lg font-bold text-black mt-2">TOTAL</div>
          </div>
        </div>

        {/* Recent Transactions */}
        <div className="bg-white border-8 border-black p-8 shadow-[12px_12px_0px_0px_rgba(0,0,0,1)]">
          <h2 className="text-3xl font-black text-black mb-6 pb-4 border-b-8 border-black">
            RECENT TRANSACTIONS
          </h2>
          {loading ? (
            <div className="text-2xl font-black text-black animate-pulse">
              LOADING TRANSACTIONS...
            </div>
          ) : transactions.length === 0 ? (
            <div className="text-xl font-bold text-black">
              No transactions yet. Start using Nim to send money!
            </div>
          ) : (
            <div className="space-y-4">
              {transactions.slice(0, 10).map((transaction, index) => {
                const isPositive = transaction.type === 'receive'
                const amount = parseFloat(transaction.amount)
                const formattedAmount = `${isPositive ? '+' : '-'}$${Math.abs(amount).toLocaleString('en-US', {
                  minimumFractionDigits: 2,
                  maximumFractionDigits: 2
                })}`
                
                const categoryColors: Record<string, string> = {
                  send: 'bg-[#FF6B6B]',
                  receive: 'bg-[#95E1D3]',
                  deposit: 'bg-[#4ECDC4]',
                  withdraw: 'bg-[#FFE951]',
                }
                
                return (
                  <div
                    key={transaction.id || index}
                    className="flex items-center justify-between p-4 border-4 border-black hover:shadow-[4px_4px_0px_0px_rgba(0,0,0,1)] transition-all"
                  >
                    <div className="flex items-center gap-4">
                      <div className={`${categoryColors[transaction.type]} border-4 border-black px-3 py-1`}>
                        <span className="text-xs font-black text-black">
                          {transaction.type.toUpperCase()}
                        </span>
                      </div>
                      <div>
                        <span className="text-lg font-bold text-black block">
                          {transaction.counterparty || transaction.type.toUpperCase()}
                        </span>
                        {transaction.note && (
                          <span className="text-sm text-gray-600">
                            {transaction.note}
                          </span>
                        )}
                      </div>
                    </div>
                    <span className={`text-2xl font-black ${isPositive ? 'text-green-600' : 'text-black'}`}>
                      {formattedAmount}
                    </span>
                  </div>
                )
              })}
            </div>
          )}
        </div>

        {/* CTA Section */}
        <div className="mt-8 bg-black border-8 border-black p-8 shadow-[12px_12px_0px_0px_rgba(255,233,81,1)]">
          <p className="text-2xl font-black text-white text-center">
            💬 CHAT WITH NIM TO MANAGE YOUR MONEY →
          </p>
          <p className="text-lg font-bold text-white text-center mt-2">
            Click the chat button in the bottom-right corner
          </p>
        </div>
      </div>
    </div>
  )
}
