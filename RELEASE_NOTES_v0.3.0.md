# 🚀 New Feature: Arbitrary Smart Contract Calls

We just shipped `execute_contract_call` - AI agents now have full access to any smart contract on any blockchain.

## What's new

- Agents can call any contract function (DeFi protocols, NFTs, custom contracts)
- Supports Arbitrum, Base, Ethereum, and any EVM chain
- User confirmation required for all calls
- Available in SDK v0.3.0

## How to use

```bash
go get github.com/becomeliminal/nim-go-sdk@v0.3.0
```

The tool takes:
- `chain_id` - Which chain (42161=Arbitrum, 8453=Base)
- `to` - Contract address
- `data` - Pre-encoded hex calldata
- `value` - Optional ETH amount in wei
- `gas_tier` - Optional (slow/standard/fast)

## Example use cases

- Swap tokens on Uniswap
- Deposit into Aave
- Mint NFTs
- Execute any custom contract logic

Your agent just needs to encode the calldata and our infrastructure handles gas sponsorship, transaction submission, and confirmation.

**Live now in production!** 🎉
