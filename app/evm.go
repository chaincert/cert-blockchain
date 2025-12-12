package app

import (
	"math/big"
)

// EVMConfig contains EVM configuration for CERT Blockchain
// Per Whitepaper Section 2.2 - Full EVM Compatibility
// Note: EVM functionality is provided via JSON-RPC proxy to external EVM node
// or via CosmWasm contracts. Direct ethermint integration deferred due to
// Cosmos SDK v0.50 compatibility issues.
type EVMConfig struct {
	// ChainID for EVM (different from Cosmos chain ID)
	ChainID *big.Int

	// ExtraEIPs defines additional EIPs to enable
	ExtraEIPs []int64

	// AllowUnprotectedTxs allows unprotected (non EIP-155) transactions
	AllowUnprotectedTxs bool

	// JSONRPCEnabled enables the JSON-RPC server
	JSONRPCEnabled bool

	// JSONRPCAddress is the address for JSON-RPC server
	JSONRPCAddress string
}

// DefaultEVMConfig returns the default EVM configuration for CERT
func DefaultEVMConfig() EVMConfig {
	return EVMConfig{
		ChainID:             big.NewInt(8888), // CERT Chain ID per Whitepaper Section 9
		ExtraEIPs:           []int64{},
		AllowUnprotectedTxs: false,
		JSONRPCEnabled:      true,
		JSONRPCAddress:      "0.0.0.0:8545",
	}
}

// EVMParams represents EVM module parameters
// This is a placeholder for future ethermint integration
type EVMParams struct {
	EvmDenom            string
	EnableCreate        bool
	EnableCall          bool
	ExtraEIPs           []int64
	AllowUnprotectedTxs bool
}

// GetEVMParams returns EVM module parameters per Whitepaper Section 2.2
func GetEVMParams() EVMParams {
	return EVMParams{
		EvmDenom:            TokenDenom, // CERT token for gas fees
		EnableCreate:        true,       // Allow contract deployment
		EnableCall:          true,       // Allow contract calls
		ExtraEIPs:           []int64{},  // No extra EIPs initially
		AllowUnprotectedTxs: false,      // Require EIP-155 signed txs
	}
}

// FeeMarketParams represents fee market (EIP-1559) parameters
type FeeMarketParams struct {
	BaseFee                  int64
	MinGasPrice              int64
	MinGasMultiplier         float64
	EnableHeight             int64
	BaseFeeChangeDenominator uint32
	ElasticityMultiplier     uint32
	NoBaseFee                bool
}

// GetFeeMarketParams returns fee market (EIP-1559) parameters
func GetFeeMarketParams() FeeMarketParams {
	return FeeMarketParams{
		BaseFee:                  1000000000, // 1 Gwei base fee
		MinGasPrice:              0,          // No minimum gas price
		MinGasMultiplier:         0.5,        // 0.5 multiplier
		EnableHeight:             0,          // Enable from genesis
		BaseFeeChangeDenominator: 8,          // EIP-1559 standard
		ElasticityMultiplier:     2,          // EIP-1559 standard
		NoBaseFee:                false,      // Use EIP-1559 base fee
	}
}

// Precompiled contract addresses for CERT-specific functionality
var (
	// AttestationPrecompileAddress is the address for the attestation precompile
	// This allows EVM contracts to interact with the attestation module
	AttestationPrecompileAddress = "0x0000000000000000000000000000000000001000"

	// CertIDPrecompileAddress is the address for CertID precompile
	CertIDPrecompileAddress = "0x0000000000000000000000000000000000001001"
)

// GetPrecompiledContractAddresses returns the list of precompiled contracts
func GetPrecompiledContractAddresses() []string {
	return []string{
		AttestationPrecompileAddress,
		CertIDPrecompileAddress,
	}
}
