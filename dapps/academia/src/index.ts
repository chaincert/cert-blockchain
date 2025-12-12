/**
 * Academia Reference dApp - Academic Credential Management
 * Per Whitepaper Section 11 - Public diplomas and private transcripts
 * Demonstrates both public EAS attestations and encrypted attestations
 */

import { CertClient, Encryption, IPFS, EncryptedAttestation } from '@certblockchain/sdk';
import { ethers } from 'ethers';

// Credential types
export const CREDENTIAL_TYPES = {
  DIPLOMA: ethers.keccak256(ethers.toUtf8Bytes('DIPLOMA')),
  TRANSCRIPT: ethers.keccak256(ethers.toUtf8Bytes('TRANSCRIPT')),
  CERTIFICATE: ethers.keccak256(ethers.toUtf8Bytes('CERTIFICATE')),
  DEGREE: ethers.keccak256(ethers.toUtf8Bytes('DEGREE')),
} as const;

// Schema UIDs
export const PUBLIC_CREDENTIAL_SCHEMA = '0x0000000000000000000000000000000000000000000000000000000000000003';
export const PRIVATE_CREDENTIAL_SCHEMA = '0x0000000000000000000000000000000000000000000000000000000000000004';

/**
 * Public credential interface (diploma, certificate)
 */
export interface PublicCredential {
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

/**
 * Private credential interface (transcript)
 */
export interface PrivateTranscript {
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

export interface Course {
  code: string;
  name: string;
  credits: number;
  grade: string;
  semester: string;
}

/**
 * Academia dApp client for credential management
 */
export class AcademiaDApp {
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
   * Issue a public credential (diploma/certificate)
   * This creates a public, on-chain attestation visible to anyone
   */
  async issuePublicCredential(credential: PublicCredential): Promise<{
    attestationUID: string;
    txHash: string;
  }> {
    // Create public attestation via standard EAS
    const attestationData = {
      ...credential,
      issuedAt: Date.now(),
      issuedBy: await this.signer.getAddress(),
      version: '1.0',
    };

    const result = await this.client.createAttestation({
      schemaUID: PUBLIC_CREDENTIAL_SCHEMA,
      data: attestationData,
      recipient: credential.studentAddress,
      revocable: true,
      expirationTime: 0, // Credentials don't expire
    });

    // Register on contract
    const contract = new ethers.Contract(
      this.contractAddress,
      ['function createPublicCredential(bytes32,address,bytes32,string,string,uint256,bytes32) external'],
      this.signer
    );

    const signature = await this.signer.signMessage(
      ethers.solidityPackedKeccak256(
        ['bytes32', 'address'],
        [result.uid, credential.studentAddress]
      )
    );

    const tx = await contract.createPublicCredential(
      result.uid,
      credential.studentAddress,
      CREDENTIAL_TYPES[credential.credentialType],
      credential.credentialName,
      credential.fieldOfStudy,
      credential.conferredDate,
      ethers.keccak256(signature)
    );
    const receipt = await tx.wait();

    return { attestationUID: result.uid, txHash: receipt.hash };
  }

  /**
   * Issue a private transcript (encrypted, selective sharing)
   * Only the student and explicitly authorized parties can view
   */
  async issuePrivateTranscript(
    transcript: PrivateTranscript,
    studentPublicKey: string
  ): Promise<{
    attestationUID: string;
    ipfsCID: string;
    txHash: string;
  }> {
    // Prepare recipient (student only initially)
    const recipientPublicKeys = new Map<string, string>();
    recipientPublicKeys.set(transcript.studentAddress, studentPublicKey);

    // Create encrypted attestation
    const result = await this.encryptedAttestation.create(
      {
        schemaUID: PRIVATE_CREDENTIAL_SCHEMA,
        data: transcript,
        recipients: [{ address: transcript.studentAddress }],
        revocable: true,
        expirationTime: 0,
      },
      recipientPublicKeys,
      this.signer
    );

    // Register on contract
    const contract = new ethers.Contract(
      this.contractAddress,
      ['function createPrivateCredential(bytes32,address,bytes32,bytes32) external'],
      this.signer
    );

    const dataHash = ethers.keccak256(ethers.toUtf8Bytes(JSON.stringify(transcript)));

    const tx = await contract.createPrivateCredential(
      result.uid,
      transcript.studentAddress,
      ethers.toUtf8Bytes(result.ipfsCID).slice(0, 32),
      dataHash
    );
    const receipt = await tx.wait();

    return {
      attestationUID: result.uid,
      ipfsCID: result.ipfsCID,
      txHash: receipt.hash,
    };
  }

  /**
   * Grant an employer access to view transcript
   */
  async grantAccessToEmployer(employerAddress: string, durationDays: number): Promise<string> {
    const contract = new ethers.Contract(
      this.contractAddress,
      ['function grantAccess(address,uint256) external'],
      this.signer
    );

    const tx = await contract.grantAccess(employerAddress, durationDays);
    const receipt = await tx.wait();
    return receipt.hash;
  }

  /**
   * Retrieve and decrypt transcript (for authorized parties)
   */
  async retrieveTranscript(attestationUID: string, privateKey: string): Promise<PrivateTranscript> {
    const attestation = await this.encryptedAttestation.retrieve(attestationUID, privateKey, this.signer);
    return attestation.data as PrivateTranscript;
  }
}

export default AcademiaDApp;

