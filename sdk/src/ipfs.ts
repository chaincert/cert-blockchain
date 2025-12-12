/**
 * IPFS module for CERT Blockchain SDK
 * Handles encrypted file storage per Whitepaper Section 3.2 Step 3
 */

import type { IPFSConfig } from './types';
import { CERT_IPFS_GATEWAY, IPFS_DEFAULT_TIMEOUT, IPFS_MAX_FILE_SIZE } from './constants';

export class IPFS {
  private url: string;
  private gateway: string;
  private timeout: number;

  constructor(config?: IPFSConfig) {
    this.url = config?.url || 'http://localhost:5001';
    this.gateway = config?.gateway || CERT_IPFS_GATEWAY;
    this.timeout = IPFS_DEFAULT_TIMEOUT;
  }

  /**
   * Step 3: Upload encrypted data to IPFS
   * Per Whitepaper Section 3.2 Step 3
   * 
   * @param data - Encrypted data to upload
   * @returns IPFS CID (Content Identifier)
   */
  async upload(data: Uint8Array): Promise<string> {
    if (data.length > IPFS_MAX_FILE_SIZE) {
      throw new Error(`File size exceeds maximum allowed (${IPFS_MAX_FILE_SIZE} bytes)`);
    }

    const formData = new FormData();
    const blob = new Blob([data.buffer as ArrayBuffer], { type: 'application/octet-stream' });
    formData.append('file', blob);

    const controller = new AbortController();
    const timeoutId = setTimeout(() => controller.abort(), this.timeout);

    try {
      const response = await fetch(`${this.url}/api/v0/add`, {
        method: 'POST',
        body: formData,
        signal: controller.signal,
      });

      clearTimeout(timeoutId);

      if (!response.ok) {
        throw new Error(`IPFS upload failed: ${response.statusText}`);
      }

      const result = await response.json();
      return result.Hash;
    } catch (error) {
      clearTimeout(timeoutId);
      if (error instanceof Error && error.name === 'AbortError') {
        throw new Error('IPFS upload timed out');
      }
      throw error;
    }
  }

  /**
   * Step 5: Retrieve encrypted data from IPFS
   * Per Whitepaper Section 3.2 Step 5
   * 
   * @param cid - IPFS Content Identifier
   * @returns Encrypted data
   */
  async retrieve(cid: string): Promise<Uint8Array> {
    if (!this.isValidCID(cid)) {
      throw new Error('Invalid IPFS CID');
    }

    const controller = new AbortController();
    const timeoutId = setTimeout(() => controller.abort(), this.timeout);

    try {
      // Try gateway first for better performance
      const response = await fetch(`${this.gateway}/ipfs/${cid}`, {
        signal: controller.signal,
      });

      clearTimeout(timeoutId);

      if (!response.ok) {
        throw new Error(`IPFS retrieval failed: ${response.statusText}`);
      }

      const arrayBuffer = await response.arrayBuffer();
      return new Uint8Array(arrayBuffer);
    } catch (error) {
      clearTimeout(timeoutId);
      if (error instanceof Error && error.name === 'AbortError') {
        throw new Error('IPFS retrieval timed out');
      }
      throw error;
    }
  }

  /**
   * Pin content to ensure persistence
   * 
   * @param cid - IPFS Content Identifier to pin
   */
  async pin(cid: string): Promise<void> {
    if (!this.isValidCID(cid)) {
      throw new Error('Invalid IPFS CID');
    }

    const response = await fetch(`${this.url}/api/v0/pin/add?arg=${cid}`, {
      method: 'POST',
    });

    if (!response.ok) {
      throw new Error(`IPFS pin failed: ${response.statusText}`);
    }
  }

  /**
   * Unpin content
   * 
   * @param cid - IPFS Content Identifier to unpin
   */
  async unpin(cid: string): Promise<void> {
    if (!this.isValidCID(cid)) {
      throw new Error('Invalid IPFS CID');
    }

    const response = await fetch(`${this.url}/api/v0/pin/rm?arg=${cid}`, {
      method: 'POST',
    });

    if (!response.ok) {
      throw new Error(`IPFS unpin failed: ${response.statusText}`);
    }
  }

  /**
   * Get the gateway URL for a CID
   * 
   * @param cid - IPFS Content Identifier
   * @returns Gateway URL
   */
  getGatewayUrl(cid: string): string {
    return `${this.gateway}/ipfs/${cid}`;
  }

  /**
   * Validate IPFS CID format
   * 
   * @param cid - CID to validate
   * @returns Whether the CID is valid
   */
  isValidCID(cid: string): boolean {
    // CIDv0 starts with Qm and is 46 characters
    // CIDv1 starts with b and varies in length
    if (cid.startsWith('Qm') && cid.length === 46) {
      return true;
    }
    if (cid.startsWith('b') && cid.length >= 50) {
      return true;
    }
    return false;
  }

  /**
   * Set custom timeout
   * 
   * @param timeout - Timeout in milliseconds
   */
  setTimeout(timeout: number): void {
    this.timeout = timeout;
  }
}

