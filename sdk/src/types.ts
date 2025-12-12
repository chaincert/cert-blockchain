/**
 * Type definitions for CERT Blockchain SDK
 * Per Whitepaper Section 3 and 8
 */

export interface ClientConfig {
  rpcUrl: string;
  apiUrl: string;
  ipfsUrl?: string;
  chainId?: string;
}

export interface AttestationData {
  uid: string;
  schemaUID: string;
  attester: string;
  recipient: string;
  data: Record<string, unknown>;
  time: number;
  expirationTime?: number;
  revocationTime?: number;
  revocable: boolean;
  refUID?: string;
}

export interface EncryptedAttestationData {
  uid: string;
  schemaUID: string;
  attester: string;
  ipfsCID: string;
  encryptedDataHash: string;
  recipients: Recipient[];
  revocable: boolean;
  revoked: boolean;
  expirationTime?: number;
  createdAt: number;
}

export interface Recipient {
  address: string;
  encryptedKey: string;
}

export interface Schema {
  uid: string;
  creator: string;
  schema: string;
  resolver?: string;
  revocable: boolean;
  createdAt: number;
}

export interface CertIDProfile {
  address: string;
  name?: string;
  bio?: string;
  avatarUrl?: string;
  publicKey?: string;
  socialLinks?: Record<string, string>;
  credentials?: string[];
  verified: boolean;
  verificationLevel: number;
  createdAt: number;
  updatedAt: number;
}

export interface EncryptionKeys {
  publicKey: string;
  privateKey: string;
}

export interface IPFSConfig {
  url: string;
  gateway?: string;
}

export interface CreateAttestationRequest {
  schemaUID: string;
  recipient: string;
  data: Record<string, unknown>;
  revocable?: boolean;
  expirationTime?: number;
  refUID?: string;
}

export interface CreateEncryptedAttestationRequest {
  schemaUID: string;
  data: Record<string, unknown>;
  recipients: string[];
  revocable?: boolean;
  expirationTime?: number;
}

export interface RegisterSchemaRequest {
  schema: string;
  resolver?: string;
  revocable?: boolean;
}

export interface UpdateProfileRequest {
  name?: string;
  bio?: string;
  avatarUrl?: string;
  publicKey?: string;
  socialLinks?: Record<string, string>;
}

export interface VerifySocialRequest {
  platform: string;
  handle: string;
  proof: string;
}

export interface TransactionResult {
  txHash: string;
  success: boolean;
  gasUsed: number;
  events?: Record<string, unknown>[];
}

export interface QueryResult<T> {
  data: T;
  pagination?: {
    nextKey?: string;
    total: number;
  };
}

// Attestation types per Whitepaper Section 3.4
export type AttestationType = 
  | 'BUSINESS_DOCUMENT'
  | 'IDENTITY_VERIFICATION'
  | 'CREDENTIAL'
  | 'CERTIFICATE';

// Default schema UIDs per Whitepaper Section 3.4
export interface DefaultSchemas {
  BUSINESS_DOCUMENT: string;
  IDENTITY_VERIFICATION: string;
  CREDENTIAL: string;
  CERTIFICATE: string;
}

