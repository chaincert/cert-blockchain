/**
 * Tests for CERT Blockchain SDK utility functions
 */

import {
  generateUID,
  hashData,
  validateAddress,
  formatCERT,
  parseCERT,
  evmToCosmosAddress,
  isValidCID,
  truncateAddress,
  formatBytes,
} from '../utils';

describe('generateUID', () => {
  it('should generate consistent UID for same input', () => {
    const uid1 = generateUID('test', 123);
    const uid2 = generateUID('test', 123);
    expect(uid1).toBe(uid2);
  });

  it('should generate different UIDs for different inputs', () => {
    const uid1 = generateUID('test1');
    const uid2 = generateUID('test2');
    expect(uid1).not.toBe(uid2);
  });

  it('should start with 0x', () => {
    const uid = generateUID('test');
    expect(uid.startsWith('0x')).toBe(true);
  });

  it('should be 66 characters long (0x + 64 hex chars)', () => {
    const uid = generateUID('test');
    expect(uid.length).toBe(66);
  });
});

describe('hashData', () => {
  it('should hash string data', () => {
    const hash = hashData('test');
    expect(hash.startsWith('0x')).toBe(true);
    expect(hash.length).toBe(66);
  });

  it('should hash Uint8Array data', () => {
    const data = new Uint8Array([1, 2, 3, 4]);
    const hash = hashData(data);
    expect(hash.startsWith('0x')).toBe(true);
    expect(hash.length).toBe(66);
  });

  it('should produce consistent hashes', () => {
    const hash1 = hashData('test');
    const hash2 = hashData('test');
    expect(hash1).toBe(hash2);
  });
});

describe('validateAddress', () => {
  it('should validate EVM addresses (checksummed)', () => {
    // Using a properly checksummed address
    expect(validateAddress('0x742d35Cc6634C0532925a3b844Bc454e4438f44e')).toBe(true);
  });

  it('should validate EVM addresses (lowercase)', () => {
    // Lowercase addresses are valid
    expect(validateAddress('0x742d35cc6634c0532925a3b844bc454e4438f44e')).toBe(true);
  });

  it('should reject invalid EVM addresses', () => {
    expect(validateAddress('0xinvalid')).toBe(false);
    expect(validateAddress('0x123')).toBe(false);
  });

  it('should validate Cosmos addresses', () => {
    expect(validateAddress('cert1' + 'a'.repeat(38))).toBe(true);
  });

  it('should reject invalid Cosmos addresses', () => {
    expect(validateAddress('cert1short')).toBe(false);
    expect(validateAddress('cosmos1' + 'a'.repeat(38))).toBe(false);
  });
});

describe('formatCERT', () => {
  it('should format whole amounts', () => {
    expect(formatCERT(1000000n)).toBe('1 CERT');
    expect(formatCERT(10000000n)).toBe('10 CERT');
  });

  it('should format fractional amounts', () => {
    expect(formatCERT(1500000n)).toBe('1.5 CERT');
    expect(formatCERT(1234567n)).toBe('1.234567 CERT');
  });

  it('should handle zero', () => {
    expect(formatCERT(0n)).toBe('0 CERT');
  });

  it('should accept string input', () => {
    expect(formatCERT('1000000')).toBe('1 CERT');
  });
});

describe('parseCERT', () => {
  it('should parse whole amounts', () => {
    expect(parseCERT(1)).toBe(1000000n);
    expect(parseCERT(10)).toBe(10000000n);
  });

  it('should parse fractional amounts', () => {
    expect(parseCERT(1.5)).toBe(1500000n);
  });

  it('should parse string input', () => {
    expect(parseCERT('1')).toBe(1000000n);
  });
});

describe('evmToCosmosAddress', () => {
  it('should convert EVM address to Cosmos format', () => {
    const evmAddr = '0x742d35Cc6634C0532925a3b844Bc9e7595f1E9a0';
    const cosmosAddr = evmToCosmosAddress(evmAddr);
    expect(cosmosAddr.startsWith('cert1')).toBe(true);
  });
});

describe('isValidCID', () => {
  it('should validate CIDv0', () => {
    expect(isValidCID('QmYwAPJzv5CZsnA625s3Xf2nemtYgPpHdWEz79ojWnPbdG')).toBe(true);
  });

  it('should validate CIDv1', () => {
    expect(isValidCID('bafybeigdyrzt5sfp7udm7hu76uh7y26nf3efuylqabf3oclgtqy55fbzdi')).toBe(true);
  });

  it('should reject invalid CIDs', () => {
    expect(isValidCID('invalid')).toBe(false);
    expect(isValidCID('Qmshort')).toBe(false);
  });
});

describe('truncateAddress', () => {
  it('should truncate long addresses', () => {
    const addr = '0x742d35Cc6634C0532925a3b844Bc9e7595f1E9a0';
    expect(truncateAddress(addr)).toBe('0x742d...E9a0');
  });

  it('should not truncate short addresses', () => {
    expect(truncateAddress('short')).toBe('short');
  });
});

describe('formatBytes', () => {
  it('should format bytes', () => {
    expect(formatBytes(0)).toBe('0 Bytes');
    expect(formatBytes(500)).toBe('500 Bytes');
    expect(formatBytes(1024)).toBe('1 KB');
    expect(formatBytes(1048576)).toBe('1 MB');
  });
});

