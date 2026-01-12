/*
 * Copyright 2026 Brandon Guynn
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */


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
  handle?: string;           // CertID handle (e.g., "alice.cert")
  bio?: string;
  avatarUrl?: string;
  metadataURI?: string;      // IPFS URI for extended metadata
  publicKey?: string;
  socialLinks?: Record<string, string>;
  credentials?: string[];
  badges?: string[];         // Soulbound badges held by this profile
  verified: boolean;
  isVerified?: boolean;      // Alias for verified
  verificationLevel: number;
  trustScore?: number;       // Trust score (0-100)
  entityType?: EntityType;   // Type of entity
  isActive?: boolean;        // Profile active status
  createdAt: number;
  updatedAt: number;
}

// Entity types for CertID profiles
export enum EntityType {
  Individual = 0,
  Institution = 1,
  SystemAdmin = 2,
  Bot = 3,
}

// Badge types for Soulbound Tokens
export type BadgeType =
  | 'KYC_L1'
  | 'KYC_L2'
  | 'ACADEMIC_ISSUER'
  | 'VERIFIED_CREATOR'
  | 'GOV_AGENCY'
  | 'LEGAL_ENTITY'
  | 'ISO_9001_CERTIFIED';

// Full identity with resolved badges
export interface FullIdentity {
  address: string;
  handle: string;
  metadata: string;
  isVerified: boolean;
  isInstitutional: boolean;
  trustScore: number;
  entityType: EntityType;
  badges: string[];
  isKYC: boolean;
  isAcademic: boolean;
  isCreator: boolean;
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

