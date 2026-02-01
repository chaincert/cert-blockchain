# @certid/sybil-sdk

Official Sybil Resistance SDK for CertID - Trust scores and bot detection for Web3.

## Installation

```bash
npm install @certid/sybil-sdk
```

## Quick Start

```javascript
const { CertID } = require('@certid/sybil-sdk');

const certid = new CertID({
  apiKey: 'your-api-key' // Optional for basic usage
});

// Check single address
const result = await certid.checkSybil('0x742d35Cc6634...');
console.log(result.trustScore); // 0-100
console.log(result.isLikelyHuman); // true/false
```

## API

### `checkSybil(address)`

Check trust score for a single address.

```javascript
const result = await certid.checkSybil('cert1abc...');
// {
//   address: 'cert1abc...',
//   trustScore: 85,
//   isLikelyHuman: true,
//   factors: { kycVerified: true, socialVerifications: 3 }
// }
```

### `batchCheck(addresses, options?)`

Batch check up to 100 addresses.

```javascript
const results = await certid.batchCheck(
  ['addr1', 'addr2', 'addr3'],
  { minScore: 50 }
);
const eligible = results.filter(r => r.isLikelyHuman);
```

### `filterReal(addresses, minScore?)`

Filter to only addresses above the trust threshold.

```javascript
// Airdrop protection
const eligible = await certid.filterReal(allAddresses, 60);
await distributeAirdrop(eligible);
```

### `filterSuspicious(addresses, maxScore?)`

Filter to only suspicious addresses.

```javascript
const suspicious = await certid.filterSuspicious(allAddresses, 30);
console.log(`Blocked ${suspicious.length} potential Sybils`);
```

## Links

- [Documentation](https://c3rt.org/developers/sybil-api)
- [API Reference](https://c3rt.org/developers/api)
- [Get API Key](https://c3rt.org/signup)

## License

Apache-2.0
