// Liminal API client for making direct API calls
// These endpoints are the same ones the AI agent uses

const LIMINAL_API_URL = process.env.NEXT_PUBLIC_API_URL || 'https://api.liminal.cash'

interface LiminalResponse<T> {
  success: boolean
  data?: T
  error?: string
}

export interface BalanceData {
  balance: {
    currency: string
    amount: string
  }[]
}

export interface Transaction {
  id: string
  type: 'send' | 'receive' | 'deposit' | 'withdraw'
  amount: string
  currency: string
  timestamp: string
  counterparty?: string
  note?: string
  status: string
}

export interface TransactionsData {
  transactions: Transaction[]
}

/**
 * Get the user's wallet balance
 * Calls the same endpoint that the AI uses: /nim/v1/agent/wallet/balance
 */
export async function getBalance(jwtToken: string): Promise<BalanceData> {
  const response = await fetch(`${LIMINAL_API_URL}/nim/v1/agent/wallet/balance`, {
    method: 'GET',
    headers: {
      'Authorization': `Bearer ${jwtToken}`,
      'Content-Type': 'application/json',
    },
  })

  if (!response.ok) {
    throw new Error(`Failed to fetch balance: ${response.status} ${response.statusText}`)
  }

  return await response.json()
}

/**
 * Get the user's transaction history
 * Calls the same endpoint that the AI uses: /nim/v1/agent/transactions
 */
export async function getTransactions(jwtToken: string, limit: number = 10): Promise<TransactionsData> {
  const response = await fetch(`${LIMINAL_API_URL}/nim/v1/agent/transactions?limit=${limit}`, {
    method: 'GET',
    headers: {
      'Authorization': `Bearer ${jwtToken}`,
      'Content-Type': 'application/json',
    },
  })

  if (!response.ok) {
    throw new Error(`Failed to fetch transactions: ${response.status} ${response.statusText}`)
  }

  return await response.json()
}

/**
 * Get JWT token from localStorage
 * NimChat stores the JWT token after authentication
 */
export function getJWTToken(): string | null {
  if (typeof window === 'undefined') return null
  return localStorage.getItem('liminal_jwt_token') || localStorage.getItem('jwt_token')
}
