package app



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
