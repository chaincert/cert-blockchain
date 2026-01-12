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
  FullIdentity,
  BadgeType,
} from './types';

export { EntityType } from './types';

// Constants
export {
  CERT_CHAIN_ID,
  CERT_RPC_URL,
  CERT_API_URL,
  CERT_IPFS_GATEWAY,
  DEFAULT_SCHEMAS,
  CONTRACT_ADDRESSES,
  CERT_ID_ABI,
  BADGE_TYPES,
} from './constants';

// Utilities
export {
  generateUID,
  hashData,
  validateAddress,
  formatCERT,
  parseCERT,
} from './utils';

