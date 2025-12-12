/**
 * @certblockchain/sdk
 * Official JavaScript SDK for CERT Blockchain
 * 
 * Per Whitepaper Section 8 - SDK and API
 */

export { CertClient } from './client';
export { EncryptedAttestation } from './attestation';
export { CertID } from './certid';
export { Encryption } from './encryption';
export { IPFS } from './ipfs';

// Types
export type {
  AttestationData,
  EncryptedAttestationData,
  Schema,
  Recipient,
  CertIDProfile,
  EncryptionKeys,
  IPFSConfig,
  ClientConfig,
} from './types';

// Constants
export {
  CERT_CHAIN_ID,
  CERT_RPC_URL,
  CERT_API_URL,
  CERT_IPFS_GATEWAY,
  DEFAULT_SCHEMAS,
} from './constants';

// Utilities
export {
  generateUID,
  hashData,
  validateAddress,
  formatCERT,
  parseCERT,
} from './utils';

