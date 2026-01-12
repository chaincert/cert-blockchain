"use strict";
var __defProp = Object.defineProperty;
var __getOwnPropDesc = Object.getOwnPropertyDescriptor;
var __getOwnPropNames = Object.getOwnPropertyNames;
var __hasOwnProp = Object.prototype.hasOwnProperty;
var __export = (target, all) => {
  for (var name in all)
    __defProp(target, name, { get: all[name], enumerable: true });
};
var __copyProps = (to, from, except, desc) => {
  if (from && typeof from === "object" || typeof from === "function") {
    for (let key of __getOwnPropNames(from))
      if (!__hasOwnProp.call(to, key) && key !== except)
        __defProp(to, key, { get: () => from[key], enumerable: !(desc = __getOwnPropDesc(from, key)) || desc.enumerable });
  }
  return to;
};
var __toCommonJS = (mod) => __copyProps(__defProp({}, "__esModule", { value: true }), mod);

// src/index.ts
var index_exports = {};
__export(index_exports, {
  BADGE_TYPES: () => BADGE_TYPES,
  CERT_API_URL: () => CERT_API_URL,
  CERT_CHAIN_ID: () => CERT_CHAIN_ID,
  CERT_ID_ABI: () => CERT_ID_ABI,
  CERT_IPFS_GATEWAY: () => CERT_IPFS_GATEWAY,
  CERT_RPC_URL: () => CERT_RPC_URL,
  CONTRACT_ADDRESSES: () => CONTRACT_ADDRESSES,
  CertClient: () => CertClient,
  CertID: () => CertID,
  DEFAULT_SCHEMAS: () => DEFAULT_SCHEMAS,
  EncryptedAttestation: () => EncryptedAttestation,
  Encryption: () => Encryption,
  EntityType: () => EntityType,
  IPFS: () => IPFS,
  formatCERT: () => formatCERT,
  generateUID: () => generateUID,
  hashData: () => hashData,
  parseCERT: () => parseCERT,
  validateAddress: () => validateAddress
});
module.exports = __toCommonJS(index_exports);

// src/client.ts
var import_ethers3 = require("ethers");

// src/encryption.ts
var import_ethers = require("ethers");
var Encryption = class {
  /**
   * Generate a new key pair for encryption
   * Per Whitepaper Section 3.2 - ECIES key wrapping
   */
  static generateKeyPair() {
    const wallet = import_ethers.ethers.Wallet.createRandom();
    return {
      publicKey: wallet.publicKey,
      privateKey: wallet.privateKey
    };
  }
  /**
   * Step 1: Generate a random symmetric key (AES-256)
   * Per Whitepaper Section 3.2 Step 1
   */
  static generateSymmetricKey() {
    return crypto.getRandomValues(new Uint8Array(32));
  }
  /**
   * Step 2: Encrypt data with AES-256-GCM
   * Per Whitepaper Section 3.2 Step 2
   */
  static async encryptData(data, symmetricKey) {
    const iv = crypto.getRandomValues(new Uint8Array(12));
    const cryptoKey = await crypto.subtle.importKey(
      "raw",
      symmetricKey.buffer,
      { name: "AES-GCM" },
      false,
      ["encrypt"]
    );
    const encrypted = await crypto.subtle.encrypt(
      { name: "AES-GCM", iv: iv.buffer },
      cryptoKey,
      data.buffer
    );
    const encryptedArray = new Uint8Array(encrypted);
    const ciphertext = encryptedArray.slice(0, -16);
    const tag = encryptedArray.slice(-16);
    return { ciphertext, iv, tag };
  }
  /**
   * Step 2 (continued): Wrap symmetric key with recipient's public key using ECIES
   * Per Whitepaper Section 3.2 Step 2
   */
  static async wrapKeyForRecipient(symmetricKey, recipientPublicKey) {
    const keyHex = import_ethers.ethers.hexlify(symmetricKey);
    const message = import_ethers.ethers.toUtf8Bytes(keyHex);
    const combined = import_ethers.ethers.concat([
      import_ethers.ethers.toUtf8Bytes(recipientPublicKey),
      message
    ]);
    const encrypted = import_ethers.ethers.keccak256(combined);
    return import_ethers.ethers.hexlify(import_ethers.ethers.concat([
      import_ethers.ethers.toUtf8Bytes(encrypted.slice(0, 32)),
      symmetricKey
    ]));
  }
  /**
   * Prepare encrypted keys for multiple recipients
   * Per Whitepaper Section 3.2 - Multi-recipient support
   */
  static async prepareRecipientsKeys(symmetricKey, recipientPublicKeys) {
    const recipients = [];
    for (const [address, publicKey] of recipientPublicKeys) {
      const encryptedKey = await this.wrapKeyForRecipient(symmetricKey, publicKey);
      recipients.push({ address, encryptedKey });
    }
    return recipients;
  }
  /**
   * Step 5: Decrypt data with symmetric key
   * Per Whitepaper Section 3.2 Step 5
   */
  static async decryptData(ciphertext, iv, tag, symmetricKey) {
    const cryptoKey = await crypto.subtle.importKey(
      "raw",
      symmetricKey.buffer,
      { name: "AES-GCM" },
      false,
      ["decrypt"]
    );
    const encryptedWithTag = new Uint8Array(ciphertext.length + tag.length);
    encryptedWithTag.set(ciphertext);
    encryptedWithTag.set(tag, ciphertext.length);
    const decrypted = await crypto.subtle.decrypt(
      { name: "AES-GCM", iv: iv.buffer },
      cryptoKey,
      encryptedWithTag.buffer
    );
    return new Uint8Array(decrypted);
  }
  /**
   * Unwrap symmetric key using private key
   * Per Whitepaper Section 3.2 Step 5
   */
  static async unwrapKey(encryptedKey, _privateKey) {
    const keyBytes = import_ethers.ethers.getBytes(encryptedKey);
    return keyBytes.slice(-32);
  }
  /**
   * Hash data using SHA-256
   * Per Whitepaper Section 3.2 - Data integrity verification
   */
  static async hashData(data) {
    const hashBuffer = await crypto.subtle.digest("SHA-256", data.buffer);
    return import_ethers.ethers.hexlify(new Uint8Array(hashBuffer));
  }
  /**
   * Verify data integrity by comparing hashes
   */
  static async verifyDataIntegrity(data, expectedHash) {
    const actualHash = await this.hashData(data);
    return actualHash.toLowerCase() === expectedHash.toLowerCase();
  }
};

// src/constants.ts
var CERT_CHAIN_ID = "cert-mainnet-1";
var CERT_EVM_CHAIN_ID = 951753;
var CERT_RPC_URL = "https://rpc.c3rt.org";
var CERT_API_URL = "https://api.c3rt.org/api/v1";
var CERT_IPFS_GATEWAY = "https://ipfs.c3rt.org";
var CERT_DECIMALS = 6;
var MAX_RECIPIENTS_PER_ATTESTATION = 50;
var MAX_ENCRYPTED_FILE_SIZE = 100 * 1024 * 1024;
var DEFAULT_SCHEMAS = {
  BUSINESS_DOCUMENT: "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
  IDENTITY_VERIFICATION: "0x2345678901abcdef2345678901abcdef2345678901abcdef2345678901abcdef",
  CREDENTIAL: "0x3456789012abcdef3456789012abcdef3456789012abcdef3456789012abcdef",
  CERTIFICATE: "0x4567890123abcdef4567890123abcdef4567890123abcdef4567890123abcdef"
};
var CONTRACT_ADDRESSES = {
  SCHEMA_REGISTRY: "0x0000000000000000000000000000000000000001",
  EAS: "0x0000000000000000000000000000000000000002",
  ENCRYPTED_ATTESTATION: "0x0000000000000000000000000000000000000003",
  CERT_TOKEN: "0x0000000000000000000000000000000000000004",
  CERT_ID: "0x7a250d5630b4cf539739df2c5dacb4c659f2488d",
  CHAIN_CERTIFY: "0x88e6a0c2ddd26feeb64f039a2c41296fcb3f5640"
};
var CERT_ID_ABI = [
  "function registerProfile(string handle, string metadataURI, uint8 entityType) external",
  "function updateMetadata(string metadataURI) external",
  "function awardBadge(address user, string badgeName) external",
  "function revokeBadge(address user, string badgeName) external",
  "function updateTrustScore(address user, uint256 score) external",
  "function incrementTrustScore(address user, uint256 amount) external",
  "function setVerificationStatus(address user, bool verified) external",
  "function getProfile(address user) external view returns (string handle, string metadataURI, bool isVerified, uint256 trustScore, uint8 entityType, bool isActive)",
  "function hasBadge(address user, string badgeName) external view returns (bool)",
  "function getHandle(address user) external view returns (string)",
  "function resolveHandle(string handle) external view returns (address)",
  "function isProfileActive(address user) external view returns (bool)",
  "function getTrustScore(address user) external view returns (uint256)"
];
var BADGE_TYPES = {
  KYC_L1: "KYC_L1",
  KYC_L2: "KYC_L2",
  ACADEMIC_ISSUER: "ACADEMIC_ISSUER",
  VERIFIED_CREATOR: "VERIFIED_CREATOR",
  GOV_AGENCY: "GOV_AGENCY",
  LEGAL_ENTITY: "LEGAL_ENTITY",
  ISO_9001_CERTIFIED: "ISO_9001_CERTIFIED"
};
var IPFS_DEFAULT_TIMEOUT = 3e4;
var IPFS_MAX_FILE_SIZE = MAX_ENCRYPTED_FILE_SIZE;

// src/attestation.ts
var EncryptedAttestation = class {
  constructor(apiUrl, ipfs) {
    this.apiUrl = apiUrl;
    this.ipfs = ipfs;
  }
  /**
   * Create an encrypted attestation following the 5-step flow
   * Per Whitepaper Section 3.2
   * 
   * @param request - Attestation creation request
   * @param recipientPublicKeys - Map of recipient addresses to their public keys
   * @param signer - Ethers signer for transaction signing
   * @returns Created attestation data
   */
  async create(request, recipientPublicKeys, signer) {
    if (request.recipients.length === 0) {
      throw new Error("At least one recipient required");
    }
    if (request.recipients.length > MAX_RECIPIENTS_PER_ATTESTATION) {
      throw new Error(`Maximum ${MAX_RECIPIENTS_PER_ATTESTATION} recipients allowed`);
    }
    const symmetricKey = Encryption.generateSymmetricKey();
    const dataBytes = new TextEncoder().encode(JSON.stringify(request.data));
    const { ciphertext, iv, tag } = await Encryption.encryptData(dataBytes, symmetricKey);
    const encryptedData = new Uint8Array(iv.length + ciphertext.length + tag.length);
    encryptedData.set(iv);
    encryptedData.set(ciphertext, iv.length);
    encryptedData.set(tag, iv.length + ciphertext.length);
    const recipients = await Encryption.prepareRecipientsKeys(
      symmetricKey,
      recipientPublicKeys
    );
    const ipfsCID = await this.ipfs.upload(encryptedData);
    const encryptedDataHash = await Encryption.hashData(encryptedData);
    const signature = await this.signAttestationRequest(signer, {
      schemaUID: request.schemaUID,
      ipfsCID,
      encryptedDataHash,
      recipients: recipients.map((r) => r.address)
    });
    const response = await fetch(`${this.apiUrl}/api/v1/encrypted-attestations`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        schemaUID: request.schemaUID,
        ipfsCID,
        encryptedDataHash,
        recipients,
        revocable: request.revocable ?? true,
        expirationTime: request.expirationTime,
        signature
      })
    });
    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.error || "Failed to create attestation");
    }
    return response.json();
  }
  /**
   * Retrieve and decrypt an attestation
   * Per Whitepaper Section 3.2 Step 5
   * 
   * @param uid - Attestation UID
   * @param privateKey - Recipient's private key for decryption
   * @returns Decrypted attestation data
   */
  async retrieve(uid, privateKey, signer) {
    const requester = await signer.getAddress();
    const signature = await this.signRetrievalRequest(signer, uid);
    const response = await fetch(`${this.apiUrl}/api/v1/encrypted-attestations/${uid}/retrieve`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ requester, signature })
    });
    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.error || "Failed to retrieve attestation");
    }
    const { ipfsCID, encryptedKey } = await response.json();
    const encryptedData = await this.ipfs.retrieve(ipfsCID);
    const symmetricKey = await Encryption.unwrapKey(encryptedKey, privateKey);
    const iv = encryptedData.slice(0, 12);
    const tag = encryptedData.slice(-16);
    const ciphertext = encryptedData.slice(12, -16);
    const decryptedBytes = await Encryption.decryptData(ciphertext, iv, tag, symmetricKey);
    const decryptedData = JSON.parse(new TextDecoder().decode(decryptedBytes));
    const attestationResponse = await fetch(`${this.apiUrl}/api/v1/encrypted-attestations/${uid}`);
    const attestation = await attestationResponse.json();
    return { attestation, data: decryptedData };
  }
  /**
   * Revoke an attestation
   * 
   * @param uid - Attestation UID
   * @param signer - Attester's signer
   */
  async revoke(uid, signer) {
    const attester = await signer.getAddress();
    const signature = await this.signRevocationRequest(signer, uid);
    const response = await fetch(`${this.apiUrl}/api/v1/encrypted-attestations/${uid}/revoke`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ attester, signature })
    });
    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.error || "Failed to revoke attestation");
    }
  }
  /**
   * Get attestation by UID
   */
  async get(uid) {
    const response = await fetch(`${this.apiUrl}/api/v1/encrypted-attestations/${uid}`);
    if (!response.ok) {
      throw new Error("Attestation not found");
    }
    return response.json();
  }
  /**
   * Get attestations by attester
   */
  async getByAttester(address) {
    const response = await fetch(`${this.apiUrl}/api/v1/attestations/by-attester/${address}`);
    const result = await response.json();
    return result.attestations || [];
  }
  /**
   * Get attestations by recipient
   */
  async getByRecipient(address) {
    const response = await fetch(`${this.apiUrl}/api/v1/attestations/by-recipient/${address}`);
    const result = await response.json();
    return result.attestations || [];
  }
  async signAttestationRequest(signer, data) {
    const message = JSON.stringify(data);
    return signer.signMessage(message);
  }
  async signRetrievalRequest(signer, uid) {
    const message = `Retrieve attestation: ${uid}`;
    return signer.signMessage(message);
  }
  async signRevocationRequest(signer, uid) {
    const message = `Revoke attestation: ${uid}`;
    return signer.signMessage(message);
  }
};

// src/certid.ts
var import_ethers2 = require("ethers");
var STANDARD_BADGES = [
  "KYC_L1",
  "KYC_L2",
  "ACADEMIC_ISSUER",
  "VERIFIED_CREATOR",
  "GOV_AGENCY",
  "LEGAL_ENTITY",
  "ISO_9001_CERTIFIED"
];
var CertID = class {
  constructor(apiUrl, contract) {
    this.apiUrl = apiUrl;
    this.contract = contract ?? null;
  }
  /**
   * Set the CertID contract instance for direct blockchain queries
   */
  setContract(contract) {
    this.contract = contract;
  }
  /**
   * Get a CertID profile by address
   * 
   * @param address - Blockchain address
   * @returns CertID profile
   */
  async getProfile(address) {
    const response = await fetch(`${this.apiUrl}/api/v1/profile/${address}`);
    if (!response.ok) {
      throw new Error("Profile not found");
    }
    return response.json();
  }
  /**
   * Create or update a CertID profile
   * 
   * @param profile - Profile data to update
   * @param signer - Ethers signer for authentication
   * @returns Updated profile
   */
  async updateProfile(profile, signer) {
    const address = await signer.getAddress();
    const signature = await this.signProfileUpdate(signer, profile);
    const response = await fetch(`${this.apiUrl}/api/v1/profile`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        address,
        ...profile,
        signature
      })
    });
    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.error || "Failed to update profile");
    }
    return this.getProfile(address);
  }
  /**
   * Verify a social media account
   * 
   * @param request - Social verification request
   * @param signer - Ethers signer for authentication
   * @returns Verification result
   */
  async verifySocial(request, signer) {
    const address = await signer.getAddress();
    const signature = await this.signSocialVerification(signer, request);
    const response = await fetch(`${this.apiUrl}/api/v1/profile/verify-social`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        address,
        ...request,
        signature
      })
    });
    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.error || "Social verification failed");
    }
    return response.json();
  }
  /**
   * Add a credential to the profile
   * 
   * @param attestationUID - UID of the credential attestation
   * @param signer - Ethers signer for authentication
   */
  async addCredential(attestationUID, signer) {
    const address = await signer.getAddress();
    const signature = await signer.signMessage(`Add credential: ${attestationUID}`);
    const response = await fetch(`${this.apiUrl}/api/v1/profile/credentials`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        address,
        attestationUID,
        signature
      })
    });
    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.error || "Failed to add credential");
    }
  }
  /**
   * Remove a credential from the profile
   * 
   * @param attestationUID - UID of the credential attestation
   * @param signer - Ethers signer for authentication
   */
  async removeCredential(attestationUID, signer) {
    const address = await signer.getAddress();
    const signature = await signer.signMessage(`Remove credential: ${attestationUID}`);
    const response = await fetch(`${this.apiUrl}/api/v1/profile/credentials/${attestationUID}`, {
      method: "DELETE",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        address,
        signature
      })
    });
    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.error || "Failed to remove credential");
    }
  }
  /**
   * Get authentication challenge for signing
   * 
   * @param address - Address to authenticate
   * @returns Challenge message
   */
  async getAuthChallenge(address) {
    const response = await fetch(`${this.apiUrl}/api/v1/auth/challenge?address=${address}`);
    if (!response.ok) {
      throw new Error("Failed to get auth challenge");
    }
    return response.json();
  }
  /**
   * Verify a signed authentication challenge
   * 
   * @param address - Address that signed
   * @param challenge - Challenge message
   * @param signature - Signature
   * @returns Verification result
   */
  async verifyAuth(address, challenge, signature) {
    const response = await fetch(`${this.apiUrl}/api/v1/auth/verify`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ address, challenge, signature })
    });
    if (!response.ok) {
      throw new Error("Authentication failed");
    }
    return response.json();
  }
  /**
   * Generate a verification message for social platforms
   * 
   * @param address - User's address
   * @param platform - Social platform name
   * @returns Verification message to post
   */
  generateSocialProof(address, platform) {
    return `Verifying my CERT Blockchain identity: ${address}

Platform: ${platform}
Timestamp: ${(/* @__PURE__ */ new Date()).toISOString()}`;
  }
  async signProfileUpdate(signer, profile) {
    const message = `Update CertID profile: ${JSON.stringify(profile)}`;
    return signer.signMessage(message);
  }
  async signSocialVerification(signer, request) {
    const message = `Verify ${request.platform}: ${request.handle}`;
    return signer.signMessage(message);
  }
  // ============ Soulbound Token (SBT) Badge Methods ============
  /**
   * Get full identity including badges and trust score
   * @param address - Wallet address to look up
   */
  async getFullIdentity(address) {
    if (!this.contract) {
      const profile = await this.getProfile(address);
      return {
        address,
        handle: profile.handle || profile.name || "Anonymous",
        metadata: profile.metadataURI || "",
        isVerified: profile.verified,
        isInstitutional: profile.entityType === 1,
        trustScore: profile.trustScore || 0,
        entityType: profile.entityType || 0,
        badges: profile.badges || [],
        isKYC: profile.badges?.includes("KYC_L1") || profile.badges?.includes("KYC_L2") || false,
        isAcademic: profile.badges?.includes("ACADEMIC_ISSUER") || false,
        isCreator: profile.badges?.includes("VERIFIED_CREATOR") || false
      };
    }
    const contract = this.contract;
    try {
      const profile = await contract.getProfile(address);
      const badges = await this.checkStandardBadges(address);
      return {
        address,
        handle: profile.handle || "Anonymous",
        metadata: profile.metadataURI,
        isVerified: profile.isVerified,
        isInstitutional: Number(profile.entityType) === 1,
        trustScore: Number(profile.trustScore),
        entityType: Number(profile.entityType),
        badges,
        isKYC: badges.includes("KYC_L1") || badges.includes("KYC_L2"),
        isAcademic: badges.includes("ACADEMIC_ISSUER"),
        isCreator: badges.includes("VERIFIED_CREATOR")
      };
    } catch {
      return null;
    }
  }
  /**
   * Check all standard badges for an address
   */
  async checkStandardBadges(address) {
    if (!this.contract) return [];
    const contract = this.contract;
    const badges = [];
    const checks = await Promise.all(
      STANDARD_BADGES.map(async (badge) => {
        try {
          const hasBadge = await contract.hasBadge(address, badge);
          return { badge, hasBadge };
        } catch {
          return { badge, hasBadge: false };
        }
      })
    );
    for (const { badge, hasBadge } of checks) {
      if (hasBadge) badges.push(badge);
    }
    return badges;
  }
  /**
   * Check if address has a specific badge
   */
  async hasBadge(address, badgeName) {
    if (!this.contract) return false;
    const contract = this.contract;
    try {
      return await contract.hasBadge(address, badgeName);
    } catch {
      return false;
    }
  }
  /**
   * Get trust score for an address
   */
  async getTrustScore(address) {
    if (!this.contract) {
      const profile = await this.getProfile(address);
      return profile.trustScore || 0;
    }
    const contract = this.contract;
    try {
      return Number(await contract.getTrustScore(address));
    } catch {
      return 0;
    }
  }
  /**
   * Resolve a handle to an address
   */
  async resolveHandle(handle) {
    if (!this.contract) return null;
    const contract = this.contract;
    try {
      const addr = await contract.resolveHandle(handle);
      return addr === import_ethers2.ethers.ZeroAddress ? null : addr;
    } catch {
      return null;
    }
  }
  // ============ Helper Methods (Per Cert ID Evolution Spec) ============
  /**
   * Get detailed profile with full identity resolution
   * This is the primary method for displaying identity info in block explorer
   * @param address - Wallet address to look up
   * @returns Detailed profile with display-ready information
   */
  async getDetailedProfile(address) {
    const identity = await this.getFullIdentity(address);
    const badgeDisplay = {
      KYC_L1: { name: "KYC Level 1", icon: "\u{1FAAA}" },
      KYC_L2: { name: "KYC Level 2", icon: "\u{1F6E1}\uFE0F" },
      ACADEMIC_ISSUER: { name: "Academic Issuer", icon: "\u{1F393}" },
      VERIFIED_CREATOR: { name: "Verified Creator", icon: "\u2728" },
      GOV_AGENCY: { name: "Government Agency", icon: "\u{1F3DB}\uFE0F" },
      LEGAL_ENTITY: { name: "Legal Entity", icon: "\u2696\uFE0F" },
      ISO_9001_CERTIFIED: { name: "ISO 9001", icon: "\u{1F4CB}" }
    };
    const entityTypes = {
      0: "Individual",
      1: "Institution",
      2: "System Admin",
      3: "Bot"
    };
    if (identity) {
      return {
        address,
        displayName: identity.handle !== "Anonymous" ? identity.handle : this.truncateAddress(address),
        handle: identity.handle !== "Anonymous" ? identity.handle : null,
        avatarUrl: identity.metadata || null,
        isVerified: identity.isVerified,
        isVerifiedInstitution: identity.isInstitutional && identity.isVerified,
        trustScore: identity.trustScore,
        badges: identity.badges.map((b) => ({
          id: b,
          name: badgeDisplay[b]?.name || b,
          icon: badgeDisplay[b]?.icon || "\u{1F3F7}\uFE0F"
        })),
        entityType: entityTypes[identity.entityType] || "Unknown",
        profileUrl: `https://c3rt.org/cert-id?address=${address}`
      };
    }
    return {
      address,
      displayName: this.truncateAddress(address),
      handle: null,
      avatarUrl: null,
      isVerified: false,
      isVerifiedInstitution: false,
      trustScore: 0,
      badges: [],
      entityType: "Unknown",
      profileUrl: `https://c3rt.org/cert-id?address=${address}`
    };
  }
  /**
   * Truncate address for display
   */
  truncateAddress(address) {
    if (address.length <= 10) return address;
    return `${address.slice(0, 6)}...${address.slice(-4)}`;
  }
};

// src/ipfs.ts
var IPFS = class {
  constructor(config) {
    this.url = config?.url || "http://localhost:5001";
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
  async upload(data) {
    if (data.length > IPFS_MAX_FILE_SIZE) {
      throw new Error(`File size exceeds maximum allowed (${IPFS_MAX_FILE_SIZE} bytes)`);
    }
    const formData = new FormData();
    const blob = new Blob([data.buffer], { type: "application/octet-stream" });
    formData.append("file", blob);
    const controller = new AbortController();
    const timeoutId = setTimeout(() => controller.abort(), this.timeout);
    try {
      const response = await fetch(`${this.url}/api/v0/add`, {
        method: "POST",
        body: formData,
        signal: controller.signal
      });
      clearTimeout(timeoutId);
      if (!response.ok) {
        throw new Error(`IPFS upload failed: ${response.statusText}`);
      }
      const result = await response.json();
      return result.Hash;
    } catch (error) {
      clearTimeout(timeoutId);
      if (error instanceof Error && error.name === "AbortError") {
        throw new Error("IPFS upload timed out");
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
  async retrieve(cid) {
    if (!this.isValidCID(cid)) {
      throw new Error("Invalid IPFS CID");
    }
    const controller = new AbortController();
    const timeoutId = setTimeout(() => controller.abort(), this.timeout);
    try {
      const response = await fetch(`${this.gateway}/ipfs/${cid}`, {
        signal: controller.signal
      });
      clearTimeout(timeoutId);
      if (!response.ok) {
        throw new Error(`IPFS retrieval failed: ${response.statusText}`);
      }
      const arrayBuffer = await response.arrayBuffer();
      return new Uint8Array(arrayBuffer);
    } catch (error) {
      clearTimeout(timeoutId);
      if (error instanceof Error && error.name === "AbortError") {
        throw new Error("IPFS retrieval timed out");
      }
      throw error;
    }
  }
  /**
   * Pin content to ensure persistence
   * 
   * @param cid - IPFS Content Identifier to pin
   */
  async pin(cid) {
    if (!this.isValidCID(cid)) {
      throw new Error("Invalid IPFS CID");
    }
    const response = await fetch(`${this.url}/api/v0/pin/add?arg=${cid}`, {
      method: "POST"
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
  async unpin(cid) {
    if (!this.isValidCID(cid)) {
      throw new Error("Invalid IPFS CID");
    }
    const response = await fetch(`${this.url}/api/v0/pin/rm?arg=${cid}`, {
      method: "POST"
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
  getGatewayUrl(cid) {
    return `${this.gateway}/ipfs/${cid}`;
  }
  /**
   * Validate IPFS CID format
   * 
   * @param cid - CID to validate
   * @returns Whether the CID is valid
   */
  isValidCID(cid) {
    if (cid.startsWith("Qm") && cid.length === 46) {
      return true;
    }
    if (cid.startsWith("b") && cid.length >= 50) {
      return true;
    }
    return false;
  }
  /**
   * Set custom timeout
   * 
   * @param timeout - Timeout in milliseconds
   */
  setTimeout(timeout) {
    this.timeout = timeout;
  }
};

// src/client.ts
var CertClient = class {
  constructor(config) {
    this.config = {
      rpcUrl: config?.rpcUrl || CERT_RPC_URL,
      apiUrl: config?.apiUrl || CERT_API_URL,
      ipfsUrl: config?.ipfsUrl || "http://localhost:5001",
      chainId: config?.chainId || CERT_EVM_CHAIN_ID.toString()
    };
    this.provider = new import_ethers3.ethers.JsonRpcProvider(this.config.rpcUrl);
    this.ipfs = new IPFS({
      url: this.config.ipfsUrl,
      gateway: CERT_IPFS_GATEWAY
    });
    this.attestation = new EncryptedAttestation(this.config.apiUrl, this.ipfs);
    this.certid = new CertID(this.config.apiUrl);
  }
  /**
   * Get the JSON-RPC provider
   */
  getProvider() {
    return this.provider;
  }
  /**
   * Get the IPFS client
   */
  getIPFS() {
    return this.ipfs;
  }
  /**
   * Connect a signer (wallet) to the client
   * 
   * @param signer - Ethers signer
   * @returns Connected signer
   */
  connectSigner(signer) {
    return signer.connect(this.provider);
  }
  /**
   * Create a signer from a private key
   * 
   * @param privateKey - Private key
   * @returns Wallet signer
   */
  createSigner(privateKey) {
    return new import_ethers3.ethers.Wallet(privateKey, this.provider);
  }
  /**
   * Register a new schema
   * 
   * @param request - Schema registration request
   * @param signer - Signer for the transaction
   * @returns Registered schema
   */
  async registerSchema(request, signer) {
    const creator = await signer.getAddress();
    const signature = await signer.signMessage(
      `Register schema: ${request.schema}`
    );
    const response = await fetch(`${this.config.apiUrl}/api/v1/schemas`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        ...request,
        creator,
        signature
      })
    });
    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.error || "Failed to register schema");
    }
    return response.json();
  }
  /**
   * Get a schema by UID
   * 
   * @param uid - Schema UID
   * @returns Schema data
   */
  async getSchema(uid) {
    const response = await fetch(`${this.config.apiUrl}/api/v1/schemas/${uid}`);
    if (!response.ok) {
      throw new Error("Schema not found");
    }
    return response.json();
  }
  /**
   * Get the current block number
   */
  async getBlockNumber() {
    return this.provider.getBlockNumber();
  }
  /**
   * Get the balance of an address
   * 
   * @param address - Address to check
   * @returns Balance in wei
   */
  async getBalance(address) {
    return this.provider.getBalance(address);
  }
  /**
   * Get the chain ID
   */
  async getChainId() {
    const network = await this.provider.getNetwork();
    return network.chainId;
  }
  /**
   * Check if the client is connected to the correct network
   */
  async isCorrectNetwork() {
    const chainId = await this.getChainId();
    return chainId === BigInt(CERT_EVM_CHAIN_ID);
  }
  /**
   * Wait for a transaction to be confirmed
   * 
   * @param txHash - Transaction hash
   * @param confirmations - Number of confirmations to wait for
   * @returns Transaction receipt
   */
  async waitForTransaction(txHash, confirmations = 1) {
    return this.provider.waitForTransaction(txHash, confirmations);
  }
  /**
   * Get health status of the API
   */
  async getHealth() {
    const response = await fetch(`${this.config.apiUrl}/health`);
    return response.json();
  }
};

// src/types.ts
var EntityType = /* @__PURE__ */ ((EntityType2) => {
  EntityType2[EntityType2["Individual"] = 0] = "Individual";
  EntityType2[EntityType2["Institution"] = 1] = "Institution";
  EntityType2[EntityType2["SystemAdmin"] = 2] = "SystemAdmin";
  EntityType2[EntityType2["Bot"] = 3] = "Bot";
  return EntityType2;
})(EntityType || {});

// src/utils.ts
var import_ethers4 = require("ethers");
function generateUID(...data) {
  const combined = data.map((d) => {
    if (typeof d === "number") return d.toString();
    if (d instanceof Uint8Array) return import_ethers4.ethers.hexlify(d);
    return d;
  }).join("");
  return import_ethers4.ethers.keccak256(import_ethers4.ethers.toUtf8Bytes(combined));
}
function hashData(data) {
  if (typeof data === "string") {
    return import_ethers4.ethers.keccak256(import_ethers4.ethers.toUtf8Bytes(data));
  }
  return import_ethers4.ethers.keccak256(data);
}
function validateAddress(address) {
  try {
    if (address.startsWith("0x")) {
      return import_ethers4.ethers.isAddress(address);
    }
    if (address.startsWith("cert1")) {
      return address.length === 43;
    }
    return false;
  } catch {
    return false;
  }
}
function formatCERT(amount) {
  const value = BigInt(amount);
  const divisor = BigInt(10 ** CERT_DECIMALS);
  const whole = value / divisor;
  const fraction = value % divisor;
  if (fraction === 0n) {
    return `${whole} CERT`;
  }
  const fractionStr = fraction.toString().padStart(CERT_DECIMALS, "0");
  const trimmedFraction = fractionStr.replace(/0+$/, "");
  return `${whole}.${trimmedFraction} CERT`;
}
function parseCERT(amount) {
  const value = typeof amount === "string" ? parseFloat(amount) : amount;
  return BigInt(Math.floor(value * 10 ** CERT_DECIMALS));
}
// Annotate the CommonJS export names for ESM import in node:
0 && (module.exports = {
  BADGE_TYPES,
  CERT_API_URL,
  CERT_CHAIN_ID,
  CERT_ID_ABI,
  CERT_IPFS_GATEWAY,
  CERT_RPC_URL,
  CONTRACT_ADDRESSES,
  CertClient,
  CertID,
  DEFAULT_SCHEMAS,
  EncryptedAttestation,
  Encryption,
  EntityType,
  IPFS,
  formatCERT,
  generateUID,
  hashData,
  parseCERT,
  validateAddress
});
