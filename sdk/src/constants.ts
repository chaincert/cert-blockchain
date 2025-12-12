/**
 * Constants for CERT Blockchain SDK
 * Per Whitepaper Sections 4, 5, 9, and 12
 */

// Network configuration
export const CERT_CHAIN_ID = 'cert-mainnet-1';
export const CERT_TESTNET_CHAIN_ID = 'cert-testnet-1';
export const CERT_EVM_CHAIN_ID = 8888;

// Default endpoints
export const CERT_RPC_URL = 'https://rpc.certblockchain.io';
export const CERT_API_URL = 'https://api.certblockchain.io';
export const CERT_IPFS_GATEWAY = 'https://ipfs.certblockchain.io';

// Token parameters per Whitepaper Section 5
export const CERT_DENOM = 'ucert';
export const CERT_DECIMALS = 6;
export const CERT_TOTAL_SUPPLY = 1_000_000_000; // 1 Billion CERT

// Network parameters per Whitepaper Section 4 and 12
export const MAX_VALIDATORS = 80;
export const BLOCK_TIME_MS = 2000; // 2 seconds
export const UNBONDING_DAYS = 21;
export const MAX_GAS_PER_BLOCK = 30_000_000;

// Attestation parameters per Whitepaper Section 12
export const MAX_RECIPIENTS_PER_ATTESTATION = 50;
export const MAX_ENCRYPTED_FILE_SIZE = 100 * 1024 * 1024; // 100 MB

// Slashing parameters per Whitepaper Section 4.1
export const DOWNTIME_SLASH_FRACTION = 0.0001; // 0.01%
export const DOUBLE_SIGN_SLASH_FRACTION = 0.05; // 5%

// Default schemas per Whitepaper Section 3.4
export const DEFAULT_SCHEMAS = {
  BUSINESS_DOCUMENT: '0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef',
  IDENTITY_VERIFICATION: '0x2345678901abcdef2345678901abcdef2345678901abcdef2345678901abcdef',
  CREDENTIAL: '0x3456789012abcdef3456789012abcdef3456789012abcdef3456789012abcdef',
  CERTIFICATE: '0x4567890123abcdef4567890123abcdef4567890123abcdef4567890123abcdef',
};

// Schema definitions per Whitepaper Section 3.4
export const SCHEMA_DEFINITIONS = {
  BUSINESS_DOCUMENT: 'string documentType, string documentHash, string issuer, uint64 issuedAt, uint64 expiresAt, string metadata',
  IDENTITY_VERIFICATION: 'string verificationType, string verifiedData, address verifier, uint64 verifiedAt, uint8 confidenceLevel',
  CREDENTIAL: 'string credentialType, string title, string issuer, uint64 issuedAt, uint64 expiresAt, string metadata',
  CERTIFICATE: 'string certificateType, string title, string issuer, address recipient, uint64 issuedAt, string metadata',
};

// Contract addresses (deployed at genesis)
export const CONTRACT_ADDRESSES = {
  SCHEMA_REGISTRY: '0x0000000000000000000000000000000000000001',
  EAS: '0x0000000000000000000000000000000000000002',
  ENCRYPTED_ATTESTATION: '0x0000000000000000000000000000000000000003',
  CERT_TOKEN: '0x0000000000000000000000000000000000000004',
};

// Encryption parameters per Whitepaper Section 3.2
export const ENCRYPTION_ALGORITHM = 'AES-256-GCM';
export const KEY_DERIVATION_ALGORITHM = 'ECIES';
export const HASH_ALGORITHM = 'SHA-256';

// IPFS configuration
export const IPFS_DEFAULT_TIMEOUT = 30000; // 30 seconds
export const IPFS_MAX_FILE_SIZE = MAX_ENCRYPTED_FILE_SIZE;

