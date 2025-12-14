# CERT Blockchain Architecture

## System Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                      Client Layer                           │
├─────────────┬─────────────┬─────────────┬──────────────────┤
│  Web dApps  │  Mobile     │  TypeScript │  CLI Tools       │
│             │  Apps       │  SDK        │                  │
└─────────────┴─────────────┴─────────────┴──────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                      API Layer                              │
├─────────────┬─────────────┬─────────────┬──────────────────┤
│  REST API   │  gRPC       │  RPC        │  WebSocket       │
│  :3000      │  :9090      │  :26657     │  :26657          │
└─────────────┴─────────────┴─────────────┴──────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                   Blockchain Layer                          │
├─────────────┬─────────────┬─────────────┬──────────────────┤
│ Attestation │  CertID     │  Bank       │  Staking         │
│ Module      │  Module     │  Module     │  Module          │
└─────────────┴─────────────┴─────────────┴──────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                   Consensus Layer                           │
│                   CometBFT v0.38.5                          │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                   Storage Layer                             │
├─────────────────────────────┬──────────────────────────────┤
│  LevelDB (Blockchain State) │  IPFS (Encrypted Data)       │
│  PostgreSQL (API Cache)     │                              │
└─────────────────────────────┴──────────────────────────────┘
```

---

## Module Architecture

### Attestation Module (x/attestation)

- **MsgCreateSchema** - Register attestation schemas
- **MsgCreateAttestation** - Create attestations
- **MsgRevokeAttestation** - Revoke attestations
- **MsgCreateEncryptedAttestation** - Privacy-preserving attestations

### CertID Module (x/certid)

- **MsgRegisterIdentity** - Register decentralized identity
- **MsgUpdateIdentity** - Update identity attributes
- **MsgLinkCredential** - Link credentials to identity

---

## Data Flow

### Attestation Creation Flow

1. Client submits attestation via SDK/API
2. API validates JWT authentication
3. Transaction broadcast to blockchain
4. CometBFT consensus validates
5. Attestation stored on-chain
6. Event emitted for indexing
7. PostgreSQL cache updated

### Encrypted Attestation Flow

1. Client encrypts data locally
2. Encrypted payload uploaded to IPFS
3. IPFS CID stored on-chain
4. Only hash visible publicly
5. Decryption requires private key

