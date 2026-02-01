import { useState, useEffect, useRef, useCallback } from 'react';
import { ChatPanel } from './components/ChatPanel';
import { LoginPanel } from './components/LoginPanel';
import { useNimWebSocket } from './hooks/useNimWebSocket';
import { getStoredTokens, clearStoredTokens } from './utils/auth';
import type { NimChatProps } from './types';

// Custom hook for drag functionality
function useDraggable(initialPosition: { x: number; y: number }) {
  const [position, setPosition] = useState(initialPosition);
  const [isDragging, setIsDragging] = useState(false);
  const dragOffset = useRef({ x: 0, y: 0 });

  const handleMouseDown = useCallback((e: React.MouseEvent) => {
    // Only allow dragging from the header area
    if ((e.target as HTMLElement).closest('[data-drag-handle]')) {
      setIsDragging(true);
      dragOffset.current = {
        x: e.clientX - position.x,
        y: e.clientY - position.y,
      };
      e.preventDefault();
    }
  }, [position]);

  const handleMouseMove = useCallback((e: MouseEvent) => {
    if (isDragging) {
      const newX = e.clientX - dragOffset.current.x;
      const newY = e.clientY - dragOffset.current.y;

      // Constrain to viewport
      const maxX = window.innerWidth - 100;
      const maxY = window.innerHeight - 100;

      setPosition({
        x: Math.max(0, Math.min(newX, maxX)),
        y: Math.max(0, Math.min(newY, maxY)),
      });
    }
  }, [isDragging]);

  const handleMouseUp = useCallback(() => {
    setIsDragging(false);
  }, []);

  useEffect(() => {
    if (isDragging) {
      document.addEventListener('mousemove', handleMouseMove);
      document.addEventListener('mouseup', handleMouseUp);
      return () => {
        document.removeEventListener('mousemove', handleMouseMove);
        document.removeEventListener('mouseup', handleMouseUp);
      };
    }
  }, [isDragging, handleMouseMove, handleMouseUp]);

  return { position, setPosition, isDragging, handleMouseDown };
}

// Custom hook for resize functionality
function useResizable(initialSize: { width: number; height: number }, minSize = { width: 320, height: 400 }) {
  const [size, setSize] = useState(initialSize);
  const [isResizing, setIsResizing] = useState(false);
  const resizeDirection = useRef<string>('');
  const startPos = useRef({ x: 0, y: 0 });
  const startSize = useRef({ width: 0, height: 0 });

  const handleResizeStart = useCallback((e: React.MouseEvent, direction: string) => {
    setIsResizing(true);
    resizeDirection.current = direction;
    startPos.current = { x: e.clientX, y: e.clientY };
    startSize.current = { ...size };
    e.preventDefault();
    e.stopPropagation();
  }, [size]);

  const handleMouseMove = useCallback((e: MouseEvent) => {
    if (!isResizing) return;

    const deltaX = e.clientX - startPos.current.x;
    const deltaY = e.clientY - startPos.current.y;
    const dir = resizeDirection.current;

    let newWidth = startSize.current.width;
    let newHeight = startSize.current.height;

    if (dir.includes('e')) newWidth = startSize.current.width + deltaX;
    if (dir.includes('w')) newWidth = startSize.current.width - deltaX;
    if (dir.includes('s')) newHeight = startSize.current.height + deltaY;
    if (dir.includes('n')) newHeight = startSize.current.height - deltaY;

    setSize({
      width: Math.max(minSize.width, Math.min(newWidth, window.innerWidth - 40)),
      height: Math.max(minSize.height, Math.min(newHeight, window.innerHeight - 100)),
    });
  }, [isResizing, minSize]);

  const handleMouseUp = useCallback(() => {
    setIsResizing(false);
  }, []);

  useEffect(() => {
    if (isResizing) {
      document.addEventListener('mousemove', handleMouseMove);
      document.addEventListener('mouseup', handleMouseUp);
      document.body.style.cursor = resizeDirection.current.includes('e') || resizeDirection.current.includes('w')
        ? (resizeDirection.current.includes('n') || resizeDirection.current.includes('s') ? 'nwse-resize' : 'ew-resize')
        : 'ns-resize';
      document.body.style.userSelect = 'none';
      return () => {
        document.removeEventListener('mousemove', handleMouseMove);
        document.removeEventListener('mouseup', handleMouseUp);
        document.body.style.cursor = '';
        document.body.style.userSelect = '';
      };
    }
  }, [isResizing, handleMouseMove, handleMouseUp]);

  return { size, setSize, isResizing, handleResizeStart };
}

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

  // Calculate initial position based on screen size
  const getInitialPosition = () => ({
    x: position === 'bottom-right' ? window.innerWidth - 420 : 20,
    y: window.innerHeight - 680,
  });

  const { position: panelPosition, isDragging, handleMouseDown, setPosition } = useDraggable(getInitialPosition());
  const { size, isResizing, handleResizeStart } = useResizable({ width: 384, height: 600 });

  // Reset position when opening
  useEffect(() => {
    if (isOpen) {
      const newPos = getInitialPosition();
      setPosition({
        x: Math.max(20, Math.min(newPos.x, window.innerWidth - size.width - 20)),
        y: Math.max(20, Math.min(newPos.y, window.innerHeight - size.height - 80)),
      });
    }
  }, [isOpen]);

  // Check for existing auth on mount
  useEffect(() => {
    const tokens = getStoredTokens();
    if (tokens) {
      setJwt(tokens.accessToken);
      setIsAuthenticated(true);
    }
  }, []);

  const handleLoginSuccess = (accessToken: string, _userId: string) => {
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
    clearMessages,
  } = useNimWebSocket({
    wsUrl,
    jwt: isAuthenticated ? jwt : null,
    onError: (error) => console.error('[NimChat]', error),
  });

  // Resize handles component
  const ResizeHandles = () => (
    <>
      {/* Edge handles */}
      <div
        className="absolute top-0 left-2 right-2 h-1 cursor-ns-resize hover:bg-nim-orange/30 transition-colors"
        onMouseDown={(e) => handleResizeStart(e, 'n')}
      />
      <div
        className="absolute bottom-0 left-2 right-2 h-1 cursor-ns-resize hover:bg-nim-orange/30 transition-colors"
        onMouseDown={(e) => handleResizeStart(e, 's')}
      />
      <div
        className="absolute left-0 top-2 bottom-2 w-1 cursor-ew-resize hover:bg-nim-orange/30 transition-colors"
        onMouseDown={(e) => handleResizeStart(e, 'w')}
      />
      <div
        className="absolute right-0 top-2 bottom-2 w-1 cursor-ew-resize hover:bg-nim-orange/30 transition-colors"
        onMouseDown={(e) => handleResizeStart(e, 'e')}
      />
      {/* Corner handles */}
      <div
        className="absolute top-0 left-0 w-3 h-3 cursor-nwse-resize"
        onMouseDown={(e) => handleResizeStart(e, 'nw')}
      />
      <div
        className="absolute top-0 right-0 w-3 h-3 cursor-nesw-resize"
        onMouseDown={(e) => handleResizeStart(e, 'ne')}
      />
      <div
        className="absolute bottom-0 left-0 w-3 h-3 cursor-nesw-resize"
        onMouseDown={(e) => handleResizeStart(e, 'sw')}
      />
      <div
        className="absolute bottom-0 right-0 w-3 h-3 cursor-nwse-resize"
        onMouseDown={(e) => handleResizeStart(e, 'se')}
      />
    </>
  );

  return (
    <>
      {/* Login Panel - Fixed position */}
      {isOpen && !isAuthenticated && (
        <div
          className="fixed z-50"
          style={{
            left: panelPosition.x,
            top: panelPosition.y,
            width: size.width,
          }}
          onMouseDown={handleMouseDown}
        >
          <div data-drag-handle className="cursor-move">
            <LoginPanel onLoginSuccess={handleLoginSuccess} apiUrl={apiUrl} />
          </div>
        </div>
      )}

      {/* Chat Panel - Draggable & Resizable */}
      {isOpen && isAuthenticated && (
        <div
          className={`fixed z-50 ${isDragging || isResizing ? '' : 'transition-shadow'}`}
          style={{
            left: panelPosition.x,
            top: panelPosition.y,
            width: size.width,
            height: size.height,
          }}
          onMouseDown={handleMouseDown}
        >
          <div className="relative w-full h-full">
            <ResizeHandles />
            <div
              className="w-full h-full"
              style={{
                cursor: isDragging ? 'grabbing' : undefined,
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
                onClearMessages={clearMessages}
              />
            </div>
          </div>
        </div>
      )}

      {/* Floating Button */}
      <div className="nim-chat-widget fixed bottom-4 sm:bottom-6 z-40" style={{ [position === 'bottom-right' ? 'right' : 'left']: '1rem' }}>
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
    </>
  );
}
