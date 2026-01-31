'use client'

import dynamic from 'next/dynamic'
import '@liminalcash/nim-chat/styles.css'

const NimChat = dynamic(
  () => import('@liminalcash/nim-chat').then((mod) => ({ default: mod.NimChat })),
  { ssr: false }
)

export default function NimChatWrapper() {
  const wsUrl = process.env.NEXT_PUBLIC_WS_URL || 'ws://localhost:8080/ws'
  const apiUrl = process.env.NEXT_PUBLIC_API_URL || 'https://api.liminal.cash'

  return (
    <NimChat
      wsUrl={wsUrl}
      apiUrl={apiUrl}
      title="Nim"
      position="bottom-right"
      defaultOpen={false}
    />
  )
}
