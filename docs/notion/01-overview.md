# CERT Blockchain

## Project Summary

CERT Blockchain is a specialized Layer-1 blockchain protocol for on-chain privacy through native encrypted attestations. Part of the Cert ecosystem alongside Chain Certify, Cert Token, and CertID.

- **Domain:** C3rt.org
- **Production Server:** 172.239.32.74

---

## Tech Stack

| Component | Technology |
|-----------|------------|
| Blockchain Framework | Cosmos SDK v0.50.5 |
| Consensus Engine | CometBFT v0.38.5 |
| Chain ID | 951753 |
| Bond Denomination | ucert |
| Bech32 Prefix | cert (accounts), certvaloper (validators) |
| Block Time | ~2 seconds |

---

## Ecosystem Components

- **CERT Blockchain** - Layer-1 blockchain with native attestation support
- **Chain Certify** - Credential verification platform
- **Cert Token** - Native token (CERT/ucert)
- **CertID** - Decentralized identity module

---

## Key Features

- ✅ Native encrypted attestations (EAS-compatible)
- ✅ On-chain privacy with IPFS storage
- ✅ Cross-chain bridge support
- ✅ Enterprise dApps (Healthcare, Academia, Legal, Governance)
- ✅ TypeScript SDK for developers
- ✅ REST API + gRPC endpoints

---

## Repository Structure

```
cert-blockchain/
├── app/                 # Cosmos app configuration
├── x/attestation/       # Attestation module
├── x/certid/            # CertID identity module
├── api/                 # REST API server
├── sdk/                 # TypeScript SDK
├── contracts/           # Solidity smart contracts
├── dapps/               # Enterprise dApps
├── services/            # Bridge validator, etc.
└── tests/               # Integration tests
```

