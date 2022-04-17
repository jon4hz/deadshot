# 🦄 Uniswap v2 SDK

An implementation of the uniswap v2 sdk in golang.

## Differences to the original SDK
- Pair addresses aren't generated automatically and must be looked up on-chain to create a pair
- Chain ids aren't checked, it's up to the developer to make sure they match
- No hardcoded implementation of the token from the native currency, must be created like any other token by passing the contract
- Internal changes:
  - currency is an interface representing tokens and the native currency
  - less functions and types are exported

## Link
- https://docs.uniswap.org/sdk/2.0.0/introduction
- https://github.com/miraclesu/uniswap-sdk-go

