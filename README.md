# Nim Go SDK

**Production-ready SDK for building Claude-powered AI agents in Go**

A comprehensive framework for building AI-powered financial assistants and conversational agents using Claude. Built with enterprise features including conversation management, user confirmation flows, guardrails, audit logging, and sub-agent delegation.

[![Go Version](https://img.shields.io/badge/go-%3E%3D1.23-blue.svg)](https://golang.org/dl/)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

---

## Table of Contents

- [Overview](#overview)
- [Core Features](#core-features)
- [Architecture](#architecture)
- [Feature Analysis & Trade-offs](#feature-analysis--trade-offs)
- [Quick Start](#quick-start)
- [Usage Examples](#usage-examples)
- [API Reference](#api-reference)
- [Performance Considerations](#performance-considerations)
- [Security](#security)
- [Contributing](#contributing)

---

## Overview

The Nim Go SDK is a production-grade framework for building Claude-powered agents with financial capabilities. It handles the complexity of multi-turn conversations, tool execution, user confirmations, and state management, allowing you to focus on building great user experiences.

### What Makes This Different?

Unlike simple API wrappers, Nim provides:
- **Conversation State Management** - Automatic persistence and restoration of multi-turn conversations
- **Confirmation Workflows** - Built-in approval flows for write operations with idempotency
- **Guardrails System** - Rate limiting, circuit breakers, and abuse prevention
- **Audit Logging** - Complete trail of all agent actions for compliance
- **Sub-Agent Delegation** - Hierarchical agent architecture for complex tasks
- **WebSocket Server** - Production-ready real-time communication layer
- **Tool Builder** - Type-safe, fluent API for creating custom tools

---

## Core Features

### 🎯 Agent Engine

The core execution engine that orchestrates Claude API interactions, tool execution, and conversation flow.

**Key Capabilities:**
- Multi-turn agentic loops with configurable limits
- Streaming and non-streaming response modes
- Automatic token usage tracking
- Context preservation across turns
- Timeout and cancellation support

**Use Case:** Building conversational agents that can handle complex, multi-step tasks requiring multiple tool calls.

### 🔧 Tool System

Declarative tool definition with automatic schema generation and execution handling.

**Features:**
- Fluent builder API for tool creation
- JSON Schema generation from Go structs
- Confirmation-required flag for write operations
- Custom summary templates for user-facing descriptions
- Liminal API integration out of the box

**Example:**
```go
tool := tools.New("send_money").
    Description("Transfer money to another user").
    Schema(tools.ObjectSchema(map[string]interface{}{
        "to": tools.StringProperty("Recipient user ID"),
        "amount": tools.NumberProperty("Amount to send"),
    }, "to", "amount")).
    RequiresConfirmation().
    SummaryTemplate("Send ${{amount}} to {{to}}").
    HandlerFunc(func(ctx context.Context, input json.RawMessage) (interface{}, error) {
        // Execute transfer
    }).
    Build()
```

### ✅ Confirmation Flow

User approval system for write operations with security built-in.

**Security Features:**
- Idempotency keys prevent duplicate executions
- 10-minute time-bucketing for deduplication
- Automatic expiration of pending actions
- Cancel capability for users
- Transaction ID tracking

**Trade-off:** Adds latency to write operations but prevents catastrophic errors like duplicate payments.

### 🛡️ Guardrails

Extensible rate limiting and circuit breaker system.

**Interface Design:**
```go
type Guardrails interface {
    Check(ctx context.Context, userID string) (*GuardrailResult, error)
    RecordSuccess(ctx context.Context, userID string)
    RecordFailure(ctx context.Context, userID string)
}
```

**Built-in NoOpGuardrails** for development; implement Redis-backed version for production.

**Benefits:**
- Prevent abuse and runaway costs
- Protect downstream APIs from overload
- Circuit breaker pattern for failing services
- Per-user rate limiting

### 📝 Audit Logging

Complete audit trail for compliance and debugging.

**What Gets Logged:**
- Tool name and input parameters
- Execution results and errors
- Duration and timestamps
- User ID and session context
- Parent-child relationships for sub-agents

**Use Case:** Financial applications requiring SOC2, GDPR, or PCI compliance.

### 🤖 Sub-Agent Delegation

Hierarchical agent architecture for specialized tasks.

**Pattern:**
```go
// Create specialist agent
analyst := subagent.New(subagent.Config{
    Name: "financial_analyst",
    SystemPrompt: "You are a financial analysis specialist...",
    Tools: []string{"research_stock", "get_market_data"},
})

// Wrap as a tool
delegationTool := subagent.NewDelegationTool(subagent.DelegationConfig{
    SubAgent: analyst,
})

// Register with parent agent
parentAgent.AddTool(delegationTool)
```

**Benefits:**
- **Modularity** - Separate concerns into specialized agents
- **Token Efficiency** - Sub-agents use smaller context windows
- **Turn Limits** - Prevent runaway loops with per-agent limits
- **Audit Chain** - Parent-child relationship tracking

**Trade-offs:**
- Additional API calls increase latency
- Context loss between parent and child
- Increased complexity in debugging

### 💾 Storage Abstraction

Clean interfaces for conversation and confirmation persistence.

**Interfaces:**
```go
type Conversations interface {
    Create(ctx context.Context, userID string) (*Conversation, error)
    Get(ctx context.Context, id string) (*Conversation, error)
    Append(ctx context.Context, msg *AppendMessage) error
}

type Confirmations interface {
    Store(ctx context.Context, action *PendingAction) error
    Confirm(ctx context.Context, userID, actionID string) (*PendingAction, error)
    Cancel(ctx context.Context, userID, actionID string) error
}
```

**Default:** In-memory stores for development
**Production:** Implement with Redis, PostgreSQL, or DynamoDB

### 🌐 WebSocket Server

Production-ready real-time communication layer.

**Message Types:**
- `new_conversation` - Start a new conversation
- `resume_conversation` - Restore conversation from storage
- `message` - Send user message
- `confirm` - Approve pending action
- `cancel` - Reject pending action

**Server Features:**
- Automatic JWT extraction and forwarding
- Streaming text chunks for real-time UX
- Session management per connection
- Health check endpoint
- CORS support

---

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│  Client Application (Web, Mobile, Desktop)                     │
│                                                                 │
│  User: "Send £50 to Sarah for dinner"                          │
└──────────────────────┬──────────────────────────────────────────┘
                       │ WebSocket /ws
                       │ JWT Authentication
                       ▼
┌─────────────────────────────────────────────────────────────────┐
│  Nim Go SDK Server                                              │
│                                                                 │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐          │
│  │   Server     │  │    Store     │  │   Engine     │          │
│  │ (WebSocket)  │──│ Conversation │──│  Agent Loop  │          │
│  │              │  │ Confirmation │  │              │          │
│  └──────────────┘  └──────────────┘  └──────┬───────┘          │
│                                              │                  │
│  ┌──────────────────────────────────────────┼────────────────┐ │
│  │                Tool Registry              │                │ │
│  │                                           │                │ │
│  │  ┌──────────┐  ┌──────────┐  ┌──────────┴─────┐           │ │
│  │  │ Liminal  │  │  Custom  │  │   Sub-Agent    │           │ │
│  │  │  Tools   │  │  Tools   │  │   Delegation   │           │ │
│  │  └──────────┘  └──────────┘  └────────────────┘           │ │
│  └─────────────────────────────────────────────────────────────┘ │
│                                                                 │
│  Optional:                                                      │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐          │
│  │  Guardrails  │  │ Audit Logger │  │   Executor   │          │
│  │ Rate Limiter │  │  Compliance  │  │  HTTP/gRPC   │          │
│  └──────────────┘  └──────────────┘  └──────────────┘          │
└─────────────────────┬──────────────────────┬────────────────────┘
                      │                      │
                      ▼                      ▼
         ┌────────────────────┐   ┌────────────────────┐
         │   Claude API       │   │   External APIs    │
         │   (Anthropic)      │   │   (Liminal, etc)   │
         │                    │   │                    │
         │  - Tool Selection  │   │  - Balances        │
         │  - Response Gen    │   │  - Transactions    │
         │  - Streaming       │   │  - Transfers       │
         └────────────────────┘   └────────────────────┘
```

### Data Flow

1. **User Message** → WebSocket Server
2. **Server** → Authenticate, restore conversation from Store
3. **Engine** → Build Messages API request with tools and history
4. **Claude API** → Select tools and generate response
5. **Engine** → Execute read-only tools immediately
6. **Engine** → Pause on write tools, request confirmation
7. **Server** → Send confirmation request to user
8. **User** → Approve or cancel
9. **Server** → Execute confirmed tool via Engine
10. **Engine** → Continue conversation with tool result

---

## Feature Analysis & Trade-offs

### 1. Confirmation Flow for Write Operations

**Design Decision:** All write operations (money transfers, data mutations) require explicit user confirmation.

**Benefits:**
- Prevents AI-initiated errors (e.g., sending money to wrong person)
- Regulatory compliance for financial operations
- User trust and transparency
- Idempotency prevents duplicate executions

**Trade-offs:**
- ❌ Increased latency (user must manually approve)
- ❌ More complex UX flow
- ✅ Prevents catastrophic errors worth the cost
- ✅ Required for financial/regulated applications

**Alternatives Considered:**
- **Auto-execute with undo** - Rejected due to irreversibility of financial operations
- **Confidence threshold** - Rejected as AI confidence doesn't guarantee correctness
- **Dry-run mode** - Complementary, but still requires confirmation

### 2. Sub-Agent Architecture

**Design Decision:** Support hierarchical agents where parent agents can delegate to specialized sub-agents.

**Benefits:**
- Modularity and separation of concerns
- Smaller context windows per agent (lower cost)
- Specialized prompts for domain expertise
- Reusable agent components

**Trade-offs:**
- ❌ Additional API calls increase latency
- ❌ Context doesn't automatically flow between agents
- ❌ More complex to debug
- ✅ Better token efficiency on complex tasks
- ✅ Cleaner separation of concerns

**When to Use:**
- Tasks requiring deep domain expertise (e.g., financial analysis, legal review)
- Operations with distinct phases (research → decision → execution)
- Reusable specialized capabilities across multiple parent agents

**When NOT to Use:**
- Simple single-turn tasks
- Latency-critical applications
- Tasks requiring full conversation context

### 3. Interface-Based Storage

**Design Decision:** Define storage as interfaces rather than concrete implementations.

**Benefits:**
- Flexibility to swap backends (Redis, Postgres, DynamoDB)
- Easy testing with in-memory implementations
- No vendor lock-in
- Gradual migration path (start in-memory, move to persistent)

**Trade-offs:**
- ❌ Requires implementing interfaces for production
- ✅ Maximum flexibility and testability
- ✅ Avoids bloating SDK with dependencies

**Implementation Guide:**
```go
// Development: Use built-in in-memory stores
store.NewMemoryConversations()
store.NewMemoryConfirmations()

// Production: Implement interfaces with your infrastructure
type RedisConversations struct {
    client *redis.Client
}

func (r *RedisConversations) Create(ctx context.Context, userID string) (*Conversation, error) {
    // Your Redis implementation
}
```

### 4. Idempotency Key Design

**Design Decision:** Generate idempotency keys using SHA256(userID + tool + input + time_bucket).

**Benefits:**
- Deterministic deduplication
- Time-bucketing (10min) allows retries while preventing duplicates
- Cryptographic hash prevents collisions
- No external coordination required

**Trade-offs:**
- ❌ 10-minute window might be too short for some workflows
- ❌ Input changes (e.g., amount: 100 → 100.00) create new keys
- ✅ Simple and stateless
- ✅ Sufficient for most financial use cases

**Configuration:**
```go
// Default: 10 minutes
engine.IdempotencyBucketDuration = 10 * time.Minute

// For longer workflows, increase bucket duration
// Note: Longer windows increase duplicate risk
```

### 5. Streaming vs. Non-Streaming

**Design Decision:** Support both streaming and non-streaming modes.

**Benefits:**
- Streaming provides better UX (real-time feedback)
- Non-streaming simpler for testing and batch processing
- Configuration option allows per-deployment choice

**Trade-offs:**
- ❌ Streaming requires SSE-compatible infrastructure
- ❌ More complex client implementation
- ✅ Better perceived performance
- ✅ Essential for conversational UX

**When to Use:**
- **Streaming:** Web/mobile apps, chat interfaces, long responses
- **Non-Streaming:** Batch processing, testing, simple integrations

### 6. Guardrails as Optional

**Design Decision:** Guardrails are optional and interface-based.

**Rationale:**
- Not all applications need rate limiting
- Different infrastructure requires different implementations
- Development should work without external dependencies

**Production Recommendation:**
```go
// ALWAYS implement guardrails for production
type RedisGuardrails struct {
    client *redis.Client
}

func (r *RedisGuardrails) Check(ctx context.Context, userID string) (*GuardrailResult, error) {
    // Rate limiting: 100 requests per hour
    count := r.getRequestCount(userID, time.Hour)
    if count >= 100 {
        return &GuardrailResult{
            Allowed: false,
            Warning: "Rate limit exceeded",
            RetryAfter: time.Now().Add(time.Hour).Unix(),
        }, nil
    }

    // Circuit breaker: Block after 5 consecutive failures
    failures := r.getFailureCount(userID)
    if failures >= 5 {
        return &GuardrailResult{
            Allowed: false,
            CircuitState: "open",
            Warning: "Circuit breaker triggered",
        }, nil
    }

    return &GuardrailResult{Allowed: true}, nil
}
```

---

## Quick Start

### Installation

```bash
go get github.com/becomeliminal/nim-go-sdk
```

### Basic Agent

```go
package main

import (
    "log"
    "github.com/becomeliminal/nim-go-sdk/server"
)

func main() {
    srv, err := server.New(server.Config{
        AnthropicKey: "sk-ant-...",
    })
    if err != nil {
        log.Fatal(err)
    }

    srv.Run(":8080")
}
```

### With Custom Tools

```go
package main

import (
    "context"
    "encoding/json"
    "log"

    "github.com/becomeliminal/nim-go-sdk/core"
    "github.com/becomeliminal/nim-go-sdk/server"
    "github.com/becomeliminal/nim-go-sdk/tools"
)

func main() {
    srv, err := server.New(server.Config{
        AnthropicKey: "sk-ant-...",
    })
    if err != nil {
        log.Fatal(err)
    }

    // Add custom tool
    srv.AddTool(createWeatherTool())

    srv.Run(":8080")
}

func createWeatherTool() core.Tool {
    return tools.New("get_weather").
        Description("Get current weather for a location").
        Schema(tools.ObjectSchema(map[string]interface{}{
            "location": tools.StringProperty("City name"),
        }, "location")).
        HandlerFunc(func(ctx context.Context, input json.RawMessage) (interface{}, error) {
            var params struct {
                Location string `json:"location"`
            }
            json.Unmarshal(input, &params)

            // Call weather API
            return map[string]interface{}{
                "temperature": 72,
                "condition": "sunny",
                "location": params.Location,
            }, nil
        }).
        Build()
}
```

---

## Usage Examples

### 1. Financial Assistant with Liminal Integration

```go
liminalExecutor := executor.NewHTTPExecutor(executor.HTTPExecutorConfig{
    BaseURL: "https://api.liminal.cash",
})

srv, _ := server.New(server.Config{
    AnthropicKey:    "sk-ant-...",
    LiminalExecutor: liminalExecutor,
    SystemPrompt:    financialAssistantPrompt,
})

// Add Liminal tools (balance, transactions, transfers, etc.)
srv.AddTools(tools.LiminalTools(liminalExecutor)...)

srv.Run(":8080")
```

**Features:**
- Check balance: "What's my balance?"
- Send money: "Send £50 to Sarah" (requires confirmation)
- Transaction history: "Show my recent transactions"
- Savings: "Deposit £100 into savings"

### 2. Agent with Sub-Agents

```go
// Create specialist sub-agents
analyst := subagent.NewFinancialAnalyst(engine, registry)
researcher := subagent.NewResearcher(engine, registry)

// Create delegation tools
srv.AddTool(subagent.DelegationToolFromAgent(analyst))
srv.AddTool(subagent.DelegationToolFromAgent(researcher))

// Now parent agent can delegate:
// User: "Should I invest in Apple stock?"
// Agent: Uses researcher to get data, then analyst to provide recommendation
```

### 3. With Guardrails and Audit Logging

```go
// Implement production guardrails
guardrails := &RedisGuardrails{client: redisClient}

// Implement audit logging
auditLogger := &PostgresAuditLogger{db: postgresDB}

srv, _ := server.New(server.Config{
    AnthropicKey: "sk-ant-...",
    Guardrails:   guardrails,
    AuditLogger:  auditLogger,
})

srv.Run(":8080")
```

### 4. Custom Storage Backends

```go
// Implement conversation storage with PostgreSQL
conversations := &PostgresConversations{db: db}
confirmations := &RedisConfirmations{client: redisClient}

srv, _ := server.New(server.Config{
    AnthropicKey:  "sk-ant-...",
    Conversations: conversations,
    Confirmations: confirmations,
})
```

---

## API Reference

### Server Configuration

```go
type Config struct {
    AnthropicKey     string              // Required: Anthropic API key
    BaseURL          string              // Optional: Custom API endpoint
    SystemPrompt     string              // Optional: Agent system prompt
    Model            string              // Optional: Claude model (default: claude-sonnet-4-20250514)
    MaxTokens        int64               // Optional: Max response tokens (default: 4096)
    LiminalExecutor  *executor.HTTPExecutor
    AuthFunc         func(*http.Request) (string, error)
    Conversations    store.Conversations
    Confirmations    store.Confirmations
    Guardrails       engine.Guardrails
    AuditLogger      engine.AuditLogger
    DisableStreaming bool                // Disable streaming for testing
}
```

### Tool Builder

```go
tools.New("tool_name").
    Description("What the tool does").
    Schema(tools.ObjectSchema(map[string]interface{}{
        "param": tools.StringProperty("Parameter description"),
    }, "param")).
    RequiresConfirmation().  // Optional: for write operations
    SummaryTemplate("Do {{action}} with {{param}}").
    HandlerFunc(func(ctx context.Context, input json.RawMessage) (interface{}, error) {
        // Implementation
    }).
    Build()
```

### Agent Capabilities

```go
type Capabilities struct {
    CanRequestConfirmation bool
    AvailableTools         []string
    Model                  string
    MaxTokens              int64
    MaxTurns               int
    SystemPrompt           string
}

// Defaults
core.DefaultCapabilities()      // Standard agent
core.SubAgentCapabilities()     // Sub-agent (no confirmations)
```

### WebSocket Protocol

**Client → Server Messages:**
```json
{"type": "new_conversation"}
{"type": "resume_conversation", "conversation_id": "..."}
{"type": "message", "content": "Send $50 to Alice"}
{"type": "confirm", "action_id": "..."}
{"type": "cancel", "action_id": "..."}
```

**Server → Client Messages:**
```json
{"type": "conversation_started", "conversation_id": "..."}
{"type": "text_chunk", "content": "..."}
{"type": "text", "content": "Complete message"}
{"type": "confirm_request", "action_id": "...", "summary": "Send $50 to Alice"}
{"type": "complete", "token_usage": {...}}
{"type": "error", "content": "Error message"}
```

---

## Performance Considerations

### Token Optimization

**Sub-Agent Strategy:**
- Parent agent: 4096 max tokens, 20 turns
- Sub-agent: 2048 max tokens, 10 turns
- **Savings:** ~50% on specialized tasks

**History Management:**
- Conversations automatically truncate old messages
- Store full history in database, send recent context to API
- Configure `MaxTurns` based on use case

### API Call Optimization

**Batching:**
```go
// BAD: Sequential tool calls
for _, user := range users {
    agent.SendMoney(user, amount)  // One API call each
}

// GOOD: Single agent invocation
agent.Run("Send $10 to Alice, Bob, and Carol")  // One agentic loop
```

**Caching:**
- Implement caching in tool handlers for expensive operations
- Cache user lookups, market data, etc.
- Use idempotency keys to prevent duplicate executions

### Streaming Benefits

**Perceived Performance:**
- Non-streaming: User waits 5s, sees complete response
- Streaming: User sees first word in 500ms, total still 5s
- **User experience:** Feels 10x faster with streaming

---

## Security

### Authentication

**JWT Flow:**
```go
// Client includes JWT in WebSocket connection
ws://localhost:8080/ws?token=eyJ...

// Server extracts and validates
AuthFunc: func(r *http.Request) (string, error) {
    token := r.URL.Query().Get("token")
    claims := validateJWT(token)  // Your implementation
    return claims.UserID, nil
}
```

### Confirmation Security

**Idempotency Protection:**
- Same action within 10-minute window uses same idempotency key
- Prevents duplicate confirmations from UI bugs
- Time-bucketing allows legitimate retries

**Expiration:**
- Pending actions expire after 10 minutes
- Prevents stale confirmations
- User must re-request if expired

### Rate Limiting

**Recommended Limits:**
- **API Calls:** 100 requests/hour per user
- **Token Usage:** 1M tokens/day per user
- **Write Operations:** 10 confirmations/minute per user

**Circuit Breaker:**
- Open circuit after 5 consecutive failures
- Half-open after 60 seconds
- Reset on successful request

---

## Environment Variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `ANTHROPIC_API_KEY` | Yes | - | Anthropic API key |
| `LIMINAL_BASE_URL` | No | `https://api.liminal.cash` | Liminal API endpoint |
| `PORT` | No | `8080` | Server port |

---

## Examples Directory

- **[basic](examples/basic)** - Minimal agent setup
- **[custom-tools](examples/custom-tools)** - Creating custom tools
- **[full-agent](examples/full-agent)** - Complete financial assistant
- **[hackathon-starter](examples/hackathon-starter)** - Full-featured demo app

---

## Roadmap

- [ ] Anthropic Prompt Caching support
- [ ] Multi-modal support (image understanding)
- [ ] Built-in Redis/Postgres implementations
- [ ] Agent analytics and observability
- [ ] Multi-agent orchestration patterns
- [ ] gRPC server option
- [ ] Function result caching

---

## Contributing

Contributions welcome! Please read [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

**Areas for Contribution:**
- Storage backend implementations (Redis, Postgres, DynamoDB)
- Guardrails implementations (Redis rate limiter)
- Additional pre-built tools
- Documentation improvements
- Example applications

---

## License

MIT License - see [LICENSE](LICENSE) for details.

---

## Team

Built by **Team Hedges** for the future of AI-powered banking.

---

## Acknowledgments

- [Anthropic](https://anthropic.com) for Claude API
- [Liminal](https://liminal.cash) for financial infrastructure
- Go community for excellent tooling
