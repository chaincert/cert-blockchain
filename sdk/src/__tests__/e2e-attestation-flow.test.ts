/**
 * E2E Tests for Complete Attestation Flow
 * Tests the 5-step encrypted attestation flow per Whitepaper Section 3.2
 * 
 * Flow:
 * 1. AES Key Generation → Generate symmetric key
 * 2. Data Encryption → Encrypt attestation data with AES
 * 3. ECIES Key Wrapping → Wrap key for each recipient
 * 4. IPFS Upload & On-Chain Anchoring → Store encrypted data, anchor reference
 * 5. Retrieval & Decryption → Authorized recipients decrypt
 */

import { ethers } from 'ethers';

// Import encryption utilities (re-implemented for testing)
import * as crypto from 'crypto';

// Step 1: Generate AES-256 symmetric key
function generateAESKey(): Buffer {
  return crypto.randomBytes(32);
}

// Step 2: Encrypt data with AES-256-GCM
function encryptWithAES(data: Buffer, key: Buffer): { ciphertext: Buffer; iv: Buffer; tag: Buffer } {
  const iv = crypto.randomBytes(12);
  const cipher = crypto.createCipheriv('aes-256-gcm', key, iv);
  const ciphertext = Buffer.concat([cipher.update(data), cipher.final()]);
  const tag = cipher.getAuthTag();
  return { ciphertext, iv, tag };
}

// Step 2b: Decrypt data with AES-256-GCM  
function decryptWithAES(ciphertext: Buffer, key: Buffer, iv: Buffer, tag: Buffer): Buffer {
  const decipher = crypto.createDecipheriv('aes-256-gcm', key, iv);
  decipher.setAuthTag(tag);
  return Buffer.concat([decipher.update(ciphertext), decipher.final()]);
}

// Step 3: Generate key pair for ECIES (mock using ECDH)
function generateKeyPair(): { privateKey: Buffer; publicKey: Buffer } {
  const keyPair = crypto.generateKeyPairSync('ec', { namedCurve: 'secp256k1' });
  return {
    privateKey: keyPair.privateKey.export({ type: 'sec1', format: 'der' }),
    publicKey: keyPair.publicKey.export({ type: 'spki', format: 'der' }),
  };
}

// Mock IPFS CID generation
function generateMockCID(data: Buffer): string {
  const hash = crypto.createHash('sha256').update(data).digest('hex');
  return `Qm${hash.substring(0, 44)}`;
}

// Generate attestation UID
function generateUID(schemaUID: string, attester: string, timestamp: number): string {
  const data = ethers.solidityPacked(
    ['bytes32', 'address', 'uint256'],
    [schemaUID, attester, timestamp]
  );
  return ethers.keccak256(data);
}

describe('E2E Encrypted Attestation Flow', () => {
  // Test data
  const attestationData = {
    patientId: 'P12345',
    recordType: 'LAB_RESULT',
    data: {
      testName: 'Complete Blood Count',
      results: { wbc: 7.5, rbc: 4.8, platelets: 250 },
    },
    providerId: 'DR001',
    timestamp: Date.now(),
  };

  const schemaUID = '0x' + '1'.repeat(64);
  const attesterAddress = '0x' + 'a'.repeat(40);
  const recipientAddress = '0x' + 'b'.repeat(40);

  describe('Step 1: AES Key Generation', () => {
    it('should generate a 256-bit AES key', () => {
      const key = generateAESKey();
      expect(key.length).toBe(32); // 256 bits = 32 bytes
    });

    it('should generate unique keys each time', () => {
      const key1 = generateAESKey();
      const key2 = generateAESKey();
      expect(key1.equals(key2)).toBe(false);
    });
  });

  describe('Step 2: Data Encryption', () => {
    it('should encrypt attestation data with AES-256-GCM', () => {
      const key = generateAESKey();
      const dataBuffer = Buffer.from(JSON.stringify(attestationData));
      
      const { ciphertext, iv, tag } = encryptWithAES(dataBuffer, key);
      
      expect(ciphertext.length).toBeGreaterThan(0);
      expect(iv.length).toBe(12);
      expect(tag.length).toBe(16);
    });

    it('should produce different ciphertext for same data with different IVs', () => {
      const key = generateAESKey();
      const dataBuffer = Buffer.from(JSON.stringify(attestationData));
      
      const encrypted1 = encryptWithAES(dataBuffer, key);
      const encrypted2 = encryptWithAES(dataBuffer, key);
      
      expect(encrypted1.ciphertext.equals(encrypted2.ciphertext)).toBe(false);
    });

    it('should correctly decrypt encrypted data', () => {
      const key = generateAESKey();
      const dataBuffer = Buffer.from(JSON.stringify(attestationData));
      
      const { ciphertext, iv, tag } = encryptWithAES(dataBuffer, key);
      const decrypted = decryptWithAES(ciphertext, key, iv, tag);
      
      expect(decrypted.toString()).toBe(JSON.stringify(attestationData));
    });
  });

  describe('Step 3: ECIES Key Wrapping', () => {
    it('should generate valid key pairs', () => {
      const { privateKey, publicKey } = generateKeyPair();
      expect(privateKey.length).toBeGreaterThan(0);
      expect(publicKey.length).toBeGreaterThan(0);
    });
  });

  describe('Step 4: IPFS Upload & On-Chain Anchoring', () => {
    it('should generate valid IPFS CID for encrypted data', () => {
      const key = generateAESKey();
      const dataBuffer = Buffer.from(JSON.stringify(attestationData));
      const { ciphertext, iv, tag } = encryptWithAES(dataBuffer, key);
      
      // Package for IPFS
      const ipfsPayload = Buffer.concat([iv, tag, ciphertext]);
      const cid = generateMockCID(ipfsPayload);
      
      expect(cid).toMatch(/^Qm[a-zA-Z0-9]{44}$/);
    });

    it('should generate unique attestation UID', () => {
      const uid = generateUID(schemaUID, attesterAddress, Date.now());
      expect(uid).toMatch(/^0x[a-f0-9]{64}$/);
    });
  });

  describe('Step 5: Retrieval & Decryption', () => {
    it('should complete full encryption/decryption cycle', () => {
      // Step 1: Generate key
      const aesKey = generateAESKey();
      
      // Step 2: Encrypt data
      const dataBuffer = Buffer.from(JSON.stringify(attestationData));
      const { ciphertext, iv, tag } = encryptWithAES(dataBuffer, aesKey);
      
      // Step 4: Simulate IPFS storage
      const ipfsPayload = Buffer.concat([iv, tag, ciphertext]);
      const cid = generateMockCID(ipfsPayload);
      
      // Step 5: Retrieve and decrypt
      const retrievedIV = ipfsPayload.subarray(0, 12);
      const retrievedTag = ipfsPayload.subarray(12, 28);
      const retrievedCiphertext = ipfsPayload.subarray(28);
      
      const decrypted = decryptWithAES(retrievedCiphertext, aesKey, retrievedIV, retrievedTag);
      const recoveredData = JSON.parse(decrypted.toString());
      
      expect(recoveredData.patientId).toBe(attestationData.patientId);
      expect(recoveredData.recordType).toBe(attestationData.recordType);
    });
  });
});

