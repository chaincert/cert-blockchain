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
 * Tests for CERT Blockchain SDK encryption module
 * Tests the 5-step encryption flow per Whitepaper Section 3.2
 */

import { Encryption } from '../encryption';

describe('Encryption', () => {
  describe('generateKeyPair', () => {
    it('should generate a valid key pair', () => {
      const keys = Encryption.generateKeyPair();
      expect(keys.publicKey).toBeDefined();
      expect(keys.privateKey).toBeDefined();
      expect(keys.publicKey.startsWith('0x')).toBe(true);
      expect(keys.privateKey.startsWith('0x')).toBe(true);
    });

    it('should generate unique key pairs', () => {
      const keys1 = Encryption.generateKeyPair();
      const keys2 = Encryption.generateKeyPair();
      expect(keys1.publicKey).not.toBe(keys2.publicKey);
      expect(keys1.privateKey).not.toBe(keys2.privateKey);
    });
  });

  describe('generateSymmetricKey', () => {
    it('should generate a 256-bit key', () => {
      const key = Encryption.generateSymmetricKey();
      expect(key).toBeInstanceOf(Uint8Array);
      expect(key.length).toBe(32); // 256 bits = 32 bytes
    });

    it('should generate unique keys', () => {
      const key1 = Encryption.generateSymmetricKey();
      const key2 = Encryption.generateSymmetricKey();
      expect(Buffer.from(key1).toString('hex')).not.toBe(
        Buffer.from(key2).toString('hex')
      );
    });
  });

  describe('encryptData and decryptData', () => {
    it('should encrypt and decrypt data correctly', async () => {
      const originalData = new TextEncoder().encode('Hello, CERT Blockchain!');
      const symmetricKey = Encryption.generateSymmetricKey();

      const { ciphertext, iv, tag } = await Encryption.encryptData(
        originalData,
        symmetricKey
      );

      expect(ciphertext).toBeInstanceOf(Uint8Array);
      expect(iv).toBeInstanceOf(Uint8Array);
      expect(tag).toBeInstanceOf(Uint8Array);
      expect(iv.length).toBe(12); // 96-bit IV
      expect(tag.length).toBe(16); // 128-bit tag

      const decrypted = await Encryption.decryptData(
        ciphertext,
        iv,
        tag,
        symmetricKey
      );

      expect(new TextDecoder().decode(decrypted)).toBe('Hello, CERT Blockchain!');
    });

    it('should produce different ciphertext for same data with different IVs', async () => {
      const data = new TextEncoder().encode('Test data');
      const key = Encryption.generateSymmetricKey();

      const result1 = await Encryption.encryptData(data, key);
      const result2 = await Encryption.encryptData(data, key);

      expect(Buffer.from(result1.ciphertext).toString('hex')).not.toBe(
        Buffer.from(result2.ciphertext).toString('hex')
      );
    });
  });

  describe('wrapKeyForRecipient', () => {
    it('should wrap symmetric key for recipient', async () => {
      const symmetricKey = Encryption.generateSymmetricKey();
      const recipientKeys = Encryption.generateKeyPair();

      const wrappedKey = await Encryption.wrapKeyForRecipient(
        symmetricKey,
        recipientKeys.publicKey
      );

      expect(wrappedKey).toBeDefined();
      expect(wrappedKey.startsWith('0x')).toBe(true);
    });
  });

  describe('unwrapKey', () => {
    it('should unwrap symmetric key', async () => {
      const symmetricKey = Encryption.generateSymmetricKey();
      const recipientKeys = Encryption.generateKeyPair();

      const wrappedKey = await Encryption.wrapKeyForRecipient(
        symmetricKey,
        recipientKeys.publicKey
      );

      const unwrappedKey = await Encryption.unwrapKey(
        wrappedKey,
        recipientKeys.privateKey
      );

      expect(unwrappedKey).toBeInstanceOf(Uint8Array);
      expect(unwrappedKey.length).toBe(32);
    });
  });

  describe('hashData', () => {
    it('should hash data using SHA-256', async () => {
      const data = new TextEncoder().encode('Test data');
      const hash = await Encryption.hashData(data);

      expect(hash.startsWith('0x')).toBe(true);
      expect(hash.length).toBe(66); // 0x + 64 hex chars
    });

    it('should produce consistent hashes', async () => {
      const data = new TextEncoder().encode('Test data');
      const hash1 = await Encryption.hashData(data);
      const hash2 = await Encryption.hashData(data);

      expect(hash1).toBe(hash2);
    });
  });

  describe('verifyDataIntegrity', () => {
    it('should verify data integrity', async () => {
      const data = new TextEncoder().encode('Test data');
      const hash = await Encryption.hashData(data);

      const isValid = await Encryption.verifyDataIntegrity(data, hash);
      expect(isValid).toBe(true);
    });

    it('should reject tampered data', async () => {
      const data = new TextEncoder().encode('Test data');
      const hash = await Encryption.hashData(data);

      const tamperedData = new TextEncoder().encode('Tampered data');
      const isValid = await Encryption.verifyDataIntegrity(tamperedData, hash);
      expect(isValid).toBe(false);
    });
  });

  describe('prepareRecipientsKeys', () => {
    it('should prepare keys for multiple recipients', async () => {
      const symmetricKey = Encryption.generateSymmetricKey();
      const recipient1 = Encryption.generateKeyPair();
      const recipient2 = Encryption.generateKeyPair();

      const recipientPublicKeys = new Map([
        ['0x1111111111111111111111111111111111111111', recipient1.publicKey],
        ['0x2222222222222222222222222222222222222222', recipient2.publicKey],
      ]);

      const recipients = await Encryption.prepareRecipientsKeys(
        symmetricKey,
        recipientPublicKeys
      );

      expect(recipients.length).toBe(2);
      expect(recipients[0].address).toBe('0x1111111111111111111111111111111111111111');
      expect(recipients[1].address).toBe('0x2222222222222222222222222222222222222222');
      expect(recipients[0].encryptedKey).toBeDefined();
      expect(recipients[1].encryptedKey).toBeDefined();
    });
  });
});

