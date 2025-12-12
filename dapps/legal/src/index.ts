/**
 * Legal Reference dApp - Confidential Document Sharing
 * Per Whitepaper Section 11 - Multi-recipient legal document management
 * Demonstrates the EncryptedMultiRecipientAttestation schema
 */

import { CertClient, Encryption, IPFS, EncryptedAttestation } from '@certblockchain/sdk';
import { ethers } from 'ethers';

// Document types
export const DOCUMENT_TYPES = {
  CONTRACT: ethers.keccak256(ethers.toUtf8Bytes('CONTRACT')),
  NDA: ethers.keccak256(ethers.toUtf8Bytes('NDA')),
  MERGER: ethers.keccak256(ethers.toUtf8Bytes('MERGER')),
  COURT_FILING: ethers.keccak256(ethers.toUtf8Bytes('COURT_FILING')),
  SETTLEMENT: ethers.keccak256(ethers.toUtf8Bytes('SETTLEMENT')),
  POWER_OF_ATTORNEY: ethers.keccak256(ethers.toUtf8Bytes('POWER_OF_ATTORNEY')),
} as const;

// Legal document schema UID
export const LEGAL_SCHEMA_UID = '0x0000000000000000000000000000000000000000000000000000000000000002';

/**
 * Legal document interface
 */
export interface LegalDocument {
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

export interface Party {
  address: string;
  name: string;
  role: 'creator' | 'signatory' | 'witness' | 'notary';
  publicKey: string;
}

/**
 * Legal dApp client for confidential document management
 */
export class LegalDApp {
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
   * Create an encrypted legal document attestation
   * Multi-recipient sharing per Whitepaper Section 3.2
   * 
   * @param document Legal document data
   */
  async createDocument(document: LegalDocument): Promise<{ 
    attestationUID: string; 
    ipfsCID: string;
    txHash: string;
  }> {
    // Validate parties (max 50 per Whitepaper Section 12)
    if (document.parties.length === 0 || document.parties.length > 50) {
      throw new Error('Invalid party count (1-50 allowed)');
    }

    // Prepare recipient public keys for all parties
    const recipientPublicKeys = new Map<string, string>();
    document.parties.forEach(party => {
      recipientPublicKeys.set(party.address, party.publicKey);
    });

    // Create attestation data
    const attestationData = {
      title: document.title,
      documentType: document.documentType,
      parties: document.parties.map(p => ({ address: p.address, name: p.name, role: p.role })),
      effectiveDate: document.effectiveDate,
      expirationDate: document.expirationDate,
      jurisdiction: document.jurisdiction,
      metadata: document.metadata,
      createdAt: Date.now(),
      createdBy: await this.signer.getAddress(),
      contentHash: ethers.keccak256(
        typeof document.content === 'string' 
          ? ethers.toUtf8Bytes(document.content)
          : document.content
      ),
    };

    // Create encrypted attestation
    const result = await this.encryptedAttestation.create(
      {
        schemaUID: LEGAL_SCHEMA_UID,
        data: attestationData,
        recipients: document.parties.map(p => ({ address: p.address })),
        revocable: true,
        expirationTime: document.expirationDate || 0,
      },
      recipientPublicKeys,
      this.signer
    );

    // Register on-chain
    const contract = new ethers.Contract(
      this.contractAddress,
      ['function createDocument(bytes32,bytes32,address[],uint256,uint256,bytes32,bytes32,uint256) external'],
      this.signer
    );

    const partyAddresses = document.parties.map(p => p.address);
    const requiredSignatures = document.parties.filter(p => p.role === 'signatory').length;

    const tx = await contract.createDocument(
      result.uid,
      DOCUMENT_TYPES[document.documentType],
      partyAddresses,
      document.effectiveDate,
      document.expirationDate || 0,
      ethers.toUtf8Bytes(result.ipfsCID).slice(0, 32),
      attestationData.contentHash,
      requiredSignatures
    );
    const receipt = await tx.wait();

    return {
      attestationUID: result.uid,
      ipfsCID: result.ipfsCID,
      txHash: receipt.hash,
    };
  }

  /**
   * Sign a legal document
   */
  async signDocument(attestationUID: string): Promise<string> {
    // Create signature over document hash
    const messageHash = ethers.keccak256(
      ethers.solidityPacked(['bytes32', 'address', 'uint256'], [attestationUID, await this.signer.getAddress(), Date.now()])
    );
    const signature = await this.signer.signMessage(ethers.getBytes(messageHash));

    const contract = new ethers.Contract(
      this.contractAddress,
      ['function signDocument(bytes32,bytes) external'],
      this.signer
    );

    const tx = await contract.signDocument(attestationUID, signature);
    const receipt = await tx.wait();
    
    return receipt.hash;
  }

  /**
   * Retrieve and decrypt a legal document
   */
  async retrieveDocument(attestationUID: string, privateKey: string): Promise<LegalDocument> {
    const attestation = await this.encryptedAttestation.retrieve(attestationUID, privateKey, this.signer);
    return attestation.data as unknown as LegalDocument;
  }
}

export default LegalDApp;

