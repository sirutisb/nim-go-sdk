# Liminal Hackathon - Building Financial Agency for AI

AGI won't happen without economic agency. An agent that can plan, reason and act but cannot command resources will forever remain a powerless tool. The absence of this financial infrastructure is a bottleneck to AGI.

The banking system was built for humans - with human identity, human legal status, and human custody. The new system needs to account for both humans and their AI agents, with verifiable identities, reputation scores and smart contract custody. **Stablecoins enable this.** They're instant, borderless, programmable money at the wavelength of machines. We're building the financial system for the agentic web.

This hackathon is a deep dive into the future of agentic payments. Come build the open source brain of financial agents.

**🔗 Links:**
- SDK: https://github.com/becomeliminal/nim-go-sdk
- Starter: https://github.com/becomeliminal/nim-go-sdk/tree/master/examples/hackathon-starter

## What is Nim?

Nim is a Go SDK that connects LLMs to a stablecoin payments ledger. It comes with all the primitives for real banking operations: send money, get balance, transactions, savings - built on stablecoin infrastructure.

**What if instead of clicking through payment flows, you could just chat to the internet of finance?**

## What You'll Build

Nim gives agents access to core banking primitives. This hackathon is about expanding what Nim can do:

- **Integrations** - Connect Nim to new financial infrastructure
- **Tools** - Build new capabilities for Nim to understand users and manage money
- **Subagents** - Give Nim a team of specialized agents to coordinate complex tasks
- **Extensions** - Extend Nim's core capabilities (planning strategies, memory patterns, safety)

## Built-in Banking Tools

9 tools out of the box: `get_balance`, `get_transactions`, `send_money`, `deposit_savings`, `withdraw_savings`, and more.

## Custom Tools - Where You Win

Tools are functions that give AI superpowers. Your tools can:
- Call Liminal banking APIs to get real transaction data
- Call ANY third-party API (OpenAI, your own backend, ML models)
- Do local computation (analysis, calculations, predictions)
- Use any Go library you want

The hackathon starter shows how to build a spending analyzer tool. Clone it, customize it, add your own intelligence.

## Frontend

We ship `@liminalcash/nim-chat` on npm - a React component that handles chat UI, auth, and WebSocket. **Totally optional** - build your own if you want, or use ours to move faster.

## Get Started (5 mins)

```bash
git clone https://github.com/becomeliminal/nim-go-sdk.git
cd nim-go-sdk/examples/hackathon-starter
cp .env.example .env  # Add ANTHROPIC_API_KEY
go run main.go

# New terminal
cd frontend && npm install && npm run dev
# Open http://localhost:5173
```

## Project Ideas

Focus on **custom tools**, not reinventing chat:
- Spending analyzer - categorize transactions, find patterns
- Savings coach - track goals, calculate compound interest
- Bill detector - find recurring payments, predict bills
- Financial health score - analyze savings rate and spending
- Tax estimator - track deductibles, estimate quarterly taxes

**The banking APIs give you real data. Your tools turn it into intelligence.**

Drop questions in this channel - let's build! 🚀
