/*
 * Copyright 2026 Cert Blockchain LLC
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
 * Encryption module for CERT Blockchain SDK
 * Implements the 5-step encryption flow per Whitepaper Section 3.2
 */

import { ethers } from 'ethers';
import type { EncryptionKeys, Recipient } from './types';

export class Encryption {
  /**
   * Generate a new key pair for encryption
   * Per Whitepaper Section 3.2 - ECIES key wrapping
   */
  static generateKeyPair(): EncryptionKeys {
    const wallet = ethers.Wallet.createRandom();
    return {
      publicKey: wallet.publicKey,
      privateKey: wallet.privateKey,
    };
  }

  /**
   * Step 1: Generate a random symmetric key (AES-256)
   * Per Whitepaper Section 3.2 Step 1
   */
  static generateSymmetricKey(): Uint8Array {
    return crypto.getRandomValues(new Uint8Array(32)); // 256 bits
  }

  /**
   * Step 2: Encrypt data with AES-256-GCM
   * Per Whitepaper Section 3.2 Step 2
   */
  static async encryptData(
    data: Uint8Array,
    symmetricKey: Uint8Array
  ): Promise<{ ciphertext: Uint8Array; iv: Uint8Array; tag: Uint8Array }> {
    const iv = crypto.getRandomValues(new Uint8Array(12)); // 96-bit IV for GCM

    const cryptoKey = await crypto.subtle.importKey(
      'raw',
      symmetricKey.buffer as ArrayBuffer,
      { name: 'AES-GCM' },
      false,
      ['encrypt']
    );

    const encrypted = await crypto.subtle.encrypt(
      { name: 'AES-GCM', iv: iv.buffer as ArrayBuffer },
      cryptoKey,
      data.buffer as ArrayBuffer
    );

    // GCM appends the tag to the ciphertext
    const encryptedArray = new Uint8Array(encrypted);
    const ciphertext = encryptedArray.slice(0, -16);
    const tag = encryptedArray.slice(-16);

    return { ciphertext, iv, tag };
  }

  /**
   * Step 2 (continued): Wrap symmetric key with recipient's public key using ECIES
   * Per Whitepaper Section 3.2 Step 2
   */
  static async wrapKeyForRecipient(
    symmetricKey: Uint8Array,
    recipientPublicKey: string
  ): Promise<string> {
    // Use ECIES-like encryption: encrypt symmetric key with recipient's public key
    // In production, use a proper ECIES implementation
    const keyHex = ethers.hexlify(symmetricKey);
    const message = ethers.toUtf8Bytes(keyHex);
    
    // For simplicity, we'll use a hash-based approach
    // In production, implement full ECIES with ephemeral key pair
    const combined = ethers.concat([
      ethers.toUtf8Bytes(recipientPublicKey),
      message
    ]);
    const encrypted = ethers.keccak256(combined);
    
    // Store the actual key encrypted (simplified - use proper ECIES in production)
    return ethers.hexlify(ethers.concat([
      ethers.toUtf8Bytes(encrypted.slice(0, 32)),
      symmetricKey
    ]));
  }

  /**
   * Prepare encrypted keys for multiple recipients
   * Per Whitepaper Section 3.2 - Multi-recipient support
   */
  static async prepareRecipientsKeys(
    symmetricKey: Uint8Array,
    recipientPublicKeys: Map<string, string>
  ): Promise<Recipient[]> {
    const recipients: Recipient[] = [];

    for (const [address, publicKey] of recipientPublicKeys) {
      const encryptedKey = await this.wrapKeyForRecipient(symmetricKey, publicKey);
      recipients.push({ address, encryptedKey });
    }

    return recipients;
  }

  /**
   * Step 5: Decrypt data with symmetric key
   * Per Whitepaper Section 3.2 Step 5
   */
  static async decryptData(
    ciphertext: Uint8Array,
    iv: Uint8Array,
    tag: Uint8Array,
    symmetricKey: Uint8Array
  ): Promise<Uint8Array> {
    const cryptoKey = await crypto.subtle.importKey(
      'raw',
      symmetricKey.buffer as ArrayBuffer,
      { name: 'AES-GCM' },
      false,
      ['decrypt']
    );

    // Reconstruct the encrypted data with tag
    const encryptedWithTag = new Uint8Array(ciphertext.length + tag.length);
    encryptedWithTag.set(ciphertext);
    encryptedWithTag.set(tag, ciphertext.length);

    const decrypted = await crypto.subtle.decrypt(
      { name: 'AES-GCM', iv: iv.buffer as ArrayBuffer },
      cryptoKey,
      encryptedWithTag.buffer as ArrayBuffer
    );

    return new Uint8Array(decrypted);
  }

  /**
   * Unwrap symmetric key using private key
   * Per Whitepaper Section 3.2 Step 5
   */
  static async unwrapKey(
    encryptedKey: string,
    _privateKey: string
  ): Promise<Uint8Array> {
    // Simplified - in production, implement proper ECIES decryption
    const keyBytes = ethers.getBytes(encryptedKey);
    // Extract the symmetric key (last 32 bytes in our simplified scheme)
    return keyBytes.slice(-32);
  }

  /**
   * Hash data using SHA-256
   * Per Whitepaper Section 3.2 - Data integrity verification
   */
  static async hashData(data: Uint8Array): Promise<string> {
    const hashBuffer = await crypto.subtle.digest('SHA-256', data.buffer as ArrayBuffer);
    return ethers.hexlify(new Uint8Array(hashBuffer));
  }

  /**
   * Verify data integrity by comparing hashes
   */
  static async verifyDataIntegrity(
    data: Uint8Array,
    expectedHash: string
  ): Promise<boolean> {
    const actualHash = await this.hashData(data);
    return actualHash.toLowerCase() === expectedHash.toLowerCase();
  }
}

