# Changelog

## [1.0.0] - December 2024

### Added

- Complete CERT Blockchain implementation
- Attestation module with EAS compatibility
- CertID decentralized identity module
- REST API with JWT authentication
- TypeScript SDK
- Solidity smart contracts
- Enterprise dApps (Healthcare, Academia, Governance, Legal)
- Bridge validator service
- IPFS integration for encrypted attestations
- Docker Compose deployment
- Production deployment scripts

### Technical Stack

- Cosmos SDK v0.50.5
- CometBFT v0.38.5
- Go 1.21+
- PostgreSQL 15
- IPFS (Kubo)

### Fixed

- InterfaceRegistry configuration for Cosmos SDK v0.50.x
- Staking keeper validator address codec
- Genesis bond denomination (stake → ucert)
- Module service registration

### Tests

- 240 tests passing across all components
- Go unit tests (68)
- TypeScript SDK tests (98)
- Solidity contract tests (28)
- dApp tests (39)
- Bridge validator tests (7)

---

## Roadmap

### v1.1.0 (Planned)

- [ ] Enhanced cross-chain bridge
- [ ] Mobile SDK
- [ ] Additional attestation schemas
- [ ] Batch attestation support

### v1.2.0 (Planned)

- [ ] Governance module enhancements
- [ ] Staking rewards optimization
- [ ] Layer-2 scaling solutions

### v2.0.0 (Future)

- [ ] Zero-knowledge proof attestations
- [ ] Multi-chain deployment
- [ ] Decentralized resolver registry

