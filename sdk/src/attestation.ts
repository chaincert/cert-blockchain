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
 * Encrypted Attestation module for CERT Blockchain SDK
 * Implements the complete 5-step encryption flow per Whitepaper Section 3.2
 */

import { ethers } from 'ethers';
import type {
  EncryptedAttestationData,
  CreateEncryptedAttestationRequest,
} from './types';
import { Encryption } from './encryption';
import { IPFS } from './ipfs';
import { MAX_RECIPIENTS_PER_ATTESTATION } from './constants';

export class EncryptedAttestation {
  private apiUrl: string;
  private ipfs: IPFS;

  constructor(apiUrl: string, ipfs: IPFS) {
    this.apiUrl = apiUrl;
    this.ipfs = ipfs;
  }

  /**
   * Create an encrypted attestation following the 5-step flow
   * Per Whitepaper Section 3.2
   * 
   * @param request - Attestation creation request
   * @param recipientPublicKeys - Map of recipient addresses to their public keys
   * @param signer - Ethers signer for transaction signing
   * @returns Created attestation data
   */
  async create(
    request: CreateEncryptedAttestationRequest,
    recipientPublicKeys: Map<string, string>,
    signer: ethers.Signer
  ): Promise<EncryptedAttestationData> {
    // Validate recipients
    if (request.recipients.length === 0) {
      throw new Error('At least one recipient required');
    }
    if (request.recipients.length > MAX_RECIPIENTS_PER_ATTESTATION) {
      throw new Error(`Maximum ${MAX_RECIPIENTS_PER_ATTESTATION} recipients allowed`);
    }

    // Step 1: Generate symmetric key
    const symmetricKey = Encryption.generateSymmetricKey();

    // Step 2: Encrypt data with AES-256-GCM
    const dataBytes = new TextEncoder().encode(JSON.stringify(request.data));
    const { ciphertext, iv, tag } = await Encryption.encryptData(dataBytes, symmetricKey);

    // Combine encrypted data for storage
    const encryptedData = new Uint8Array(iv.length + ciphertext.length + tag.length);
    encryptedData.set(iv);
    encryptedData.set(ciphertext, iv.length);
    encryptedData.set(tag, iv.length + ciphertext.length);

    // Step 2 (continued): Wrap symmetric key for each recipient
    const recipients = await Encryption.prepareRecipientsKeys(
      symmetricKey,
      recipientPublicKeys
    );

    // Step 3: Upload encrypted data to IPFS
    const ipfsCID = await this.ipfs.upload(encryptedData);

    // Calculate hash of encrypted data for integrity verification
    const encryptedDataHash = await Encryption.hashData(encryptedData);

    // Step 4: Anchor on-chain via API
    const signature = await this.signAttestationRequest(signer, {
      schemaUID: request.schemaUID,
      ipfsCID,
      encryptedDataHash,
      recipients: recipients.map(r => r.address),
    });

    const response = await fetch(`${this.apiUrl}/api/v1/encrypted-attestations`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        schemaUID: request.schemaUID,
        ipfsCID,
        encryptedDataHash,
        recipients,
        revocable: request.revocable ?? true,
        expirationTime: request.expirationTime,
        signature,
      }),
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.error || 'Failed to create attestation');
    }

    return response.json();
  }

  /**
   * Retrieve and decrypt an attestation
   * Per Whitepaper Section 3.2 Step 5
   * 
   * @param uid - Attestation UID
   * @param privateKey - Recipient's private key for decryption
   * @returns Decrypted attestation data
   */
  async retrieve(
    uid: string,
    privateKey: string,
    signer: ethers.Signer
  ): Promise<{ attestation: EncryptedAttestationData; data: Record<string, unknown> }> {
    const requester = await signer.getAddress();
    const signature = await this.signRetrievalRequest(signer, uid);

    // Get attestation metadata and encrypted key
    const response = await fetch(`${this.apiUrl}/api/v1/encrypted-attestations/${uid}/retrieve`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ requester, signature }),
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.error || 'Failed to retrieve attestation');
    }

    const { ipfsCID, encryptedKey } = await response.json();

    // Step 5a: Retrieve encrypted data from IPFS
    const encryptedData = await this.ipfs.retrieve(ipfsCID);

    // Step 5b: Unwrap symmetric key
    const symmetricKey = await Encryption.unwrapKey(encryptedKey, privateKey);

    // Step 5c: Decrypt data
    const iv = encryptedData.slice(0, 12);
    const tag = encryptedData.slice(-16);
    const ciphertext = encryptedData.slice(12, -16);

    const decryptedBytes = await Encryption.decryptData(ciphertext, iv, tag, symmetricKey);
    const decryptedData = JSON.parse(new TextDecoder().decode(decryptedBytes));

    // Get full attestation metadata
    const attestationResponse = await fetch(`${this.apiUrl}/api/v1/encrypted-attestations/${uid}`);
    const attestation = await attestationResponse.json();

    return { attestation, data: decryptedData };
  }

  /**
   * Revoke an attestation
   * 
   * @param uid - Attestation UID
   * @param signer - Attester's signer
   */
  async revoke(uid: string, signer: ethers.Signer): Promise<void> {
    const attester = await signer.getAddress();
    const signature = await this.signRevocationRequest(signer, uid);

    const response = await fetch(`${this.apiUrl}/api/v1/encrypted-attestations/${uid}/revoke`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ attester, signature }),
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.error || 'Failed to revoke attestation');
    }
  }

  /**
   * Get attestation by UID
   */
  async get(uid: string): Promise<EncryptedAttestationData> {
    const response = await fetch(`${this.apiUrl}/api/v1/encrypted-attestations/${uid}`);
    if (!response.ok) {
      throw new Error('Attestation not found');
    }
    return response.json();
  }

  /**
   * Get attestations by attester
   */
  async getByAttester(address: string): Promise<EncryptedAttestationData[]> {
    const response = await fetch(`${this.apiUrl}/api/v1/attestations/by-attester/${address}`);
    const result = await response.json();
    return result.attestations || [];
  }

  /**
   * Get attestations by recipient
   */
  async getByRecipient(address: string): Promise<EncryptedAttestationData[]> {
    const response = await fetch(`${this.apiUrl}/api/v1/attestations/by-recipient/${address}`);
    const result = await response.json();
    return result.attestations || [];
  }

  private async signAttestationRequest(signer: ethers.Signer, data: object): Promise<string> {
    const message = JSON.stringify(data);
    return signer.signMessage(message);
  }

  private async signRetrievalRequest(signer: ethers.Signer, uid: string): Promise<string> {
    const message = `Retrieve attestation: ${uid}`;
    return signer.signMessage(message);
  }

  private async signRevocationRequest(signer: ethers.Signer, uid: string): Promise<string> {
    const message = `Revoke attestation: ${uid}`;
    return signer.signMessage(message);
  }
}

