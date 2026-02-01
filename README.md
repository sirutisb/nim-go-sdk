# Preview Video
[![Watch the demo](https://img.youtube.com/vi/yspFSU1v2lw/maxresdefault.jpg)](https://youtu.be/yspFSU1v2lw)

# Nimbus Go SDK

<img width="746" height="827" alt="image" src="https://github.com/user-attachments/assets/1bf90187-f5a5-42a8-8f3e-e6fffafb7077" />


Flow:
Message → Engine → Claude → Tools → Claude → Response




---
**Production-ready SDK for building Claude-powered AI agents in Go**

Enterprise framework for AI agents with conversation management, user confirmations, guardrails, audit logging, and sub-agent delegation.

[![Go Version](https://img.shields.io/badge/go-%3E%3D1.23-blue.svg)](https://golang.org/dl/)

---

## Quick Start

### Installation

#### Backend
```bash
go get github.com/becomeliminal/nim-go-sdk
cd examples/hackathon-starter/
go run .
```
#### Frontend
```bash
cd examples/hackathon-starter/frontend
npm install
npm run dev
```

### Running

```bash
# Set API key
export ANTHROPIC_API_KEY="sk-ant-..."
export PERPLEXITY_API_KEY=""


## What's Included

| Feature | Description |
|---------|-------------|
| **Agent Engine** | Multi-turn conversations with Claude API |
| **Tool System** | Declarative builder with automatic schema generation |
| **Confirmation Flow** | User approval for write operations + idempotency |
| **Guardrails** | Rate limiting and circuit breaker interfaces |
| **Audit Logging** | Complete audit trail for compliance |
| **Sub-Agents** | Hierarchical agent delegation |
| **Storage** | Pluggable conversation/confirmation persistence |
| **WebSocket Server** | Production-ready real-time communication |

---

## Architecture

```
Client → WebSocket → Server → Engine → Claude API
                        ├─ Tool Registry
                        ├─ Store (Conversations, Confirmations)
                        └─ Optional (Guardrails, Audit, Sub-Agents)
```

**Flow:** Message → Load conversation → Call Claude with tools → Execute reads → Request confirmation for writes → User approves → Execute → Continue

---

## Critical Design Analysis

### 1. Confirmation Flow for Write Operations

**Decision:** All write operations require explicit user confirmation.

✅ **Benefits:**
- Prevents AI errors (e.g., sending money to wrong person)
- Regulatory compliance for financial operations
- Idempotency prevents duplicate executions
- User trust and transparency

❌ **Trade-offs:**
- Adds latency (manual approval required)
- More complex UX flow

**Verdict:** Essential for financial/regulated applications. Security > UX cost.

**Rejected Alternatives:**
- Auto-execute with undo → Irreversible financial ops
- Confidence threshold → AI confidence ≠ correctness

---

### 2. Sub-Agent Architecture

**Decision:** Parent agents can delegate to specialized sub-agents.

✅ **Benefits:**
- Modularity and separation of concerns
- ~50% token savings on specialized tasks
- Specialized prompts for domain expertise
- Reusable components

❌ **Trade-offs:**
- Additional API calls increase latency
- Context doesn't flow between agents
- More complex debugging

**When to Use:**
- Deep domain expertise (financial analysis, legal)
- Distinct task phases (research → decide → execute)
- Reusable specialized capabilities

**When NOT to Use:**
- Simple single-turn tasks
- Latency-critical applications
- Tasks requiring full context

---

### 3. Interface-Based Storage

**Decision:** Storage defined as interfaces, not concrete implementations.

✅ **Benefits:**
- Swap backends (Redis, Postgres, DynamoDB)
- Easy testing with in-memory stores
- No vendor lock-in
- Gradual migration path

❌ **Trade-offs:**
- Must implement interfaces for production

**Verdict:** Maximum flexibility. Start in-memory, implement for production.

```go
// Development
store.NewMemoryConversations()

// Production - implement your own
type RedisConversations struct { client *redis.Client }
func (r *RedisConversations) Create(...) (*Conversation, error) { ... }
```

---

### 4. Idempotency Key Design

**Decision:** Keys via `SHA256(userID + tool + input + time_bucket)`.

✅ **Benefits:**
- Deterministic deduplication
- 10-min time-bucketing allows retries
- Cryptographic hash prevents collisions
- Stateless (no coordination needed)

❌ **Trade-offs:**
- 10-min window might be short for some workflows
- Input variations (100 vs 100.00) create different keys

**Verdict:** Simple and effective for 99% of cases. Adjust if needed:

```go
engine.IdempotencyBucketDuration = 30 * time.Minute
```

---

### 5. Streaming vs. Non-Streaming

**Decision:** Support both modes.

✅ **Benefits:**
- Streaming: Better UX (real-time feedback)
- Non-streaming: Simpler testing/batch jobs
- Configurable per deployment

❌ **Trade-offs:**
- Streaming needs SSE-compatible infrastructure
- More complex client code

**Use Cases:**
- **Streaming:** Web/mobile apps, chat interfaces, long responses
- **Non-Streaming:** Batch processing, testing, simple integrations

**Perceived Performance:** Non-streaming waits 5s then shows all. Streaming shows first word in 500ms → Feels 10x faster.

---

### 6. Guardrails as Optional

**Decision:** Guardrails are optional and interface-based.

**Rationale:**
- Not all apps need rate limiting
- Different infrastructure needs different implementations
- Development should work without external deps

⚠️ **Production Warning:** ALWAYS implement guardrails for production:

```go
type RedisGuardrails struct { client *redis.Client }

func (r *RedisGuardrails) Check(ctx context.Context, userID string) (*GuardrailResult, error) {
    // Rate limit: 100 req/hour
    if r.getRequestCount(userID, time.Hour) >= 100 {
        return &GuardrailResult{Allowed: false, Warning: "Rate limit"}, nil
    }

    // Circuit breaker: Block after 5 failures
    if r.getFailureCount(userID) >= 5 {
        return &GuardrailResult{Allowed: false, CircuitState: "open"}, nil
    }

    return &GuardrailResult{Allowed: true}, nil
}
```

---

## Usage Examples

### Custom Tool

```go
srv.AddTool(tools.New("get_weather").
    Description("Get current weather").
    Schema(tools.ObjectSchema(map[string]interface{}{
        "location": tools.StringProperty("City name"),
    }, "location")).
    HandlerFunc(func(ctx context.Context, input json.RawMessage) (interface{}, error) {
        var params struct { Location string `json:"location"` }
        json.Unmarshal(input, &params)
        return map[string]interface{}{"temp": 72, "condition": "sunny"}, nil
    }).
    Build())
```

### Write Operation (Requires Confirmation)

```go
srv.AddTool(tools.New("send_money").
    Description("Transfer money").
    RequiresConfirmation().  // Triggers user approval
    SummaryTemplate("Send ${{amount}} to {{to}}").
    HandlerFunc(func(ctx context.Context, input json.RawMessage) (interface{}, error) {
        // Execute transfer
        return map[string]interface{}{"success": true, "txId": "tx_123"}, nil
    }).
    Build())
```

### Production Config

```go
srv, _ := server.New(server.Config{
    AnthropicKey:  "sk-ant-...",
    Model:         "claude-sonnet-4-20250514",
    MaxTokens:     4096,

    // Storage
    Conversations: &PostgresConversations{db: db},
    Confirmations: &RedisConfirmations{client: redis},

    // Safety
    Guardrails:    &RedisGuardrails{client: redis},
    AuditLogger:   &PostgresAuditLogger{db: db},

    // Auth
    AuthFunc: func(r *http.Request) (string, error) {
        token := r.URL.Query().Get("token")
        return validateJWT(token).UserID, nil
    },
})
```

---

## WebSocket Protocol

**Client → Server:**
```json
{"type": "new_conversation"}
{"type": "message", "content": "Send $50 to Alice"}
{"type": "confirm", "action_id": "..."}
{"type": "cancel", "action_id": "..."}
```

**Server → Client:**
```json
{"type": "conversation_started", "conversation_id": "..."}
{"type": "text_chunk", "content": "..."}
{"type": "confirm_request", "action_id": "...", "summary": "Send $50 to Alice"}
{"type": "complete", "token_usage": {...}}
```

---

## Performance Tips

**Token Optimization:**
- Use sub-agents: ~50% savings on specialized tasks
- Configure `MaxTurns` based on use case
- Store full history in DB, send recent context to API

**API Optimization:**
```go
// ❌ BAD: Multiple runs
for _, user := range users { agent.SendMoney(user, 10) }

// ✅ GOOD: Single run
agent.Run("Send $10 to Alice, Bob, and Carol")
```

**Recommended Limits:**
- API: 100 req/hour per user
- Tokens: 1M/day per user
- Writes: 10/min per user

---

## Configuration

```go
type Config struct {
    AnthropicKey     string   // Required
    Model            string   // Default: claude-sonnet-4-20250514
    MaxTokens        int64    // Default: 4096
    SystemPrompt     string
    LiminalExecutor  *executor.HTTPExecutor
    AuthFunc         func(*http.Request) (string, error)
    Conversations    store.Conversations
    Confirmations    store.Confirmations
    Guardrails       engine.Guardrails
    AuditLogger      engine.AuditLogger
    DisableStreaming bool
}
```

---

## Environment Variables

| Variable | Required | Default |
|----------|----------|---------|
| `ANTHROPIC_API_KEY` | Yes | - |
| `LIMINAL_BASE_URL` | No | `https://api.liminal.cash` |
| `PORT` | No | `8080` |

---

## Examples

- [**basic**](examples/basic) - Minimal setup
- [**custom-tools**](examples/custom-tools) - Creating tools
- [**full-agent**](examples/full-agent) - Financial assistant
- [**hackathon-starter**](examples/hackathon-starter) - Full demo

---

## License

MIT License - See [LICENSE](LICENSE)

**Built by Team Hedges**
