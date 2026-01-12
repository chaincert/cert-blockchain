import { ethers } from 'ethers';

/**
 * Type definitions for CERT Blockchain SDK
 * Per Whitepaper Section 3 and 8
 */
interface ClientConfig {
    rpcUrl: string;
    apiUrl: string;
    ipfsUrl?: string;
    chainId?: string;
}
interface AttestationData {
    uid: string;
    schemaUID: string;
    attester: string;
    recipient: string;
    data: Record<string, unknown>;
    time: number;
    expirationTime?: number;
    revocationTime?: number;
    revocable: boolean;
    refUID?: string;
}
interface EncryptedAttestationData {
    uid: string;
    schemaUID: string;
    attester: string;
    ipfsCID: string;
    encryptedDataHash: string;
    recipients: Recipient[];
    revocable: boolean;
    revoked: boolean;
    expirationTime?: number;
    createdAt: number;
}
interface Recipient {
    address: string;
    encryptedKey: string;
}
interface Schema {
    uid: string;
    creator: string;
    schema: string;
    resolver?: string;
    revocable: boolean;
    createdAt: number;
}
interface CertIDProfile {
    address: string;
    name?: string;
    handle?: string;
    bio?: string;
    avatarUrl?: string;
    metadataURI?: string;
    publicKey?: string;
    socialLinks?: Record<string, string>;
    credentials?: string[];
    badges?: string[];
    verified: boolean;
    isVerified?: boolean;
    verificationLevel: number;
    trustScore?: number;
    entityType?: EntityType;
    isActive?: boolean;
    createdAt: number;
    updatedAt: number;
}
declare enum EntityType {
    Individual = 0,
    Institution = 1,
    SystemAdmin = 2,
    Bot = 3
}
type BadgeType = 'KYC_L1' | 'KYC_L2' | 'ACADEMIC_ISSUER' | 'VERIFIED_CREATOR' | 'GOV_AGENCY' | 'LEGAL_ENTITY' | 'ISO_9001_CERTIFIED';
interface FullIdentity {
    address: string;
    handle: string;
    metadata: string;
    isVerified: boolean;
    isInstitutional: boolean;
    trustScore: number;
    entityType: EntityType;
    badges: string[];
    isKYC: boolean;
    isAcademic: boolean;
    isCreator: boolean;
}
interface EncryptionKeys {
    publicKey: string;
    privateKey: string;
}
interface IPFSConfig {
    url: string;
    gateway?: string;
}
interface CreateEncryptedAttestationRequest {
    schemaUID: string;
    data: Record<string, unknown>;
    recipients: string[];
    revocable?: boolean;
    expirationTime?: number;
}
interface RegisterSchemaRequest {
    schema: string;
    resolver?: string;
    revocable?: boolean;
}
interface UpdateProfileRequest {
    name?: string;
    bio?: string;
    avatarUrl?: string;
    publicKey?: string;
    socialLinks?: Record<string, string>;
}
interface VerifySocialRequest {
    platform: string;
    handle: string;
    proof: string;
}

/**
 * IPFS module for CERT Blockchain SDK
 * Handles encrypted file storage per Whitepaper Section 3.2 Step 3
 */

declare class IPFS {
    private url;
    private gateway;
    private timeout;
    constructor(config?: IPFSConfig);
    /**
     * Step 3: Upload encrypted data to IPFS
     * Per Whitepaper Section 3.2 Step 3
     *
     * @param data - Encrypted data to upload
     * @returns IPFS CID (Content Identifier)
     */
    upload(data: Uint8Array): Promise<string>;
    /**
     * Step 5: Retrieve encrypted data from IPFS
     * Per Whitepaper Section 3.2 Step 5
     *
     * @param cid - IPFS Content Identifier
     * @returns Encrypted data
     */
    retrieve(cid: string): Promise<Uint8Array>;
    /**
     * Pin content to ensure persistence
     *
     * @param cid - IPFS Content Identifier to pin
     */
    pin(cid: string): Promise<void>;
    /**
     * Unpin content
     *
     * @param cid - IPFS Content Identifier to unpin
     */
    unpin(cid: string): Promise<void>;
    /**
     * Get the gateway URL for a CID
     *
     * @param cid - IPFS Content Identifier
     * @returns Gateway URL
     */
    getGatewayUrl(cid: string): string;
    /**
     * Validate IPFS CID format
     *
     * @param cid - CID to validate
     * @returns Whether the CID is valid
     */
    isValidCID(cid: string): boolean;
    /**
     * Set custom timeout
     *
     * @param timeout - Timeout in milliseconds
     */
    setTimeout(timeout: number): void;
}

/**
 * Encrypted Attestation module for CERT Blockchain SDK
 * Implements the complete 5-step encryption flow per Whitepaper Section 3.2
 */

declare class EncryptedAttestation {
    private apiUrl;
    private ipfs;
    constructor(apiUrl: string, ipfs: IPFS);
    /**
     * Create an encrypted attestation following the 5-step flow
     * Per Whitepaper Section 3.2
     *
     * @param request - Attestation creation request
     * @param recipientPublicKeys - Map of recipient addresses to their public keys
     * @param signer - Ethers signer for transaction signing
     * @returns Created attestation data
     */
    create(request: CreateEncryptedAttestationRequest, recipientPublicKeys: Map<string, string>, signer: ethers.Signer): Promise<EncryptedAttestationData>;
    /**
     * Retrieve and decrypt an attestation
     * Per Whitepaper Section 3.2 Step 5
     *
     * @param uid - Attestation UID
     * @param privateKey - Recipient's private key for decryption
     * @returns Decrypted attestation data
     */
    retrieve(uid: string, privateKey: string, signer: ethers.Signer): Promise<{
        attestation: EncryptedAttestationData;
        data: Record<string, unknown>;
    }>;
    /**
     * Revoke an attestation
     *
     * @param uid - Attestation UID
     * @param signer - Attester's signer
     */
    revoke(uid: string, signer: ethers.Signer): Promise<void>;
    /**
     * Get attestation by UID
     */
    get(uid: string): Promise<EncryptedAttestationData>;
    /**
     * Get attestations by attester
     */
    getByAttester(address: string): Promise<EncryptedAttestationData[]>;
    /**
     * Get attestations by recipient
     */
    getByRecipient(address: string): Promise<EncryptedAttestationData[]>;
    private signAttestationRequest;
    private signRetrievalRequest;
    private signRevocationRequest;
}

/**
 * CertID module for CERT Blockchain SDK
 * Implements decentralized identity per Whitepaper CertID Section
 * Includes Soulbound Token (SBT) badge support and trust scores
 */

declare class CertID {
    private apiUrl;
    private contract;
    constructor(apiUrl: string, contract?: ethers.Contract);
    /**
     * Set the CertID contract instance for direct blockchain queries
     */
    setContract(contract: ethers.Contract): void;
    /**
     * Get a CertID profile by address
     *
     * @param address - Blockchain address
     * @returns CertID profile
     */
    getProfile(address: string): Promise<CertIDProfile>;
    /**
     * Create or update a CertID profile
     *
     * @param profile - Profile data to update
     * @param signer - Ethers signer for authentication
     * @returns Updated profile
     */
    updateProfile(profile: UpdateProfileRequest, signer: ethers.Signer): Promise<CertIDProfile>;
    /**
     * Verify a social media account
     *
     * @param request - Social verification request
     * @param signer - Ethers signer for authentication
     * @returns Verification result
     */
    verifySocial(request: VerifySocialRequest, signer: ethers.Signer): Promise<{
        verified: boolean;
        platform: string;
    }>;
    /**
     * Add a credential to the profile
     *
     * @param attestationUID - UID of the credential attestation
     * @param signer - Ethers signer for authentication
     */
    addCredential(attestationUID: string, signer: ethers.Signer): Promise<void>;
    /**
     * Remove a credential from the profile
     *
     * @param attestationUID - UID of the credential attestation
     * @param signer - Ethers signer for authentication
     */
    removeCredential(attestationUID: string, signer: ethers.Signer): Promise<void>;
    /**
     * Get authentication challenge for signing
     *
     * @param address - Address to authenticate
     * @returns Challenge message
     */
    getAuthChallenge(address: string): Promise<{
        challenge: string;
        expiresAt: string;
    }>;
    /**
     * Verify a signed authentication challenge
     *
     * @param address - Address that signed
     * @param challenge - Challenge message
     * @param signature - Signature
     * @returns Verification result
     */
    verifyAuth(address: string, challenge: string, signature: string): Promise<{
        verified: boolean;
    }>;
    /**
     * Generate a verification message for social platforms
     *
     * @param address - User's address
     * @param platform - Social platform name
     * @returns Verification message to post
     */
    generateSocialProof(address: string, platform: string): string;
    private signProfileUpdate;
    private signSocialVerification;
    /**
     * Get full identity including badges and trust score
     * @param address - Wallet address to look up
     */
    getFullIdentity(address: string): Promise<FullIdentity | null>;
    /**
     * Check all standard badges for an address
     */
    checkStandardBadges(address: string): Promise<string[]>;
    /**
     * Check if address has a specific badge
     */
    hasBadge(address: string, badgeName: BadgeType): Promise<boolean>;
    /**
     * Get trust score for an address
     */
    getTrustScore(address: string): Promise<number>;
    /**
     * Resolve a handle to an address
     */
    resolveHandle(handle: string): Promise<string | null>;
    /**
     * Get detailed profile with full identity resolution
     * This is the primary method for displaying identity info in block explorer
     * @param address - Wallet address to look up
     * @returns Detailed profile with display-ready information
     */
    getDetailedProfile(address: string): Promise<{
        address: string;
        displayName: string;
        handle: string | null;
        avatarUrl: string | null;
        isVerified: boolean;
        isVerifiedInstitution: boolean;
        trustScore: number;
        badges: Array<{
            id: string;
            name: string;
            icon: string;
        }>;
        entityType: string;
        profileUrl: string;
    }>;
    /**
     * Truncate address for display
     */
    private truncateAddress;
}

/**
 * Main client for CERT Blockchain SDK
 * Per Whitepaper Section 8 - SDK and API
 */

declare class CertClient {
    private config;
    private provider;
    private ipfs;
    attestation: EncryptedAttestation;
    certid: CertID;
    constructor(config?: Partial<ClientConfig>);
    /**
     * Get the JSON-RPC provider
     */
    getProvider(): ethers.JsonRpcProvider;
    /**
     * Get the IPFS client
     */
    getIPFS(): IPFS;
    /**
     * Connect a signer (wallet) to the client
     *
     * @param signer - Ethers signer
     * @returns Connected signer
     */
    connectSigner(signer: ethers.Signer): ethers.Signer;
    /**
     * Create a signer from a private key
     *
     * @param privateKey - Private key
     * @returns Wallet signer
     */
    createSigner(privateKey: string): ethers.Wallet;
    /**
     * Register a new schema
     *
     * @param request - Schema registration request
     * @param signer - Signer for the transaction
     * @returns Registered schema
     */
    registerSchema(request: RegisterSchemaRequest, signer: ethers.Signer): Promise<Schema>;
    /**
     * Get a schema by UID
     *
     * @param uid - Schema UID
     * @returns Schema data
     */
    getSchema(uid: string): Promise<Schema>;
    /**
     * Get the current block number
     */
    getBlockNumber(): Promise<number>;
    /**
     * Get the balance of an address
     *
     * @param address - Address to check
     * @returns Balance in wei
     */
    getBalance(address: string): Promise<bigint>;
    /**
     * Get the chain ID
     */
    getChainId(): Promise<bigint>;
    /**
     * Check if the client is connected to the correct network
     */
    isCorrectNetwork(): Promise<boolean>;
    /**
     * Wait for a transaction to be confirmed
     *
     * @param txHash - Transaction hash
     * @param confirmations - Number of confirmations to wait for
     * @returns Transaction receipt
     */
    waitForTransaction(txHash: string, confirmations?: number): Promise<ethers.TransactionReceipt | null>;
    /**
     * Get health status of the API
     */
    getHealth(): Promise<{
        status: string;
        service: string;
    }>;
}

declare class Encryption {
    /**
     * Generate a new key pair for encryption
     * Per Whitepaper Section 3.2 - ECIES key wrapping
     */
    static generateKeyPair(): EncryptionKeys;
    /**
     * Step 1: Generate a random symmetric key (AES-256)
     * Per Whitepaper Section 3.2 Step 1
     */
    static generateSymmetricKey(): Uint8Array;
    /**
     * Step 2: Encrypt data with AES-256-GCM
     * Per Whitepaper Section 3.2 Step 2
     */
    static encryptData(data: Uint8Array, symmetricKey: Uint8Array): Promise<{
        ciphertext: Uint8Array;
        iv: Uint8Array;
        tag: Uint8Array;
    }>;
    /**
     * Step 2 (continued): Wrap symmetric key with recipient's public key using ECIES
     * Per Whitepaper Section 3.2 Step 2
     */
    static wrapKeyForRecipient(symmetricKey: Uint8Array, recipientPublicKey: string): Promise<string>;
    /**
     * Prepare encrypted keys for multiple recipients
     * Per Whitepaper Section 3.2 - Multi-recipient support
     */
    static prepareRecipientsKeys(symmetricKey: Uint8Array, recipientPublicKeys: Map<string, string>): Promise<Recipient[]>;
    /**
     * Step 5: Decrypt data with symmetric key
     * Per Whitepaper Section 3.2 Step 5
     */
    static decryptData(ciphertext: Uint8Array, iv: Uint8Array, tag: Uint8Array, symmetricKey: Uint8Array): Promise<Uint8Array>;
    /**
     * Unwrap symmetric key using private key
     * Per Whitepaper Section 3.2 Step 5
     */
    static unwrapKey(encryptedKey: string, _privateKey: string): Promise<Uint8Array>;
    /**
     * Hash data using SHA-256
     * Per Whitepaper Section 3.2 - Data integrity verification
     */
    static hashData(data: Uint8Array): Promise<string>;
    /**
     * Verify data integrity by comparing hashes
     */
    static verifyDataIntegrity(data: Uint8Array, expectedHash: string): Promise<boolean>;
}

/**
 * Constants for CERT Blockchain SDK
 * Per Whitepaper Sections 4, 5, 9, and 12
 */
declare const CERT_CHAIN_ID = "cert-mainnet-1";
declare const CERT_RPC_URL = "https://rpc.c3rt.org";
declare const CERT_API_URL = "https://api.c3rt.org/api/v1";
declare const CERT_IPFS_GATEWAY = "https://ipfs.c3rt.org";
declare const DEFAULT_SCHEMAS: {
    BUSINESS_DOCUMENT: string;
    IDENTITY_VERIFICATION: string;
    CREDENTIAL: string;
    CERTIFICATE: string;
};
declare const CONTRACT_ADDRESSES: {
    SCHEMA_REGISTRY: string;
    EAS: string;
    ENCRYPTED_ATTESTATION: string;
    CERT_TOKEN: string;
    CERT_ID: string;
    CHAIN_CERTIFY: string;
};
declare const CERT_ID_ABI: string[];
declare const BADGE_TYPES: {
    KYC_L1: string;
    KYC_L2: string;
    ACADEMIC_ISSUER: string;
    VERIFIED_CREATOR: string;
    GOV_AGENCY: string;
    LEGAL_ENTITY: string;
    ISO_9001_CERTIFIED: string;
};

/**
 * Generate a unique identifier (UID) for attestations
 *
 * @param data - Data to hash
 * @returns UID as hex string
 */
declare function generateUID(...data: (string | number | Uint8Array)[]): string;
/**
 * Hash data using keccak256
 *
 * @param data - Data to hash
 * @returns Hash as hex string
 */
declare function hashData(data: string | Uint8Array): string;
/**
 * Validate a blockchain address
 *
 * @param address - Address to validate
 * @returns Whether the address is valid
 */
declare function validateAddress(address: string): boolean;
/**
 * Format CERT amount from ucert (micro CERT)
 *
 * @param amount - Amount in ucert
 * @returns Formatted CERT amount
 */
declare function formatCERT(amount: bigint | string | number): string;
/**
 * Parse CERT amount to ucert (micro CERT)
 *
 * @param amount - Amount in CERT
 * @returns Amount in ucert
 */
declare function parseCERT(amount: string | number): bigint;

export { type AttestationData, BADGE_TYPES, type BadgeType, CERT_API_URL, CERT_CHAIN_ID, CERT_ID_ABI, CERT_IPFS_GATEWAY, CERT_RPC_URL, CONTRACT_ADDRESSES, CertClient, CertID, type CertIDProfile, type ClientConfig, DEFAULT_SCHEMAS, EncryptedAttestation, type EncryptedAttestationData, Encryption, type EncryptionKeys, EntityType, type FullIdentity, IPFS, type IPFSConfig, type Recipient, type Schema, formatCERT, generateUID, hashData, parseCERT, validateAddress };
