# @becomeliminal/nim-chat

A React chat widget that connects to any nim SDK backend (Go, Python, etc).

## Installation

```bash
npm install @becomeliminal/nim-chat
```

## Quick Start

```tsx
import { NimChat } from '@becomeliminal/nim-chat';
import '@becomeliminal/nim-chat/styles.css';

function App() {
  return (
    <NimChat
      wsUrl="ws://localhost:8080/ws"
      title="Nim"
      position="bottom-right"
    />
  );
}
```

## Props

| Prop | Type | Default | Description |
|------|------|---------|-------------|
| `wsUrl` | `string` | **required** | WebSocket URL to connect to |
| `title` | `string` | `"Nim"` | Header title in the chat panel |
| `position` | `"bottom-right" \| "bottom-left"` | `"bottom-right"` | Widget position |
| `defaultOpen` | `boolean` | `false` | Whether the panel starts open |

## WebSocket Protocol

The widget implements the nim SDK WebSocket protocol:

### Client → Server

| Type | Fields | Description |
|------|--------|-------------|
| `new_conversation` | - | Start a new conversation |
| `resume_conversation` | `conversationId` | Resume existing conversation |
| `message` | `content` | Send a user message |
| `confirm` | `actionId` | Approve a pending action |
| `cancel` | `actionId` | Cancel a pending action |

### Server → Client

| Type | Fields | Description |
|------|--------|-------------|
| `conversation_started` | `conversationId` | New conversation created |
| `conversation_resumed` | `conversationId`, `messages[]` | Conversation restored |
| `text_chunk` | `content` | Streaming text chunk |
| `text` | `content` | Complete message |
| `confirm_request` | `actionId`, `tool`, `summary`, `expiresAt` | Action requires approval |
| `complete` | `tokenUsage?` | Turn completed |
| `error` | `content` | Error occurred |

## Advanced Usage

### Custom Implementation

You can use individual components and the hook for custom implementations:

```tsx
import {
  ChatPanel,
  ChatMessage,
  ChatInput,
  ConfirmationCard,
  ThinkingIndicator,
  useNimWebSocket,
} from '@becomeliminal/nim-chat';
import '@becomeliminal/nim-chat/styles.css';

function CustomChat() {
  const {
    messages,
    isStreaming,
    connectionState,
    confirmationRequest,
    sendMessage,
    confirmAction,
    cancelAction,
  } = useNimWebSocket({
    wsUrl: 'ws://localhost:8080/ws',
    onError: (error) => console.error(error),
  });

  // Build your own UI...
}
```

### Theme Tokens

Access design tokens for consistent styling:

```tsx
import { theme, colors, typography, spacing } from '@becomeliminal/nim-chat';

// colors.orange = '#FF6D00'
// colors.cream = '#F1EDE7'
// etc.
```

## Development

```bash
# Install dependencies
npm install

# Start dev server with example
npm run dev

# Build for production
npm run build

# Type check
npm run typecheck
```

## Testing with nim-go-sdk

1. Start the nim-go-sdk example server:
   ```bash
   cd /path/to/nim-go-sdk/examples/basic
   ANTHROPIC_API_KEY=your-key go run main.go
   ```

2. Run the nim-chat example:
   ```bash
   cd nim-chat
   npm run dev
   ```

3. Open http://localhost:5173 and test:
   - Click bubble → panel opens
   - Send message → streaming response
   - Trigger tool → confirmation card appears
   - Approve/cancel → action executes or cancels
   - Refresh → conversation resumes from localStorage

## Design System

Based on the Liminal mobile app design:

- **Orange** `#FF6D00` - Primary, user bubbles, CTAs
- **Cream** `#F1EDE7` - Background, assistant bubbles
- **Blue** `#9BC1F3` - Focus states, interactive elements
- **Black** `#231F18` - Primary text
- **Brown** `#492610` - Secondary text
- **Green** `#188A31` - Success, approval

## License

MIT
