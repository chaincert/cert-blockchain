# CLI Patching Scripts Analysis - Genesis Structure Compatibility

## Executive Summary

This document analyzes the existing CLI patching scripts in the CERT blockchain project and their compatibility with the new custom genesis structure. The analysis reveals significant conflicts between the existing shell scripts that manually patch `genesis.json` and the new programmatic genesis initialization system.

## Current CLI Patching Scripts

### 1. docker-entrypoint.sh
**Location**: `cert-blockchain/scripts/docker-entrypoint.sh`
**Purpose**: Docker container initialization script that sets up the blockchain node

**Current Patches Applied**:
- Line 83-84: Changes bond denom from "stake" to "ucert"
- Line 88-91: Disables inflation (sets to 0.000000000000000000)
- Lines 152-182: Adds genesis accounts with Tokenomics v2.1 distribution
- Lines 190-197: Creates gentx for validator
- Lines 201-205: Collects and validates genesis

**Issues Identified**:
- Uses `sed` to manually patch JSON (fragile)
- Duplicates functionality now handled by `NewDefaultGenesisState()`
- Token distribution logic conflicts with new custom genesis approach
- EVM address format issues (lines 159-182 show workarounds for unsupported formats)

### 2. init.sh
**Location**: `cert-blockchain/scripts/init.sh`
**Purpose**: Manual node initialization script

**Current Patches Applied**:
- Lines 102: Adds genesis account with 1 billion CERT
- Lines 106-113: Creates gentx for validator
- Lines 116-119: Collects and validates genesis

**Issues Identified**:
- Simple approach but conflicts with new genesis structure
- Uses `certd genesis add-genesis-account` which may not work with custom genesis

### 3. deploy-production.sh
**Location**: `cert-blockchain/scripts/deploy-production.sh`
**Purpose**: Production deployment script

**Current Patches Applied**: None directly to genesis.json
**Notes**: This script only orchestrates Docker containers and doesn't modify genesis files

## New Genesis Structure

### Custom Genesis System
**Location**: `cert-blockchain/app/genesis.go`

**Key Components**:
1. **NewDefaultGenesisState()**: Programmatic genesis state creation
2. **Module-specific genesis helpers**:
   - `GetStakingGenesisState()` - Sets bond denom to "ucert", disables inflation
   - `GetBankGenesisState()` - Adds CERT token metadata
   - `GetMintGenesisState()` - Sets zero inflation per Whitepaper Section 5.3
   - `GetGovGenesisState()` - Governance parameters
   - `GetSlashingGenesisState()` - Slashing parameters
   - `GetCrisisGenesisState()` - Crisis constant fee

3. **Module Basic Overrides**:
   - `StakingModuleBasicGenesis` - Custom staking genesis
   - `BankModuleBasicGenesis` - Custom bank genesis with token metadata
   - `SlashingModuleBasicGenesis` - Custom slashing genesis
   - `GovModuleBasicGenesis` - Custom governance genesis
   - `MintModuleBasicGenesis` - Custom mint genesis
   - `CrisisModuleBasicGenesis` - Custom crisis genesis

### Module Genesis Files
- **attestation/genesis.go**: Pre-deployed EAS schemas per Whitepaper Section 3.4
- **certid/genesis.go**: Default CertID module parameters

## Conflicts and Redundancies

### 1. Bond Denom Configuration
**Conflict**: 
- `docker-entrypoint.sh` line 83: `sed -i 's/"bond_denom": "stake"/"bond_denom": "ucert"/'`
- `NewDefaultGenesisState()` already sets this via `GetStakingGenesisState()`

**Resolution**: Remove sed patch from docker-entrypoint.sh

### 2. Inflation Configuration
**Conflict**:
- `docker-entrypoint.sh` lines 88-91: Multiple sed commands to disable inflation
- `NewDefaultGenesisState()` already sets zero inflation via `GetMintGenesisState()`

**Resolution**: Remove all inflation-related sed patches

### 3. Token Metadata
**Conflict**:
- No current script adds token metadata
- `NewDefaultGenesisState()` adds comprehensive CERT token metadata via `GetBankGenesisState()`

**Resolution**: No action needed - new system handles this properly

### 4. Genesis Accounts and Token Distribution
**Major Conflict**:
- `docker-entrypoint.sh` lines 152-182: Manual token distribution using `certd genesis add-genesis-account`
- `NewDefaultGenesisState()` creates genesis state programmatically without individual accounts

**Resolution**: This is the most complex issue requiring careful analysis

## Recommended Actions

### 1. Remove Redundant Patches
Remove the following from `docker-entrypoint.sh`:
- Lines 83-84: Bond denom sed patch
- Lines 88-91: Inflation sed patches
- Lines 152-182: Manual genesis account additions

### 2. Preserve Necessary Functionality
Keep in `docker-entrypoint.sh`:
- Validator key creation (lines 136-141)
- Validator gentx creation (lines 189-197)
- Genesis collection and validation (lines 200-205)
- Configuration file modifications (app.toml, config.toml)

### 3. Update init.sh
Modify `init.sh` to:
- Use `certd init` without manual genesis account additions
- Rely on `NewDefaultGenesisState()` for initial state
- Keep validator setup and gentx creation

### 4. Token Distribution Strategy
**Decision Required**: How to handle the 1 billion CERT token distribution

**Option A**: Keep manual distribution in scripts
- Pros: Maintains current tokenomics approach
- Cons: Conflicts with new genesis system, fragile

**Option B**: Implement distribution in custom genesis
- Pros: Clean integration with new system
- Cons: Requires modifying `NewDefaultGenesisState()` to include accounts

**Recommendation**: Option B - Modify custom genesis to include pre-allocated accounts

## Implementation Plan

### Phase 1: Remove Conflicting Patches
1. Update `docker-entrypoint.sh` to remove redundant sed patches
2. Update `init.sh` to remove manual genesis account additions
3. Test basic node initialization

### Phase 2: Implement Token Distribution in Custom Genesis
1. Modify `NewDefaultGenesisState()` to include genesis accounts
2. Create helper function for token distribution
3. Ensure compatibility with existing module genesis

### Phase 3: Validation and Testing
1. Test Docker container initialization
2. Test manual node initialization
3. Verify token distribution works correctly
4. Validate genesis state integrity

## Files to Modify

1. **cert-blockchain/scripts/docker-entrypoint.sh**
   - Remove lines 83-84 (bond denom patch)
   - Remove lines 88-91 (inflation patches)
   - Remove lines 152-182 (manual account additions)
   - Update comments to reflect new approach

2. **cert-blockchain/scripts/init.sh**
   - Remove line 102 (manual account addition)
   - Update to use new genesis approach
   - Keep validator setup and gentx creation

3. **cert-blockchain/app/genesis.go**
   - Add token distribution to `NewDefaultGenesisState()`
   - Create helper function for genesis accounts
   - Ensure proper integration with module genesis

## Risk Assessment

### High Risk
- **Token Distribution**: Changing how tokens are distributed could affect mainnet launch
- **Genesis State Integrity**: Modifications could break existing testnets

### Medium Risk
- **Script Compatibility**: Changes might break existing deployment workflows
- **Address Format Issues**: EVM vs Bech32 address compatibility

### Low Risk
- **Configuration Changes**: App.toml and config.toml modifications are low risk

## Testing Strategy

1. **Unit Tests**: Test new genesis state creation
2. **Integration Tests**: Test Docker container initialization
3. **End-to-End Tests**: Test complete node setup and token distribution
4. **Regression Tests**: Ensure existing functionality still works

## Conclusion

The existing CLI patching scripts contain significant redundancies and conflicts with the new custom genesis structure. The recommended approach is to:

1. Remove redundant manual patches that are now handled by the custom genesis system
2. Implement token distribution directly in the custom genesis state
3. Preserve necessary validator setup and configuration functionality
4. Thoroughly test the modified approach

This will result in a cleaner, more maintainable system that properly integrates with the new genesis architecture while preserving the intended tokenomics and functionality.