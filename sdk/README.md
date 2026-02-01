# CertID Sybil Resistance SDK

JavaScript SDK for CertID's Sybil Resistance API. Validate user authenticity and filter bots from your Web3 applications.

## Installation

```bash
npm install @certid/sybil-sdk
# or
yarn add @certid/sybil-sdk
```

## Quick Start

```javascript
const CertIDSybil = require('@certid/sybil-sdk');

const client = new CertIDSybil('YOUR_API_KEY');

// Check a single address
const result = await client.check('cert1abc...');
console.log(`Trust Score: ${result.trust_score}`);
console.log(`Is Human: ${result.is_likely_human}`);
```

## Features

### Single Address Check
```javascript
const result = await client.check('cert1abc...');
// Returns: { address, trust_score, is_likely_human, factors, checked_at }
```

### Batch Check
```javascript
const results = await client.batchCheck(['cert1abc...', 'cert1xyz...'], 60);
// Returns: { results: [...], summary: { total, likely_real, suspicious } }
```

### Filter Real Users
```javascript
const addresses = ['cert1abc...', 'cert1xyz...', 'cert1def...'];
const realUsers = await client.filterReal(addresses, 50);
// Returns only addresses with trust_score >= 50
```

### Filter Suspicious Accounts
```javascript
const suspicious = await client.filterSuspicious(addresses, 30);
// Returns only addresses with trust_score < 30
```

### Get Detailed Factors
```javascript
const factors = await client.getDetailedFactors('cert1abc...');
// Returns: { kyc_verified, social_verifications, onchain_activity, account_age_months, staked_amount, attestations_received }
```

## Trust Score Factors

| Factor | Max Points | Description |
|--------|------------|-------------|
| KYC Verification | 30 | Completed identity verification |
| Social Accounts | 40 | 10 pts per verified platform |
| On-Chain Activity | 20 | 5 pts per transaction |
| Account Age | 10 | 1 pt per month |
| Staked Amount | 20 | 0.01 pts per CERT staked |
| Attestations | 20 | 2 pts per attestation received |

**Total Maximum Score: 100**

## Use Cases

### Airdrop Protection
```javascript
const eligible = await client.filterReal(airdropAddresses, 60);
// Only airdrop to addresses with 60+ trust score
```

### Governance Gating
```javascript
const result = await client.check(voterAddress);
if (result.trust_score >= 70 && result.factors.kyc_verified) {
  allowVote();
}
```

### NFT Minting
```javascript
const result = await client.check(minterAddress);
if (result.is_likely_human) {
  allowMint(1); // Limit mints for verified humans
}
```

## API Reference

### Constructor
```javascript
new CertIDSybil(apiKey, options)
```
- `apiKey` (string): Your CertID API key
- `options.baseUrl` (string): Custom API URL (default: https://api.c3rt.org/api/v1)
- `options.timeout` (number): Request timeout in ms (default: 10000)

### Methods

| Method | Parameters | Returns |
|--------|------------|---------|
| `check(address)` | address: string | SybilCheckResponse |
| `batchCheck(addresses, threshold?)` | addresses: string[], threshold?: number | BatchCheckResponse |
| `filterReal(addresses, minScore?)` | addresses: string[], minScore?: number | string[] |
| `filterSuspicious(addresses, maxScore?)` | addresses: string[], maxScore?: number | string[] |
| `getDetailedFactors(address)` | address: string | TrustFactors |

## Pricing

| Tier | Monthly | Requests/Day | Support |
|------|---------|--------------|---------|
| Free | $0 | 100 | Community |
| Developer | $49 | 10,000 | Email |
| Enterprise | $499 | 1,000,000 | Priority |

Get your API key at [c3rt.org/api-dashboard](https://c3rt.org/api-dashboard)

## Links

- [Documentation](https://c3rt.org/docs)
- [API Dashboard](https://c3rt.org/api-dashboard)
- [Discord](https://discord.gg/certid)

## License

MIT
