# @certblockchain/sdk

Official JavaScript/TypeScript SDK for the CERT Blockchain ecosystem. Build privacy-preserving applications with encrypted attestations, decentralized identity (CertID), and verifiable credentials.

[![npm version](https://img.shields.io/npm/v/@certblockchain/sdk)](https://www.npmjs.com/package/@certblockchain/sdk)
[![License: Apache 2.0](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Tests](https://img.shields.io/badge/tests-98%20passing-brightgreen)](https://github.com/chaincertify/cert-blockchain)

## Features
> [!IMPORTANT]
> **Chain ID Notice**: The correct Chain ID is **`cert-testnet-1`**. Do not use `951753` unless you are specifically interacting with the EVM JSON-RPC directly.

- 🔐 **Encrypted Attestations** - AES-256-GCM encryption with ECIES key wrapping

- 🆔 **CertID Integration** - Decentralized identity management
- 📦 **IPFS Storage** - Decentralized storage for encrypted data
- ⛓️ **Dual Chain Support** - Works with both Cosmos and EVM chains
- 🔑 **Wallet Integration** - Compatible with Keplr, MetaMask, and other wallets
- 📝 **TypeScript** - Full type safety and IntelliSense support
- ✅ **Production Ready** - 98 passing tests, thoroughly tested

## Installation

```bash
npm install @certblockchain/sdk ethers
```

## Quick Start

### Initialize the Client

```typescript
import { CertClient } from '@certblockchain/sdk';

const client = new CertClient({
  rpcUrl: 'https://rpc.c3rt.org',
  apiUrl: 'https://api.c3rt.org/api/v1',
  ipfsUrl: 'https://ipfs.c3rt.org',
});
```

### Create an Encrypted Attestation

```typescript
import { ethers } from 'ethers';

// Connect your wallet
const provider = new ethers.BrowserProvider(window.ethereum);
const signer = await provider.getSigner();

// Create attestation data
const attestationData = {
  patientId: 'P12345',
  recordType: 'LAB_RESULT',
  data: {
    testName: 'Complete Blood Count',
    results: { wbc: 7.5, rbc: 4.8 }
  },
  timestamp: Date.now()
};

// Define recipients (up to 50)
const recipients = [
  {
    address: '0x742d35Cc6634C0532925a3b844Bc454e4438f44e',
    publicKey: '0x04...' // Recipient's public key
  }
];

// Create encrypted attestation
const result = await client.attestation.create({
  schemaUID: '0x1234...', // Your schema UID
  data: attestationData,
  recipients: recipients,
  signer: signer
});

console.log('Attestation UID:', result.uid);
console.log('IPFS CID:', result.ipfsCID);
```

### Retrieve and Decrypt an Attestation

```typescript
// Retrieve encrypted attestation
const attestation = await client.attestation.get('0xabc123...'); // UID

// Decrypt with your private key
const decrypted = await client.attestation.decrypt(
  attestation,
  yourPrivateKey
);

console.log('Decrypted data:', decrypted.data);
```

### CertID - Decentralized Identity

```typescript
// Register a CertID profile
await client.certid.register({
  username: 'alice',
  displayName: 'Alice Smith',
  bio: 'Healthcare professional',
  avatar: 'ipfs://Qm...',
  signer: signer
});

// Resolve a CertID
const profile = await client.certid.resolve('alice');
console.log(profile.displayName); // "Alice Smith"

// Get profile by address
const profileByAddress = await client.certid.getByAddress(
  '0x742d35Cc6634C0532925a3b844Bc454e4438f44e'
);
```

## Core Concepts

### Encrypted Attestations

CERT uses a 5-step encryption process per the whitepaper:

1. **AES Key Generation** - Generate a 256-bit symmetric key
2. **Data Encryption** - Encrypt attestation data with AES-256-GCM
3. **ECIES Key Wrapping** - Wrap the AES key for each recipient
4. **IPFS Upload** - Store encrypted data on IPFS
5. **On-Chain Anchoring** - Record the IPFS CID and metadata on-chain

Only authorized recipients can decrypt the data using their private keys.

### Schemas

Define the structure of your attestations:

```typescript
import { DEFAULT_SCHEMAS } from '@certblockchain/sdk';

// Use built-in schemas
const schemaUID = DEFAULT_SCHEMAS.IDENTITY_VERIFICATION;

// Or register a custom schema
const customSchema = await client.attestation.registerSchema({
  schema: 'string name, uint256 age, string email',
  resolver: '0x0000000000000000000000000000000000000000',
  revocable: true,
  signer: signer
});
```

### Utility Functions

```typescript
import {
  formatCERT,
  parseCERT,
  validateAddress,
  evmToCosmosAddress,
  hashData,
  generateUID
} from '@certblockchain/sdk';

// Format token amounts
formatCERT(1500000n); // "1.5 CERT"
parseCERT(1.5); // 1500000n

// Validate addresses
validateAddress('0x742d35Cc6634C0532925a3b844Bc454e4438f44e'); // true
validateAddress('cert1abc...'); // true

// Convert EVM to Cosmos address
evmToCosmosAddress('0x742d35Cc6634C0532925a3b844Bc454e4438f44e');
// Returns: "cert1..."

// Hash data
hashData('sensitive data'); // "0xabc123..."
```

## API Reference

### CertClient

Main client for interacting with CERT Blockchain.

**Constructor Options:**
- `rpcUrl` - RPC endpoint (default: https://rpc.c3rt.org)
- `apiUrl` - API endpoint (default: https://api.c3rt.org/api/v1)
- `ipfsUrl` - IPFS endpoint (optional)
- `chainId` - Chain ID (default: 951753)

**Methods:**
- `getProvider()` - Get ethers JSON-RPC provider
- `connectSigner(signer)` - Connect a wallet signer
- `createSigner(privateKey)` - Create signer from private key

### EncryptedAttestation

Manage encrypted attestations.

**Methods:**
- `create(request)` - Create new encrypted attestation
- `get(uid)` - Retrieve attestation by UID
- `decrypt(attestation, privateKey)` - Decrypt attestation data
- `verify(uid)` - Verify attestation integrity
- `revoke(uid, signer)` - Revoke an attestation

### CertID

Decentralized identity management.

**Methods:**
- `register(profile, signer)` - Register new CertID
- `resolve(username)` - Get profile by username
- `getByAddress(address)` - Get profile by wallet address
- `update(profile, signer)` - Update profile
- `addCredential(credential, signer)` - Add verifiable credential

### Encryption

Low-level encryption utilities.

**Methods:**
- `generateKeyPair()` - Generate ECIES key pair
- `generateSymmetricKey()` - Generate AES-256 key
- `encryptData(data, key)` - Encrypt with AES-256-GCM
- `decryptData(encrypted, key)` - Decrypt data
- `wrapKeyForRecipient(key, publicKey)` - Wrap key with ECIES
- `unwrapKey(wrapped, privateKey)` - Unwrap key

## Constants

```typescript
import {
  CERT_CHAIN_ID,        // 'cert-mainnet-1'
  CERT_EVM_CHAIN_ID,    // 951753
  CERT_RPC_URL,         // 'https://rpc.c3rt.org'
  CERT_API_URL,         // 'https://api.c3rt.org/api/v1'
  CERT_DENOM,           // 'ucert'
  CERT_DECIMALS,        // 6
  MAX_VALIDATORS,       // 80
  BLOCK_TIME_MS,        // 2000 (2 seconds)
} from '@certblockchain/sdk';
```

## Examples

See the [examples directory](./examples) for complete working examples:

- [Healthcare Records](./examples/healthcare.ts)
- [Academic Credentials](./examples/credentials.ts)
- [Business Documents](./examples/documents.ts)
- [Identity Verification](./examples/identity.ts)

## Testing

The SDK includes comprehensive tests:

```bash
npm test              # Run all tests
npm run test:watch    # Watch mode
```

**Test Coverage:**
- ✅ 98 tests passing
- ✅ E2E encryption flow
- ✅ Attestation creation and retrieval
- ✅ CertID operations
- ✅ Utility functions
- ✅ IPFS integration

## TypeScript Support

Full TypeScript support with exported types:

```typescript
import type {
  AttestationData,
  EncryptedAttestationData,
  CertIDProfile,
  ClientConfig,
  Recipient,
  Schema,
} from '@certblockchain/sdk';
```

## Browser Support

Works in modern browsers with Web3 wallet extensions:

```html
<script type="module">
  import { CertClient } from 'https://cdn.jsdelivr.net/npm/@certblockchain/sdk/+esm';
  
  const client = new CertClient();
  // Use the SDK
</script>
```

## Requirements

- Node.js >= 18.0.0
- ethers ^6.0.0

## License

MIT © CERT Blockchain Team

## Links

- [Website](https://c3rt.org)
- [Documentation](https://c3rt.org/docs)
- [GitHub](https://github.com/chaincertify/cert-blockchain)
- [Whitepaper](https://c3rt.org/whitepaper)
- [Discord](https://discord.gg/cert)

## Support

- 📧 Email: dev@c3rt.org
- 💬 Discord: https://discord.gg/cert
- 🐛 Issues: https://github.com/chaincertify/cert-blockchain/issues

