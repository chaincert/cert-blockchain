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
 * Tests for CERT Blockchain SDK constants
 */

import {
  CERT_CHAIN_ID,
  CERT_EVM_CHAIN_ID,
  CERT_DENOM,
  CERT_DECIMALS,
  CERT_TOTAL_SUPPLY,
  MAX_GAS_PER_BLOCK,
  BLOCK_TIME_MS,
  MAX_VALIDATORS,
  CONTRACT_ADDRESSES,
  CERT_RPC_URL,
  CERT_API_URL,
  CERT_IPFS_GATEWAY,
  MAX_RECIPIENTS_PER_ATTESTATION,
  ENCRYPTION_ALGORITHM,
} from '../constants';

describe('Chain Constants', () => {
  it('should have correct chain ID', () => {
    expect(CERT_CHAIN_ID).toBe('cert-mainnet-1');
  });

  it('should have correct EVM chain ID', () => {
    expect(CERT_EVM_CHAIN_ID).toBe(951753);
  });

  it('should have correct block time (2000ms)', () => {
    expect(BLOCK_TIME_MS).toBe(2000);
  });

  it('should have correct max validators', () => {
    expect(MAX_VALIDATORS).toBe(80);
  });
});

describe('Token Constants', () => {
  it('should have correct CERT denomination', () => {
    expect(CERT_DENOM).toBe('ucert');
  });

  it('should have correct decimals', () => {
    expect(CERT_DECIMALS).toBe(6);
  });

  it('should have correct total supply (1B CERT)', () => {
    expect(CERT_TOTAL_SUPPLY).toBe(1_000_000_000);
  });
});

describe('Gas Constants', () => {
  it('should have max gas per block', () => {
    expect(MAX_GAS_PER_BLOCK).toBeDefined();
    expect(MAX_GAS_PER_BLOCK).toBe(30_000_000);
  });
});

describe('Contract Addresses', () => {
  it('should have EAS contract address', () => {
    expect(CONTRACT_ADDRESSES.EAS).toBeDefined();
    expect(CONTRACT_ADDRESSES.EAS.startsWith('0x')).toBe(true);
  });

  it('should have SchemaRegistry contract address', () => {
    expect(CONTRACT_ADDRESSES.SCHEMA_REGISTRY).toBeDefined();
    expect(CONTRACT_ADDRESSES.SCHEMA_REGISTRY.startsWith('0x')).toBe(true);
  });

  it('should have EncryptedAttestation contract address', () => {
    expect(CONTRACT_ADDRESSES.ENCRYPTED_ATTESTATION).toBeDefined();
    expect(CONTRACT_ADDRESSES.ENCRYPTED_ATTESTATION.startsWith('0x')).toBe(true);
  });

  it('should have CERT token contract address', () => {
    expect(CONTRACT_ADDRESSES.CERT_TOKEN).toBeDefined();
    expect(CONTRACT_ADDRESSES.CERT_TOKEN.startsWith('0x')).toBe(true);
  });
});

describe('Endpoints', () => {
  it('should have RPC endpoint', () => {
    expect(CERT_RPC_URL).toBeDefined();
    expect(CERT_RPC_URL).toContain('rpc');
  });

  it('should have API endpoint', () => {
    expect(CERT_API_URL).toBeDefined();
    expect(CERT_API_URL).toContain('api');
  });

  it('should have IPFS gateway', () => {
    expect(CERT_IPFS_GATEWAY).toBeDefined();
    expect(CERT_IPFS_GATEWAY).toContain('ipfs');
  });
});

describe('Attestation Constants', () => {
  it('should have max recipients per attestation', () => {
    expect(MAX_RECIPIENTS_PER_ATTESTATION).toBe(50);
  });
});

describe('Encryption Constants', () => {
  it('should use AES-256-GCM', () => {
    expect(ENCRYPTION_ALGORITHM).toBe('AES-256-GCM');
  });
});

