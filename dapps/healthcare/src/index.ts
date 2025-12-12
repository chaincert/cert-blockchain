/**
 * Healthcare Reference dApp - TypeScript SDK Integration
 * Per Whitepaper Section 11 - HIPAA-compliant medical record management
 * Demonstrates the EncryptedFileAttestation schema
 */

import { CertClient, Encryption, IPFS, EncryptedAttestation } from '@certblockchain/sdk';
import { ethers } from 'ethers';

// Medical record types per HIPAA standards
export const RECORD_TYPES = {
  LAB_RESULT: ethers.keccak256(ethers.toUtf8Bytes('LAB_RESULT')),
  IMAGING: ethers.keccak256(ethers.toUtf8Bytes('IMAGING')),
  PRESCRIPTION: ethers.keccak256(ethers.toUtf8Bytes('PRESCRIPTION')),
  CLINICAL_NOTE: ethers.keccak256(ethers.toUtf8Bytes('CLINICAL_NOTE')),
  DISCHARGE_SUMMARY: ethers.keccak256(ethers.toUtf8Bytes('DISCHARGE_SUMMARY')),
  IMMUNIZATION: ethers.keccak256(ethers.toUtf8Bytes('IMMUNIZATION')),
} as const;

// Healthcare attestation schema UID (would be registered on-chain)
export const HEALTHCARE_SCHEMA_UID = '0x0000000000000000000000000000000000000000000000000000000000000001';

/**
 * Medical record interface
 */
export interface MedicalRecord {
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

/**
 * Healthcare dApp client for medical record attestations
 */
export class HealthcareDApp {
  private client: CertClient;
  private ipfs: IPFS;
  private encryptedAttestation: EncryptedAttestation;
  private contractAddress: string;

  constructor(
    apiUrl: string,
    ipfsUrl: string,
    contractAddress: string,
    private signer: ethers.Signer
  ) {
    this.client = new CertClient({ apiUrl });
    this.ipfs = new IPFS(ipfsUrl);
    this.encryptedAttestation = new EncryptedAttestation(apiUrl, this.ipfs);
    this.contractAddress = contractAddress;
  }

  /**
   * Create an encrypted medical record attestation
   * Following the 5-step encryption flow per Whitepaper Section 3.2
   * 
   * @param record Medical record data
   * @param patientPublicKey Patient's public key for encryption
   * @param additionalRecipients Optional additional recipients (specialists, etc.)
   */
  async createMedicalRecord(
    record: MedicalRecord,
    patientPublicKey: string,
    additionalRecipients: Map<string, string> = new Map()
  ): Promise<{ attestationUID: string; ipfsCID: string }> {
    // Validate HIPAA-required fields
    if (!record.patientId || !record.recordType || !record.data) {
      throw new Error('Missing required HIPAA fields');
    }

    // Prepare recipient public keys (patient + any authorized parties)
    const recipientPublicKeys = new Map<string, string>();
    recipientPublicKeys.set(record.patientId, patientPublicKey);
    additionalRecipients.forEach((pubKey, address) => {
      recipientPublicKeys.set(address, pubKey);
    });

    // Add audit metadata
    const attestationData = {
      ...record,
      createdAt: Date.now(),
      createdBy: await this.signer.getAddress(),
      version: '1.0',
      hipaaCompliant: true,
    };

    // Create encrypted attestation
    const result = await this.encryptedAttestation.create(
      {
        schemaUID: HEALTHCARE_SCHEMA_UID,
        data: attestationData,
        recipients: Array.from(recipientPublicKeys.keys()).map(addr => ({ address: addr })),
        revocable: true,
        expirationTime: 0, // No expiration for medical records
      },
      recipientPublicKeys,
      this.signer
    );

    console.log(`Medical record created: ${result.uid}`);
    return {
      attestationUID: result.uid,
      ipfsCID: result.ipfsCID,
    };
  }

  /**
   * Retrieve and decrypt a medical record
   * Patient or authorized party only
   * 
   * @param attestationUID The attestation UID
   * @param privateKey Recipient's private key for decryption
   */
  async retrieveMedicalRecord(
    attestationUID: string,
    privateKey: string
  ): Promise<MedicalRecord> {
    const attestation = await this.encryptedAttestation.retrieve(
      attestationUID,
      privateKey,
      this.signer
    );

    return attestation.data as MedicalRecord;
  }

  /**
   * Grant consent for another party to access records
   * 
   * @param authorizedAddress Address to grant access
   * @param recordTypes Types of records to authorize
   * @param durationDays Duration of consent in days
   */
  async grantConsent(
    authorizedAddress: string,
    recordTypes: (keyof typeof RECORD_TYPES)[],
    durationDays: number
  ): Promise<string> {
    const contract = new ethers.Contract(
      this.contractAddress,
      ['function grantConsent(address,uint256,bytes32[]) external'],
      this.signer
    );

    const recordTypeBytes = recordTypes.map(t => RECORD_TYPES[t]);
    const durationSeconds = durationDays * 24 * 60 * 60;

    const tx = await contract.grantConsent(authorizedAddress, durationSeconds, recordTypeBytes);
    const receipt = await tx.wait();
    
    console.log(`Consent granted to ${authorizedAddress} for ${durationDays} days`);
    return receipt.hash;
  }

  /**
   * Revoke consent for a party
   */
  async revokeConsent(authorizedAddress: string): Promise<string> {
    const contract = new ethers.Contract(
      this.contractAddress,
      ['function revokeConsent(address) external'],
      this.signer
    );

    const tx = await contract.revokeConsent(authorizedAddress);
    const receipt = await tx.wait();
    
    console.log(`Consent revoked for ${authorizedAddress}`);
    return receipt.hash;
  }
}

export default HealthcareDApp;

