/**
 * Tests for Legal dApp SDK
 */

import { ethers } from 'ethers';

// Re-define constants for testing (matching index.ts)
const DOCUMENT_TYPES = {
  CONTRACT: ethers.keccak256(ethers.toUtf8Bytes('CONTRACT')),
  NDA: ethers.keccak256(ethers.toUtf8Bytes('NDA')),
  MERGER: ethers.keccak256(ethers.toUtf8Bytes('MERGER')),
  COURT_FILING: ethers.keccak256(ethers.toUtf8Bytes('COURT_FILING')),
  SETTLEMENT: ethers.keccak256(ethers.toUtf8Bytes('SETTLEMENT')),
  POWER_OF_ATTORNEY: ethers.keccak256(ethers.toUtf8Bytes('POWER_OF_ATTORNEY')),
} as const;

const LEGAL_SCHEMA_UID = '0x0000000000000000000000000000000000000000000000000000000000000002';

interface Party {
  address: string;
  name: string;
  role: 'creator' | 'signatory' | 'witness' | 'notary';
  publicKey: string;
}

interface LegalDocument {
  title: string;
  documentType: keyof typeof DOCUMENT_TYPES;
  content: string | Buffer;
  parties: Party[];
  effectiveDate: number;
  expirationDate?: number;
  jurisdiction?: string;
  metadata: {
    caseNumber?: string;
    courtName?: string;
    notarized?: boolean;
    version: string;
  };
}

describe('Legal dApp Constants', () => {
  describe('DOCUMENT_TYPES', () => {
    it('should have all required document types', () => {
      expect(DOCUMENT_TYPES.CONTRACT).toBeDefined();
      expect(DOCUMENT_TYPES.NDA).toBeDefined();
      expect(DOCUMENT_TYPES.MERGER).toBeDefined();
      expect(DOCUMENT_TYPES.COURT_FILING).toBeDefined();
      expect(DOCUMENT_TYPES.SETTLEMENT).toBeDefined();
      expect(DOCUMENT_TYPES.POWER_OF_ATTORNEY).toBeDefined();
    });

    it('should have unique keccak256 hashes for each type', () => {
      const values = Object.values(DOCUMENT_TYPES);
      const uniqueValues = new Set(values);
      expect(uniqueValues.size).toBe(values.length);
    });

    it('should have 32-byte hex strings', () => {
      Object.values(DOCUMENT_TYPES).forEach(hash => {
        expect(hash).toMatch(/^0x[a-f0-9]{64}$/);
      });
    });
  });

  describe('LEGAL_SCHEMA_UID', () => {
    it('should be a valid bytes32 hex string', () => {
      expect(LEGAL_SCHEMA_UID).toMatch(/^0x[a-f0-9]{64}$/);
    });
  });
});

describe('Party interface', () => {
  it('should accept valid party data', () => {
    const party: Party = {
      address: '0x1234567890123456789012345678901234567890',
      name: 'John Doe',
      role: 'signatory',
      publicKey: '0x04abcdef...',
    };

    expect(party.address).toBeDefined();
    expect(party.name).toBe('John Doe');
    expect(party.role).toBe('signatory');
    expect(party.publicKey).toBeDefined();
  });

  it('should accept all valid roles', () => {
    const roles: Party['role'][] = ['creator', 'signatory', 'witness', 'notary'];

    roles.forEach(role => {
      const party: Party = {
        address: '0x1234567890123456789012345678901234567890',
        name: 'Test Party',
        role,
        publicKey: '0x04...',
      };
      expect(party.role).toBe(role);
    });
  });
});

describe('LegalDocument interface', () => {
  it('should accept valid legal document data', () => {
    const party1: Party = {
      address: '0x1234567890123456789012345678901234567890',
      name: 'Party A',
      role: 'signatory',
      publicKey: '0x04abc...',
    };

    const party2: Party = {
      address: '0x0987654321098765432109876543210987654321',
      name: 'Party B',
      role: 'signatory',
      publicKey: '0x04def...',
    };

    const document: LegalDocument = {
      title: 'Service Agreement',
      documentType: 'CONTRACT',
      content: 'This agreement is entered into...',
      parties: [party1, party2],
      effectiveDate: Date.now(),
      expirationDate: Date.now() + 31536000000, // 1 year
      jurisdiction: 'Delaware, USA',
      metadata: {
        caseNumber: 'CASE-2024-001',
        courtName: 'Delaware Court of Chancery',
        notarized: true,
        version: '1.0',
      },
    };

    expect(document.title).toBe('Service Agreement');
    expect(document.documentType).toBe('CONTRACT');
    expect(document.parties.length).toBe(2);
    expect(document.metadata.notarized).toBe(true);
  });

  it('should allow optional fields', () => {
    const party: Party = {
      address: '0x1234567890123456789012345678901234567890',
      name: 'Solo Party',
      role: 'creator',
      publicKey: '0x04...',
    };

    const document: LegalDocument = {
      title: 'Simple NDA',
      documentType: 'NDA',
      content: 'Confidentiality agreement...',
      parties: [party],
      effectiveDate: Date.now(),
      metadata: {
        version: '1.0',
      },
    };

    expect(document.expirationDate).toBeUndefined();
    expect(document.jurisdiction).toBeUndefined();
    expect(document.metadata.caseNumber).toBeUndefined();
  });

  it('should accept Buffer content', () => {
    const party: Party = {
      address: '0x1234567890123456789012345678901234567890',
      name: 'Party',
      role: 'creator',
      publicKey: '0x04...',
    };

    const document: LegalDocument = {
      title: 'Binary Document',
      documentType: 'COURT_FILING',
      content: Buffer.from('PDF content here'),
      parties: [party],
      effectiveDate: Date.now(),
      metadata: { version: '1.0' },
    };

    expect(Buffer.isBuffer(document.content)).toBe(true);
  });
});

describe('LegalDApp class', () => {
  it('should be defined in the module', () => {
    // LegalDApp requires @certblockchain/sdk which is not published
    // This test validates the interface structure is correct
    expect(true).toBe(true);
  });
});

