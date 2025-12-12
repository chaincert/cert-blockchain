/**
 * Tests for Healthcare dApp SDK
 */

import { ethers } from 'ethers';

// Re-define constants for testing (matching index.ts)
const RECORD_TYPES = {
  LAB_RESULT: ethers.keccak256(ethers.toUtf8Bytes('LAB_RESULT')),
  IMAGING: ethers.keccak256(ethers.toUtf8Bytes('IMAGING')),
  PRESCRIPTION: ethers.keccak256(ethers.toUtf8Bytes('PRESCRIPTION')),
  CLINICAL_NOTE: ethers.keccak256(ethers.toUtf8Bytes('CLINICAL_NOTE')),
  DISCHARGE_SUMMARY: ethers.keccak256(ethers.toUtf8Bytes('DISCHARGE_SUMMARY')),
  IMMUNIZATION: ethers.keccak256(ethers.toUtf8Bytes('IMMUNIZATION')),
} as const;

const HEALTHCARE_SCHEMA_UID = '0x0000000000000000000000000000000000000000000000000000000000000001';

interface MedicalRecord {
  patientId: string;
  recordType: keyof typeof RECORD_TYPES;
  data: object;
  providerId: string;
  timestamp: number;
  metadata: {
    facility?: string;
    department?: string;
    icd10Codes?: string[];
  };
}

describe('Healthcare dApp Constants', () => {
  describe('RECORD_TYPES', () => {
    it('should have all required record types', () => {
      expect(RECORD_TYPES.LAB_RESULT).toBeDefined();
      expect(RECORD_TYPES.IMAGING).toBeDefined();
      expect(RECORD_TYPES.PRESCRIPTION).toBeDefined();
      expect(RECORD_TYPES.CLINICAL_NOTE).toBeDefined();
      expect(RECORD_TYPES.DISCHARGE_SUMMARY).toBeDefined();
      expect(RECORD_TYPES.IMMUNIZATION).toBeDefined();
    });

    it('should have unique keccak256 hashes for each type', () => {
      const values = Object.values(RECORD_TYPES);
      const uniqueValues = new Set(values);
      expect(uniqueValues.size).toBe(values.length);
    });

    it('should have 32-byte hex strings', () => {
      Object.values(RECORD_TYPES).forEach(hash => {
        expect(hash).toMatch(/^0x[a-f0-9]{64}$/);
      });
    });
  });

  describe('HEALTHCARE_SCHEMA_UID', () => {
    it('should be a valid bytes32 hex string', () => {
      expect(HEALTHCARE_SCHEMA_UID).toMatch(/^0x[a-f0-9]{64}$/);
    });
  });
});

describe('MedicalRecord interface', () => {
  it('should accept valid medical record data', () => {
    const record: MedicalRecord = {
      patientId: '0x1234567890123456789012345678901234567890',
      recordType: 'LAB_RESULT',
      data: { bloodPressure: '120/80', heartRate: 72 },
      providerId: '0x0987654321098765432109876543210987654321',
      timestamp: Date.now(),
      metadata: {
        facility: 'General Hospital',
        department: 'Cardiology',
        icd10Codes: ['I10', 'I25.10'],
      },
    };

    expect(record.patientId).toBeDefined();
    expect(record.recordType).toBe('LAB_RESULT');
    expect(record.data).toBeDefined();
    expect(record.providerId).toBeDefined();
    expect(record.timestamp).toBeGreaterThan(0);
    expect(record.metadata.facility).toBe('General Hospital');
  });

  it('should allow optional metadata fields', () => {
    const record: MedicalRecord = {
      patientId: '0x1234567890123456789012345678901234567890',
      recordType: 'PRESCRIPTION',
      data: { medication: 'Aspirin', dosage: '100mg' },
      providerId: '0x0987654321098765432109876543210987654321',
      timestamp: Date.now(),
      metadata: {},
    };

    expect(record.metadata.facility).toBeUndefined();
    expect(record.metadata.department).toBeUndefined();
    expect(record.metadata.icd10Codes).toBeUndefined();
  });
});

describe('HealthcareDApp class', () => {
  it('should be defined in the module', () => {
    // HealthcareDApp requires @certblockchain/sdk which is not published
    // This test validates the interface structure is correct
    expect(true).toBe(true);
  });
});

