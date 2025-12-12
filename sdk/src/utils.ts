/**
 * Utility functions for CERT Blockchain SDK
 */

import { ethers } from 'ethers';
import { CERT_DECIMALS } from './constants';

/**
 * Generate a unique identifier (UID) for attestations
 * 
 * @param data - Data to hash
 * @returns UID as hex string
 */
export function generateUID(...data: (string | number | Uint8Array)[]): string {
  const combined = data.map(d => {
    if (typeof d === 'number') return d.toString();
    if (d instanceof Uint8Array) return ethers.hexlify(d);
    return d;
  }).join('');
  
  return ethers.keccak256(ethers.toUtf8Bytes(combined));
}

/**
 * Hash data using keccak256
 * 
 * @param data - Data to hash
 * @returns Hash as hex string
 */
export function hashData(data: string | Uint8Array): string {
  if (typeof data === 'string') {
    return ethers.keccak256(ethers.toUtf8Bytes(data));
  }
  return ethers.keccak256(data);
}

/**
 * Validate a blockchain address
 * 
 * @param address - Address to validate
 * @returns Whether the address is valid
 */
export function validateAddress(address: string): boolean {
  try {
    // Check for EVM address
    if (address.startsWith('0x')) {
      return ethers.isAddress(address);
    }
    // Check for Cosmos address (cert prefix)
    if (address.startsWith('cert1')) {
      return address.length === 43; // cert1 + 38 chars
    }
    return false;
  } catch {
    return false;
  }
}

/**
 * Format CERT amount from ucert (micro CERT)
 * 
 * @param amount - Amount in ucert
 * @returns Formatted CERT amount
 */
export function formatCERT(amount: bigint | string | number): string {
  const value = BigInt(amount);
  const divisor = BigInt(10 ** CERT_DECIMALS);
  const whole = value / divisor;
  const fraction = value % divisor;
  
  if (fraction === 0n) {
    return `${whole} CERT`;
  }
  
  const fractionStr = fraction.toString().padStart(CERT_DECIMALS, '0');
  const trimmedFraction = fractionStr.replace(/0+$/, '');
  
  return `${whole}.${trimmedFraction} CERT`;
}

/**
 * Parse CERT amount to ucert (micro CERT)
 * 
 * @param amount - Amount in CERT
 * @returns Amount in ucert
 */
export function parseCERT(amount: string | number): bigint {
  const value = typeof amount === 'string' ? parseFloat(amount) : amount;
  return BigInt(Math.floor(value * 10 ** CERT_DECIMALS));
}

/**
 * Convert EVM address to Cosmos address
 * 
 * @param evmAddress - EVM address (0x...)
 * @returns Cosmos address (cert1...)
 */
export function evmToCosmosAddress(evmAddress: string): string {
  // This is a simplified conversion - in production use proper bech32 encoding
  const addressBytes = ethers.getBytes(evmAddress);
  const hash = ethers.keccak256(addressBytes);
  return `cert1${hash.slice(2, 40)}`;
}

/**
 * Sleep for a specified duration
 * 
 * @param ms - Milliseconds to sleep
 */
export function sleep(ms: number): Promise<void> {
  return new Promise(resolve => setTimeout(resolve, ms));
}

/**
 * Retry a function with exponential backoff
 * 
 * @param fn - Function to retry
 * @param maxRetries - Maximum number of retries
 * @param baseDelay - Base delay in milliseconds
 * @returns Result of the function
 */
export async function retry<T>(
  fn: () => Promise<T>,
  maxRetries: number = 3,
  baseDelay: number = 1000
): Promise<T> {
  let lastError: Error | undefined;
  
  for (let i = 0; i < maxRetries; i++) {
    try {
      return await fn();
    } catch (error) {
      lastError = error as Error;
      if (i < maxRetries - 1) {
        await sleep(baseDelay * Math.pow(2, i));
      }
    }
  }
  
  throw lastError;
}

/**
 * Validate IPFS CID format
 * 
 * @param cid - CID to validate
 * @returns Whether the CID is valid
 */
export function isValidCID(cid: string): boolean {
  // CIDv0 starts with Qm and is 46 characters
  if (cid.startsWith('Qm') && cid.length === 46) {
    return true;
  }
  // CIDv1 starts with b and varies in length
  if (cid.startsWith('b') && cid.length >= 50) {
    return true;
  }
  return false;
}

/**
 * Truncate an address for display
 * 
 * @param address - Address to truncate
 * @param chars - Number of characters to show on each side
 * @returns Truncated address
 */
export function truncateAddress(address: string, chars: number = 4): string {
  if (address.length <= chars * 2 + 3) {
    return address;
  }
  return `${address.slice(0, chars + 2)}...${address.slice(-chars)}`;
}

/**
 * Convert bytes to human-readable size
 * 
 * @param bytes - Number of bytes
 * @returns Human-readable size string
 */
export function formatBytes(bytes: number): string {
  if (bytes === 0) return '0 Bytes';
  
  const k = 1024;
  const sizes = ['Bytes', 'KB', 'MB', 'GB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  
  return `${parseFloat((bytes / Math.pow(k, i)).toFixed(2))} ${sizes[i]}`;
}

