# @certblockchain/sdk

Official JavaScript/TypeScript SDK for building on **CERT Blockchain** — a Cosmos-SDK based chain with EVM support, focused on encrypted attestations and decentralized identity.

[![npm version](https://img.shields.io/npm/v/@certblockchain/sdk.svg)](https://www.npmjs.com/package/@certblockchain/sdk)
[![License](https://img.shields.io/badge/license-Apache--2.0-blue.svg)](LICENSE)

## Installation

```bash
npm install @certblockchain/sdk
```

## Is This SDK Right for You?

✅ **Yes, if you want to:**
- Build dApps with encrypted, privacy-preserving attestations
- Integrate decentralized identity (CertID) into your application
- Issue and verify credentials on-chain
- Check trust scores and detect Sybil accounts
- Deploy smart contracts on CERT's EVM

❌ **Consider alternatives if you:**
- Need a lighter package (use `@certblockchain/sybil-sdk` for Sybil checks only)
- Want to interact with other chains

## Quick Start

```typescript
import { CertClient } from '@certblockchain/sdk';

// Initialize the client
const client = new CertClient({
  rpcUrl: 'https://evm.c3rt.org',      // EVM RPC
  apiUrl: 'https://api.c3rt.org',       // REST API
});

// Check connection
const blockNumber = await client.getBlockNumber();
console.log(`Connected to CERT at block ${blockNumber}`);
```

## Core Features

### 1. CertClient — Main Entry Point

The unified client for interacting with CERT Blockchain:

```typescript
import { CertClient } from '@certblockchain/sdk';

const client = new CertClient();

// Access sub-modules
client.attestation  // Encrypted attestations
client.certid       // Decentralized identity

// Provider utilities
const provider = client.getProvider();
const balance = await client.getBalance('0x742d35...');
const isCorrect = await client.isCorrectNetwork();
```

### 2. Encrypted Attestations

Create privacy-preserving credentials that only authorized recipients can decrypt:

```typescript
import { ethers } from 'ethers';

const signer = new ethers.Wallet(privateKey, client.getProvider());

// Create an encrypted attestation
const attestation = await client.attestation.create(
  {
    schemaUID: '0x123...', // Your schema
    data: {
      name: 'John Doe',
      degree: 'Computer Science',
      graduationYear: 2024
    },
    recipients: ['0xRecipient1...', '0xRecipient2...'],
    revocable: true
  },
  recipientPublicKeys, // Map<address, publicKey>
  signer
);

console.log(`Attestation created: ${attestation.uid}`);
```

**Retrieve and decrypt:**

```typescript
const { attestation, data } = await client.attestation.retrieve(
  attestationUID,
  recipientPrivateKey,
  signer
);

console.log(data); // Decrypted credential data
```

### 3. CertID — Decentralized Identity

Manage identity profiles, social verifications, and trust scores:

```typescript
// Get a user's profile
const profile = await client.certid.getProfile('0x742d35...');
console.log(profile.displayName);
console.log(profile.trustScore);

// Get full identity with badges
const identity = await client.certid.getFullIdentity('0x742d35...');
console.log(identity.badges);       // ['KYC_L1', 'VERIFIED_CREATOR']
console.log(identity.entityType);   // 'individual' | 'institution'

// Check if user has a specific badge
const isKYCVerified = await client.certid.hasBadge('0x...', 'KYC_L1');

// Get trust score
const trustScore = await client.certid.getTrustScore('0x...');
```

**Social verification:**

```typescript
// Generate proof message
const proofMessage = client.certid.generateSocialProof(
  '0x742d35...',
  'twitter'
);
// User posts this to Twitter, then:

await client.certid.verifySocial(
  { platform: 'twitter', handle: 'username', proofUrl: 'https://...' },
  signer
);
```

### 4. Schema Registration

Define custom credential schemas:

```typescript
const schema = await client.registerSchema(
  {
    schema: 'string name, string degree, uint256 year',
    name: 'Academic Credential',
    description: 'University degree attestation',
    revocable: true
  },
  signer
);

console.log(`Schema registered: ${schema.uid}`);
```

### 5. Encryption Utilities

Direct access to encryption primitives:

```typescript
import { Encryption } from '@certblockchain/sdk';

// Generate symmetric key
const key = Encryption.generateSymmetricKey();

// Encrypt data
const { ciphertext, iv, tag } = await Encryption.encryptData(data, key);

// Decrypt data
const decrypted = await Encryption.decryptData(ciphertext, iv, tag, key);
```

### 6. IPFS Integration

Store and retrieve encrypted data:

```typescript
const ipfs = client.getIPFS();

// Upload
const cid = await ipfs.upload(encryptedData);

// Retrieve
const data = await ipfs.retrieve(cid);
```

## Network Configuration

```typescript
import { 
  CERT_CHAIN_ID,       // Chain ID (77551)
  CERT_RPC_URL,        // https://rpc.c3rt.org
  CERT_API_URL,        // https://api.c3rt.org
  CERT_IPFS_GATEWAY,   // IPFS gateway URL
  CONTRACT_ADDRESSES,  // Deployed contract addresses
  CERT_ID_ABI          // CertID contract ABI
} from '@certblockchain/sdk';
```

## Utility Functions

```typescript
import { 
  generateUID,      // Generate unique identifier
  hashData,         // SHA-256 hash
  validateAddress,  // Validate blockchain address
  formatCERT,       // Format token amounts
  parseCERT         // Parse token strings
} from '@certblockchain/sdk';
```

## TypeScript Support

Full TypeScript definitions included:

```typescript
import type {
  AttestationData,
  EncryptedAttestationData,
  Schema,
  Recipient,
  CertIDProfile,
  EncryptionKeys,
  IPFSConfig,
  ClientConfig,
  FullIdentity,
  BadgeType,
} from '@certblockchain/sdk';

import { EntityType } from '@certblockchain/sdk';
```

## Common Use Cases

### Airdrop with Sybil Protection

```typescript
const addresses = ['0x...', '0x...', '0x...'];

const eligible = [];
for (const addr of addresses) {
  const score = await client.certid.getTrustScore(addr);
  if (score >= 60) eligible.push(addr);
}

console.log(`${eligible.length} eligible for airdrop`);
```

### Gated Access by Badge

```typescript
const hasKYC = await client.certid.hasBadge(userAddress, 'KYC_L1');
if (!hasKYC) {
  throw new Error('KYC verification required');
}
```

### Issue Verifiable Credential

```typescript
const credential = await client.attestation.create(
  {
    schemaUID: EMPLOYMENT_SCHEMA,
    data: { employer: 'Acme Corp', role: 'Engineer', startDate: '2024-01-15' },
    recipients: [employeeAddress],
    revocable: true
  },
  new Map([[employeeAddress, employeePublicKey]]),
  hrSigner
);
```

## API Reference

### CertClient

| Method | Description |
|--------|-------------|
| `getProvider()` | Get ethers JsonRpcProvider |
| `getIPFS()` | Get IPFS client |
| `connectSigner(signer)` | Connect wallet signer |
| `createSigner(privateKey)` | Create signer from key |
| `registerSchema(request, signer)` | Register attestation schema |
| `getSchema(uid)` | Get schema by UID |
| `getBlockNumber()` | Current block number |
| `getBalance(address)` | Get CERT balance |
| `getChainId()` | Get chain ID |
| `isCorrectNetwork()` | Verify network |
| `waitForTransaction(hash, confirmations)` | Wait for tx confirmation |
| `getHealth()` | API health check |

### CertID

| Method | Description |
|--------|-------------|
| `getProfile(address)` | Get CertID profile |
| `updateProfile(profile, signer)` | Update profile |
| `verifySocial(request, signer)` | Verify social account |
| `addCredential(uid, signer)` | Add credential |
| `getFullIdentity(address)` | Get full identity with badges |
| `getTrustScore(address)` | Get trust score (0-100) |
| `hasBadge(address, badge)` | Check for specific badge |
| `resolveHandle(handle)` | Resolve handle to address |

### EncryptedAttestation

| Method | Description |
|--------|-------------|
| `create(request, publicKeys, signer)` | Create encrypted attestation |
| `retrieve(uid, privateKey, signer)` | Retrieve and decrypt |
| `revoke(uid, signer)` | Revoke attestation |
| `get(uid)` | Get attestation metadata |
| `getByAttester(address)` | List by attester |
| `getByRecipient(address)` | List by recipient |

## Links

- **Documentation**: https://c3rt.org/developers
- **API Reference**: https://c3rt.org/developers/api
- **Block Explorer**: https://c3rt.org/blocks
- **GitHub**: https://github.com/chaincertify/cert-blockchain
- **Discord**: https://discord.gg/certid

## License

Apache-2.0 — see [LICENSE](LICENSE)
