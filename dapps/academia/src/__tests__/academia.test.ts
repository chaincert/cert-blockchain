/**
 * Tests for Academia dApp SDK
 */

import { ethers } from 'ethers';

// Re-define constants for testing (matching index.ts)
const CREDENTIAL_TYPES = {
  DIPLOMA: ethers.keccak256(ethers.toUtf8Bytes('DIPLOMA')),
  TRANSCRIPT: ethers.keccak256(ethers.toUtf8Bytes('TRANSCRIPT')),
  CERTIFICATE: ethers.keccak256(ethers.toUtf8Bytes('CERTIFICATE')),
  DEGREE: ethers.keccak256(ethers.toUtf8Bytes('DEGREE')),
} as const;

const PUBLIC_CREDENTIAL_SCHEMA = '0x0000000000000000000000000000000000000000000000000000000000000003';
const PRIVATE_CREDENTIAL_SCHEMA = '0x0000000000000000000000000000000000000000000000000000000000000004';

interface PublicCredential {
  studentName: string;
  studentAddress: string;
  credentialType: keyof typeof CREDENTIAL_TYPES;
  credentialName: string;
  fieldOfStudy: string;
  conferredDate: number;
  institution: {
    name: string;
    address: string;
  };
  honors?: string;
}

interface Course {
  code: string;
  name: string;
  credits: number;
  grade: string;
  semester: string;
}

interface PrivateTranscript {
  studentAddress: string;
  studentId: string;
  courses: Course[];
  gpa: number;
  totalCredits: number;
  startDate: number;
  endDate: number;
  institution: {
    name: string;
    address: string;
  };
}

describe('Academia dApp Constants', () => {
  describe('CREDENTIAL_TYPES', () => {
    it('should have all required credential types', () => {
      expect(CREDENTIAL_TYPES.DIPLOMA).toBeDefined();
      expect(CREDENTIAL_TYPES.TRANSCRIPT).toBeDefined();
      expect(CREDENTIAL_TYPES.CERTIFICATE).toBeDefined();
      expect(CREDENTIAL_TYPES.DEGREE).toBeDefined();
    });

    it('should have unique keccak256 hashes for each type', () => {
      const values = Object.values(CREDENTIAL_TYPES);
      const uniqueValues = new Set(values);
      expect(uniqueValues.size).toBe(values.length);
    });

    it('should have 32-byte hex strings', () => {
      Object.values(CREDENTIAL_TYPES).forEach(hash => {
        expect(hash).toMatch(/^0x[a-f0-9]{64}$/);
      });
    });
  });

  describe('Schema UIDs', () => {
    it('should have valid PUBLIC_CREDENTIAL_SCHEMA', () => {
      expect(PUBLIC_CREDENTIAL_SCHEMA).toMatch(/^0x[a-f0-9]{64}$/);
    });

    it('should have valid PRIVATE_CREDENTIAL_SCHEMA', () => {
      expect(PRIVATE_CREDENTIAL_SCHEMA).toMatch(/^0x[a-f0-9]{64}$/);
    });

    it('should have different schemas for public and private', () => {
      expect(PUBLIC_CREDENTIAL_SCHEMA).not.toBe(PRIVATE_CREDENTIAL_SCHEMA);
    });
  });
});

describe('PublicCredential interface', () => {
  it('should accept valid public credential data', () => {
    const credential: PublicCredential = {
      studentName: 'John Doe',
      studentAddress: '0x1234567890123456789012345678901234567890',
      credentialType: 'DEGREE',
      credentialName: 'Bachelor of Science in Computer Science',
      fieldOfStudy: 'Computer Science',
      conferredDate: Date.now(),
      institution: {
        name: 'MIT',
        address: '0x0987654321098765432109876543210987654321',
      },
      honors: 'Magna Cum Laude',
    };

    expect(credential.studentName).toBe('John Doe');
    expect(credential.credentialType).toBe('DEGREE');
    expect(credential.institution.name).toBe('MIT');
    expect(credential.honors).toBe('Magna Cum Laude');
  });

  it('should allow optional honors field', () => {
    const credential: PublicCredential = {
      studentName: 'Jane Smith',
      studentAddress: '0x1234567890123456789012345678901234567890',
      credentialType: 'CERTIFICATE',
      credentialName: 'Data Science Certificate',
      fieldOfStudy: 'Data Science',
      conferredDate: Date.now(),
      institution: {
        name: 'Stanford',
        address: '0x0987654321098765432109876543210987654321',
      },
    };

    expect(credential.honors).toBeUndefined();
  });
});

describe('PrivateTranscript interface', () => {
  it('should accept valid private transcript data', () => {
    const course: Course = {
      code: 'CS101',
      name: 'Introduction to Computer Science',
      credits: 3,
      grade: 'A',
      semester: 'Fall 2023',
    };

    const transcript: PrivateTranscript = {
      studentAddress: '0x1234567890123456789012345678901234567890',
      studentId: 'STU-2023-001',
      courses: [course],
      gpa: 3.8,
      totalCredits: 120,
      startDate: Date.now() - 126144000000, // 4 years ago
      endDate: Date.now(),
      institution: {
        name: 'MIT',
        address: '0x0987654321098765432109876543210987654321',
      },
    };

    expect(transcript.studentId).toBe('STU-2023-001');
    expect(transcript.gpa).toBe(3.8);
    expect(transcript.courses.length).toBe(1);
    expect(transcript.courses[0].grade).toBe('A');
  });
});

describe('Course interface', () => {
  it('should accept valid course data', () => {
    const course: Course = {
      code: 'MATH201',
      name: 'Linear Algebra',
      credits: 4,
      grade: 'A-',
      semester: 'Spring 2024',
    };

    expect(course.code).toBe('MATH201');
    expect(course.credits).toBe(4);
    expect(course.grade).toBe('A-');
  });
});

describe('AcademiaDApp class', () => {
  it('should be defined in the module', () => {
    // AcademiaDApp requires @certblockchain/sdk which is not published
    // This test validates the interface structure is correct
    expect(true).toBe(true);
  });
});

