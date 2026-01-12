# CLI Patching Scripts - Genesis Structure Migration

## Overview

This document details the migration of CLI patching scripts to work with the new custom genesis structure in the CERT blockchain. The migration removes redundant manual patches and integrates token distribution directly into the programmatic genesis system.

## Changes Made

### 1. docker-entrypoint.sh Modifications

**File**: `cert-blockchain/scripts/docker-entrypoint.sh`

**Removed Redundant Patches**:
- **Lines 83-84**: Removed `sed` patches for bond denom (`"stake"` → `"ucert"`)
- **Lines 88-91**: Removed `sed` patches for inflation parameters (set to 0)
- **Lines 152-182**: Removed manual genesis account additions using `certd genesis add-genesis-account`

**Rationale**:
- Bond denom and inflation are now set by `GetStakingGenesisState()` and `GetMintGenesisState()`
- Manual account additions conflicted with the new custom genesis approach
- EVM address format issues were causing failures in the shell scripts

**Preserved Functionality**:
- Validator key creation and gentx generation
- Configuration file modifications (app.toml, config.toml)
- Genesis collection and validation

### 2. init.sh Modifications

**File**: `cert-blockchain/scripts/init.sh`

**Removed**:
- **Line 102**: Removed manual genesis account addition with 1 billion CERT

**Rationale**:
- Token distribution is now handled by the custom genesis system
- Eliminates duplication with `NewDefaultGenesisState()`

**Preserved Functionality**:
- Validator setup and gentx creation
- Configuration file setup
- Genesis collection and validation

### 3. app/genesis.go Enhancements

**File**: `cert-blockchain/app/genesis.go`

**Added**:
- **New function**: `GetBankGenesisStateWithAccounts()` - Implements Tokenomics v2.1 distribution
- **Modified**: `NewDefaultGenesisState()` - Now includes genesis accounts via bank module

**Token Distribution (Tokenomics v2.1)**:
```
Total Supply: 1,000,000,000 CERT = 1,000,000,000,000,000 ucert

Treasury (32%):     320,000,000 CERT
Staking Pool (30%): 300,000,000 CERT  
Team+Private (30%): 300,000,000 CERT
Advisors (5%):       50,000,000 CERT
Airdrop (3%):        30,000,000 CERT
```

**Genesis Accounts**:
- Validator account: 110,000 CERT (for staking + operations)
- Treasury account: 320,000,000 CERT
- Staking pool account: 300,000,000 CERT
- Team/Private account: 300,000,000 CERT
- Advisors account: 50,000,000 CERT
- Airdrop account: 30,000,000 CERT

## Benefits of the Migration

### 1. **Improved Reliability**
- Eliminates fragile `sed` patches that could break with JSON format changes
- Removes EVM address format compatibility issues
- Uses proper SDK genesis initialization patterns

### 2. **Better Maintainability**
- Single source of truth for genesis configuration
- Programmatic approach allows for validation and testing
- Easier to modify token distribution in the future

### 3. **Enhanced Security**
- No more manual JSON manipulation that could introduce syntax errors
- Proper address validation through SDK types
- Consistent with Cosmos SDK best practices

### 4. **Cleaner Architecture**
- Separation of concerns: genesis logic in app code, deployment logic in scripts
- Scripts focus on deployment orchestration rather than genesis manipulation
- Better integration with existing module system

## Testing Strategy

### 1. **Unit Tests**
```bash
# Test genesis state creation
go test ./app -run TestGenesisState

# Test token distribution
go test ./app -run TestTokenDistribution
```

### 2. **Integration Tests**
```bash
# Test Docker container initialization
docker-compose up -d certd
docker-compose logs certd

# Test manual node initialization
./scripts/init.sh
certd start --home ~/.certd
```

### 3. **End-to-End Tests**
```bash
# Test complete deployment
./scripts/deploy-production.sh

# Verify token distribution
curl http://localhost:1317/cosmos/bank/v1beta1/supply
```

## Migration Impact

### **Breaking Changes**
- None - existing deployments will continue to work
- New deployments will use the improved genesis system

### **Backward Compatibility**
- Existing testnets can continue using current genesis
- New testnets will use the enhanced genesis with proper token distribution

### **Deployment Changes**
- Docker containers will initialize faster (no manual patches)
- Manual deployments will be more reliable
- Production deployments will have consistent genesis state

## Future Enhancements

### 1. **Address Management**
- Replace hardcoded Bech32 addresses with configurable environment variables
- Add support for EVM address format conversion

### 2. **Token Distribution Flexibility**
- Make distribution percentages configurable
- Add support for vesting schedules

### 3. **Genesis Validation**
- Add comprehensive validation of genesis state
- Include token distribution verification

### 4. **Documentation**
- Update deployment guides to reflect new approach
- Add troubleshooting guide for genesis-related issues

## Files Modified

1. **cert-blockchain/scripts/docker-entrypoint.sh**
   - Removed redundant sed patches
   - Removed manual genesis account additions
   - Updated comments to reflect new approach

2. **cert-blockchain/scripts/init.sh**
   - Removed manual genesis account addition
   - Updated to use custom genesis state

3. **cert-blockchain/app/genesis.go**
   - Added `GetBankGenesisStateWithAccounts()` function
   - Modified `NewDefaultGenesisState()` to include accounts
   - Implemented Tokenomics v2.1 distribution

4. **cert-blockchain/GENESIS_SCRIPT_ANALYSIS.md** (new)
   - Comprehensive analysis of existing patches
   - Documentation of conflicts and resolutions
   - Implementation plan and rationale

## Conclusion

The migration successfully removes redundant CLI patches while maintaining all functionality. The new approach is more reliable, maintainable, and follows Cosmos SDK best practices. Token distribution is now properly integrated into the genesis system, eliminating the fragile shell script patches that were prone to failure.

The changes are backward compatible and provide a solid foundation for future enhancements to the genesis system.