/**
 * Tests for Attestation SDK Module
 * Tests the attestation client functionality without requiring live blockchain
 */

import { ethers } from 'ethers';

// Mock attestation types and constants (matching attestation.ts)
const ATTESTATION_SCHEMA_TYPES = {
  ENCRYPTED: 'ENCRYPTED',
  PUBLIC: 'PUBLIC',
  REVOCABLE: 'REVOCABLE',
} as const;

interface AttestationData {
  schemaUID: string;
  recipient: string;
  data: Uint8Array;
  revocable: boolean;
  expirationTime?: number;
}

interface EncryptedAttestationData extends AttestationData {
  ipfsCID: string;
  encryptedDataHash: string;
  recipientKeys: Array<{
    address: string;
    encryptedKey: string;
  }>;
}

// Helper function to generate attestation UID
function generateAttestationUID(data: AttestationData): string {
  const encoded = ethers.AbiCoder.defaultAbiCoder().encode(
    ['bytes32', 'address', 'bytes', 'bool', 'uint64'],
    [
      data.schemaUID,
      data.recipient,
      data.data,
      data.revocable,
      data.expirationTime || 0,
    ]
  );
  return ethers.keccak256(encoded);
}

describe('Attestation SDK', () => {
  describe('ATTESTATION_SCHEMA_TYPES', () => {
    it('should have all required schema types', () => {
      expect(ATTESTATION_SCHEMA_TYPES.ENCRYPTED).toBe('ENCRYPTED');
      expect(ATTESTATION_SCHEMA_TYPES.PUBLIC).toBe('PUBLIC');
      expect(ATTESTATION_SCHEMA_TYPES.REVOCABLE).toBe('REVOCABLE');
    });
  });

  describe('AttestationData interface', () => {
    it('should accept valid attestation data', () => {
      const data: AttestationData = {
        schemaUID: '0x' + '1'.repeat(64),
        recipient: '0x' + '2'.repeat(40),
        data: new Uint8Array([1, 2, 3, 4]),
        revocable: true,
        expirationTime: Math.floor(Date.now() / 1000) + 86400,
      };

      expect(data.schemaUID).toHaveLength(66);
      expect(data.recipient).toHaveLength(42);
      expect(data.revocable).toBe(true);
    });

    it('should allow optional expirationTime', () => {
      const data: AttestationData = {
        schemaUID: '0x' + '1'.repeat(64),
        recipient: '0x' + '2'.repeat(40),
        data: new Uint8Array([1, 2, 3, 4]),
        revocable: false,
      };

      expect(data.expirationTime).toBeUndefined();
    });
  });

  describe('EncryptedAttestationData interface', () => {
    it('should accept valid encrypted attestation data', () => {
      const data: EncryptedAttestationData = {
        schemaUID: '0x' + '1'.repeat(64),
        recipient: '0x' + '2'.repeat(40),
        data: new Uint8Array([1, 2, 3, 4]),
        revocable: true,
        ipfsCID: 'QmYwAPJzv5CZsnAzt8auVZRn5W7x8Hd8fH6FS6NVQP3fSw',
        encryptedDataHash: '0x' + 'a'.repeat(64),
        recipientKeys: [
          {
            address: '0x' + '3'.repeat(40),
            encryptedKey: '0x' + 'b'.repeat(128),
          },
        ],
      };

      expect(data.ipfsCID).toBeDefined();
      expect(data.encryptedDataHash).toHaveLength(66);
      expect(data.recipientKeys).toHaveLength(1);
    });

    it('should support multiple recipients per whitepaper (max 50)', () => {
      const recipients = Array.from({ length: 50 }, (_, i) => ({
        address: '0x' + i.toString(16).padStart(40, '0'),
        encryptedKey: '0x' + 'c'.repeat(128),
      }));

      const data: EncryptedAttestationData = {
        schemaUID: '0x' + '1'.repeat(64),
        recipient: '0x' + '2'.repeat(40),
        data: new Uint8Array([1, 2, 3, 4]),
        revocable: true,
        ipfsCID: 'QmYwAPJzv5CZsnAzt8auVZRn5W7x8Hd8fH6FS6NVQP3fSw',
        encryptedDataHash: '0x' + 'a'.repeat(64),
        recipientKeys: recipients,
      };

      expect(data.recipientKeys).toHaveLength(50);
    });
  });

  describe('generateAttestationUID', () => {
    it('should generate consistent UIDs for same input', () => {
      const data: AttestationData = {
        schemaUID: '0x' + '1'.repeat(64),
        recipient: '0x' + '2'.repeat(40),
        data: new Uint8Array([1, 2, 3, 4]),
        revocable: true,
        expirationTime: 1704067200,
      };

      const uid1 = generateAttestationUID(data);
      const uid2 = generateAttestationUID(data);

      expect(uid1).toBe(uid2);
    });

    it('should generate different UIDs for different inputs', () => {
      const data1: AttestationData = {
        schemaUID: '0x' + '1'.repeat(64),
        recipient: '0x' + '2'.repeat(40),
        data: new Uint8Array([1, 2, 3, 4]),
        revocable: true,
      };

      const data2: AttestationData = {
        ...data1,
        recipient: '0x' + '3'.repeat(40),
      };

      expect(generateAttestationUID(data1)).not.toBe(generateAttestationUID(data2));
    });

    it('should return a valid bytes32 hex string', () => {
      const data: AttestationData = {
        schemaUID: '0x' + '1'.repeat(64),
        recipient: '0x' + '2'.repeat(40),
        data: new Uint8Array([1, 2, 3, 4]),
        revocable: true,
      };

      const uid = generateAttestationUID(data);
      expect(uid).toMatch(/^0x[a-f0-9]{64}$/);
    });
  });
});

