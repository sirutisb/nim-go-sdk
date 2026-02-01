import { useState, useEffect } from 'react';
import { ChatPanel } from './components/ChatPanel';
import { LoginPanel } from './components/LoginPanel';
import { useNimWebSocket } from './hooks/useNimWebSocket';
import { getStoredTokens, clearStoredTokens } from './utils/auth';
import type { NimChatProps } from './types';

export function NimChat({
  wsUrl,
  apiUrl = 'https://api.liminal.cash',
  title = 'Nim++',
  position = 'bottom-right',
  defaultOpen = false,
}: NimChatProps) {
  const [isOpen, setIsOpen] = useState(defaultOpen);
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [jwt, setJwt] = useState<string | null>(null);

  // Check for existing auth on mount
  useEffect(() => {
    const tokens = getStoredTokens();
    if (tokens) {
      setJwt(tokens.accessToken);
      setIsAuthenticated(true);
    }
  }, []);

  const handleLoginSuccess = (accessToken: string, userId: string) => {
    setJwt(accessToken);
    setIsAuthenticated(true);
  };

  const handleLogout = () => {
    clearStoredTokens();
    setJwt(null);
    setIsAuthenticated(false);
  };

  const {
    messages,
    isStreaming,
    connectionState,
    confirmationRequest,
    sendMessage,
    confirmAction,
    cancelAction,
  } = useNimWebSocket({
    wsUrl,
    jwt: isAuthenticated ? jwt : null,
    onError: (error) => console.error('[NimChat]', error),
  });

  const positionClasses =
    position === 'bottom-right' ? 'right-4 sm:right-6' : 'left-4 sm:left-6';

  return (
    <div className="nim-chat-widget fixed bottom-4 sm:bottom-6 z-50" style={{ [position === 'bottom-right' ? 'right' : 'left']: '1rem' }}>
      {/* Login Panel */}
      {isOpen && !isAuthenticated && (
        <div
          className={`
            absolute bottom-16
            w-[calc(100vw-2rem)] sm:w-96
          `}
          style={{
            [position === 'bottom-right' ? 'right' : 'left']: 0,
          }}
        >
          <LoginPanel onLoginSuccess={handleLoginSuccess} apiUrl={apiUrl} />
        </div>
      )}

      {/* Chat Panel */}
      {isOpen && isAuthenticated && (
        <div
          className={`
            absolute bottom-16 ${positionClasses}
            w-[calc(100vw-2rem)] sm:w-96
            h-[min(600px,calc(100vh-8rem))]
          `}
          style={{
            [position === 'bottom-right' ? 'right' : 'left']: 0,
          }}
        >
          <ChatPanel
            title={title}
            messages={messages}
            isStreaming={isStreaming}
            connectionState={connectionState}
            confirmationRequest={confirmationRequest}
            onSendMessage={sendMessage}
            onConfirm={confirmAction}
            onCancel={cancelAction}
            onClose={() => setIsOpen(false)}
            onLogout={handleLogout}
          />
        </div>
      )}

      {/* Floating Button */}
      <button
        onClick={() => setIsOpen(!isOpen)}
        className={`
          w-14 h-14
          bg-nim-orange text-white
          rounded-full
          shadow-lg
          flex items-center justify-center
          transition-all duration-200
          hover:scale-110 hover:shadow-xl
          active:scale-95
          ${isOpen ? 'rotate-0' : 'rotate-0'}
        `}
        aria-label={isOpen ? 'Close chat' : 'Open chat'}
      >
        {isOpen ? (
          <svg
            width="24"
            height="24"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            strokeWidth="2"
            strokeLinecap="round"
            strokeLinejoin="round"
          >
            <line x1="18" y1="6" x2="6" y2="18" />
            <line x1="6" y1="6" x2="18" y2="18" />
          </svg>
        ) : (
          <svg
            width="24"
            height="24"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            strokeWidth="2"
            strokeLinecap="round"
            strokeLinejoin="round"
          >
            <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z" />
          </svg>
        )}
      </button>
    </div>
  );
}
