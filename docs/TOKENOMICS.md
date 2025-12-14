# CERT Blockchain Tokenomics v2.1

## Official Token Economics Specification

**Version:** 2.1 | **Updated:** December 2025 | **Domain:** C3rt.org

---

## 1. Token Overview

| Parameter | Value |
|-----------|-------|
| **Token Name** | CERT |
| **Blockchain** | Cert Layer-1 (Native Asset) |
| **Total Supply** | 1,000,000,000 CERT (Fixed Max Supply) |
| **Decimals** | 6 (native: ucert) / 18 (ERC-20 bridge) |
| **Base Denomination** | ucert (1,000,000 ucert = 1 CERT) |
| **Chain ID** | 951753 |

---

## 2. Core Utility

| Utility | Description |
|---------|-------------|
| **Gas Fees** | All transactions require CERT for gas |
| **Attestation Fees** | Creating/querying encrypted attestations |
| **Staking** | Validators must stake CERT to participate in consensus |
| **Governance** | Token holders vote on protocol upgrades and parameters |
| **Protocol Access** | Enterprise clients hold/burn CERT for registry access |

---

## 3. Token Allocation

| Category | % | CERT Tokens | ucert (on-chain) |
|----------|---|-------------|------------------|
| Treasury & Ecosystem | 32% | 320,000,000 | 320,000,000,000,000 |
| Community & Staking Rewards | 30% | 300,000,000 | 300,000,000,000,000 |
| Private Sale & Liquidity | 15% | 150,000,000 | 150,000,000,000,000 |
| Core Team & Founders | 15% | 150,000,000 | 150,000,000,000,000 |
| Advisors & Future Hires | 5% | 50,000,000 | 50,000,000,000,000 |
| Community Airdrop | 3% | 30,000,000 | 30,000,000,000,000 |
| **Total** | **100%** | **1,000,000,000** | **1,000,000,000,000,000** |

### Allocation Purpose

- **Treasury (32%)**: Wyoming Protocol LLC - Development grants, partnerships, marketing, future DAO
- **Staking Rewards (30%)**: Protocol security - Validator and delegator rewards
- **Private Sale (15%)**: Cayman Token Sale LLC - Seed/Private sale and exchange liquidity
- **Team & Founders (15%)**: Founder/Engineer compensation and deferred compensation
- **Advisors (5%)**: Board members, strategic hires, talent acquisition
- **Airdrop (3%)**: Community distribution for adoption and decentralization

---

## 4. Vesting Schedule

| Stakeholder | Cliff | Vesting | Release | TGE Unlock |
|-------------|-------|---------|---------|------------|
| Founders (15%) | 12 months | 48 months | Monthly linear | 0% |
| Core Team | 6 months | 36 months | Monthly linear | 0% |
| Private Sale (15%) | 3 months | 18 months | Monthly linear | 10% |
| Advisors (5%) | 6 months | 12 months | Monthly linear | 0% |
| Airdrop (3%) | None | 6 months | Monthly linear | 25% |
| Staking Rewards (30%) | None | 10 years | Halving schedule | Active |
| Treasury (32%) | None | Governance | DAO proposals | 5% |

---

## 5. Staking Rewards Emission (10-Year Halving Schedule)

Total Pool: **300,000,000 CERT**

| Year | Annual Emission | Cumulative | Remaining Pool | Est. APY* |
|------|-----------------|------------|----------------|-----------|
| 1 | 75,000,000 | 75,000,000 | 225,000,000 | ~25% |
| 2 | 75,000,000 | 150,000,000 | 150,000,000 | ~20% |
| 3 | 37,500,000 | 187,500,000 | 112,500,000 | ~12% |
| 4 | 37,500,000 | 225,000,000 | 75,000,000 | ~10% |
| 5 | 18,750,000 | 243,750,000 | 56,250,000 | ~6% |
| 6 | 18,750,000 | 262,500,000 | 37,500,000 | ~5% |
| 7 | 9,375,000 | 271,875,000 | 28,125,000 | ~3% |
| 8 | 9,375,000 | 281,250,000 | 18,750,000 | ~2.5% |
| 9 | 9,375,000 | 290,625,000 | 9,375,000 | ~2% |
| 10 | 9,375,000 | 300,000,000 | 0 | ~1.5% |

*APY estimates based on projected staking participation (30-50% of supply)

**After Year 10:** Validator rewards come exclusively from transaction fees.

---

## 6. Fee Economics

### Fee Distribution (EIP-1559 Style)

| Destination | Percentage | Purpose |
|-------------|------------|---------|
| **Burned** | 50% | Deflationary pressure, increases scarcity |
| **Validators** | 50% | Block producer rewards |

### Fee Burn Mechanism

```
Total Fee = Base Fee + Priority Fee
Burned Amount = Total Fee × 50%
Validator Reward = Total Fee × 50%
```

**Governance Adjustable:** Fee burn rate can be modified via on-chain governance (range: 25-75%).

---

## 7. Validator Requirements

| Parameter | Value |
|-----------|-------|
| **Minimum Self-Stake** | 10,000 CERT |
| **Maximum Validators** | 80 active |
| **Unbonding Period** | 21 days |
| **Downtime Slashing** | 0.01% of stake |
| **Double-Sign Slashing** | 5% of stake |
| **Commission Range** | 0% - 100% (validator choice) |

---

## 8. Circulating Supply Projection

### At TGE (Token Generation Event)

| Source | Unlocked | CERT |
|--------|----------|------|
| Private Sale (10% TGE) | 10% of 150M | 15,000,000 |
| Airdrop (25% TGE) | 25% of 30M | 7,500,000 |
| Treasury (5% operational) | 5% of 320M | 16,000,000 |
| **Initial Circulating** | | **38,500,000** |

**Initial Circulating Supply: ~3.85% of total supply**

### Circulating Supply Over Time

| Month | Est. Circulating | % of Total |
|-------|------------------|------------|
| TGE | 38,500,000 | 3.85% |
| Month 3 | 65,000,000 | 6.5% |
| Month 6 | 120,000,000 | 12% |
| Month 12 | 200,000,000 | 20% |
| Month 24 | 350,000,000 | 35% |
| Month 48 | 550,000,000 | 55% |

---

## 9. Deflationary Mechanics

### Supply Reduction Drivers

1. **Fee Burns (50%)** - Every transaction burns CERT
2. **Slashing** - Validator misbehavior destroys tokens
3. **Protocol Burns** - Enterprise attestation fees partially burned

### Projected Burn Rate

| Network Activity | Daily Txns | Daily Burn | Annual Burn |
|------------------|------------|------------|-------------|
| Low | 10,000 | ~500 CERT | ~182,500 CERT |
| Medium | 100,000 | ~5,000 CERT | ~1,825,000 CERT |
| High | 1,000,000 | ~50,000 CERT | ~18,250,000 CERT |

*Assumes average fee of 0.1 CERT per transaction*

---

## 10. Governance Parameters

### Voting Power

- 1 CERT = 1 Vote
- Staked CERT can vote (no need to unstake)
- Delegation supported (delegate votes to validators)

### Proposal Thresholds

| Parameter | Value |
|-----------|-------|
| Minimum Deposit | 10,000 CERT |
| Voting Period | 14 days |
| Quorum | 33.4% of staked supply |
| Pass Threshold | 50% Yes votes |
| Veto Threshold | 33.4% No with Veto |

### Governable Parameters

- Fee burn rate (25-75%)
- Minimum validator stake
- Slashing percentages
- Treasury fund allocation
- Protocol upgrades

---

## 11. Legal Structure

| Entity | Jurisdiction | Responsibility | Token Allocation |
|--------|--------------|----------------|------------------|
| **Protocol LLC** | Wyoming, USA | IP, Codebase, Development | Treasury (32%) |
| **Token Sale LLC** | Cayman Islands | Fundraising, Token Issuance | Private Sale (15%) |

### DAO Transition

The Wyoming Protocol LLC will transition treasury control to a DAO governed by CERT holders, leveraging Wyoming's DAO LLC statute for legal recognition.

---

## 12. Summary

| Metric | Value |
|--------|-------|
| Total Supply | 1,000,000,000 CERT |
| Initial Circulating | ~38,500,000 CERT (3.85%) |
| Staking Rewards Duration | 10 years (halving) |
| Fee Burn Rate | 50% |
| Min Validator Stake | 10,000 CERT |
| Fully Vested | ~48 months post-TGE |

---

## Appendix: Token Addresses (Mainnet)

| Wallet | Purpose | EVM Address | Allocation |
|--------|---------|-------------|------------|
| Treasury | Ecosystem development | `0xc68a92163f496ADCc7A8502fB2fdc7341fFdF589` | 320,000,000 CERT (32%) |
| Staking Pool | Validator rewards | `0x5813612e4736cE42FC8582e0dBC7Ef51cAe906b9` | 300,000,000 CERT (30%) |
| Team + Private Sale | Founder/investor | `0x0D756101183fe368C3364aBDe6Bf063CC3e7fcFD` | 300,000,000 CERT (30%) |
| Advisors | Advisor lockup | `0x0547711B2aC90Cead95E010B863a304602b40bF6` | 50,000,000 CERT (5%) |
| Airdrop | Community distribution | `0xc68a92163f496ADCc7A8502fB2fdc7341fFdF589`* | 30,000,000 CERT (3%) |

*Airdrop funds held in Treasury wallet until distribution event.

---

**Document Version:** 2.1
**Last Updated:** December 2024
**Approved By:** CERT Protocol LLC

