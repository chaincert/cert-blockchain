/*
 * Copyright 2026 Cert Blockchain LLC
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */


/**
 * CertID module for CERT Blockchain SDK
 * Implements decentralized identity per Whitepaper CertID Section
 * Includes Soulbound Token (SBT) badge support and trust scores
 */

import { ethers } from 'ethers';
import type {
  CertIDProfile,
  UpdateProfileRequest,
  VerifySocialRequest,
  FullIdentity,
  BadgeType,
  EntityType,
} from './types';
import { CERT_ID_ABI, CONTRACT_ADDRESSES } from './constants';

// Standard badges to check
const STANDARD_BADGES = [
  'KYC_L1',
  'KYC_L2',
  'ACADEMIC_ISSUER',
  'VERIFIED_CREATOR',
  'GOV_AGENCY',
  'LEGAL_ENTITY',
  'ISO_9001_CERTIFIED',
];

export class CertID {
  private apiUrl: string;
  private contract: ethers.Contract | null;

  constructor(apiUrl: string, contractOrSigner?: ethers.Contract | ethers.Signer | ethers.Provider, contractAddress?: string) {
    this.apiUrl = apiUrl;

    if (contractOrSigner) {
      if (contractOrSigner instanceof ethers.Contract) {
        this.contract = contractOrSigner;
      } else {
        const address = contractAddress || CONTRACT_ADDRESSES.CERT_ID;
        this.contract = new ethers.Contract(address, CERT_ID_ABI, contractOrSigner);
      }
    } else {
      this.contract = null;
    }
  }

  /**
   * Set the CertID contract instance for direct blockchain queries
   */
  setContract(contract: ethers.Contract): void {
    this.contract = contract;
  }

  /**
   * Get a CertID profile by address
   * 
   * @param address - Blockchain address
   * @returns CertID profile
   */
  async getProfile(address: string): Promise<CertIDProfile> {
    const response = await fetch(`${this.apiUrl}/api/v1/profile/${address}`);
    if (!response.ok) {
      throw new Error('Profile not found');
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
  async updateProfile(
    profile: UpdateProfileRequest,
    signer: ethers.Signer
  ): Promise<CertIDProfile> {
    const address = await signer.getAddress();
    const signature = await this.signProfileUpdate(signer, profile);

    const response = await fetch(`${this.apiUrl}/api/v1/profile`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        address,
        ...profile,
        signature,
      }),
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.error || 'Failed to update profile');
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
  async verifySocial(
    request: VerifySocialRequest,
    signer: ethers.Signer
  ): Promise<{ verified: boolean; platform: string }> {
    const address = await signer.getAddress();
    const signature = await this.signSocialVerification(signer, request);

    const response = await fetch(`${this.apiUrl}/api/v1/profile/verify-social`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        address,
        ...request,
        signature,
      }),
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.error || 'Social verification failed');
    }

    return response.json();
  }

  /**
   * Add a credential to the profile
   * 
   * @param attestationUID - UID of the credential attestation
   * @param signer - Ethers signer for authentication
   */
  async addCredential(attestationUID: string, signer: ethers.Signer): Promise<void> {
    const address = await signer.getAddress();
    const signature = await signer.signMessage(`Add credential: ${attestationUID}`);

    const response = await fetch(`${this.apiUrl}/api/v1/profile/credentials`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        address,
        attestationUID,
        signature,
      }),
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.error || 'Failed to add credential');
    }
  }

  /**
   * Remove a credential from the profile
   * 
   * @param attestationUID - UID of the credential attestation
   * @param signer - Ethers signer for authentication
   */
  async removeCredential(attestationUID: string, signer: ethers.Signer): Promise<void> {
    const address = await signer.getAddress();
    const signature = await signer.signMessage(`Remove credential: ${attestationUID}`);

    const response = await fetch(`${this.apiUrl}/api/v1/profile/credentials/${attestationUID}`, {
      method: 'DELETE',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        address,
        signature,
      }),
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.error || 'Failed to remove credential');
    }
  }

  /**
   * Get authentication challenge for signing
   * 
   * @param address - Address to authenticate
   * @returns Challenge message
   */
  async getAuthChallenge(address: string): Promise<{ challenge: string; expiresAt: string }> {
    const response = await fetch(`${this.apiUrl}/api/v1/auth/challenge?address=${address}`);
    if (!response.ok) {
      throw new Error('Failed to get auth challenge');
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
  async verifyAuth(
    address: string,
    challenge: string,
    signature: string
  ): Promise<{ verified: boolean }> {
    const response = await fetch(`${this.apiUrl}/api/v1/auth/verify`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ address, challenge, signature }),
    });

    if (!response.ok) {
      throw new Error('Authentication failed');
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
  generateSocialProof(address: string, platform: string): string {
    return `Verifying my CERT Blockchain identity: ${address}\n\nPlatform: ${platform}\nTimestamp: ${new Date().toISOString()}`;
  }

  private async signProfileUpdate(
    signer: ethers.Signer,
    profile: UpdateProfileRequest
  ): Promise<string> {
    const message = `Update CertID profile: ${JSON.stringify(profile)}`;
    return signer.signMessage(message);
  }

  private async signSocialVerification(
    signer: ethers.Signer,
    request: VerifySocialRequest
  ): Promise<string> {
    const message = `Verify ${request.platform}: ${request.handle}`;
    return signer.signMessage(message);
  }

  // ============ Soulbound Token (SBT) Badge Methods ============

  /**
   * Get full identity including badges and trust score
   * @param address - Wallet address to look up
   */
  async getFullIdentity(address: string): Promise<FullIdentity | null> {
    if (!this.contract) {
      // Fallback to API if no contract
      const profile = await this.getProfile(address);
      return {
        address,
        handle: profile.handle || profile.name || 'Anonymous',
        metadata: profile.metadataURI || '',
        isVerified: profile.verified,
        isInstitutional: profile.entityType === 1,
        trustScore: profile.trustScore || 0,
        entityType: profile.entityType || 0,
        badges: profile.badges || [],
        isKYC: profile.badges?.includes('KYC_L1') || profile.badges?.includes('KYC_L2') || false,
        isAcademic: profile.badges?.includes('ACADEMIC_ISSUER') || false,
        isCreator: profile.badges?.includes('VERIFIED_CREATOR') || false,
      };
    }

    const contract = this.contract;
    try {
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      const profile = await (contract as any).getProfile(address);
      const badges = await this.checkStandardBadges(address);

      return {
        address,
        handle: profile.handle || 'Anonymous',
        metadata: profile.metadataURI,
        isVerified: profile.isVerified,
        isInstitutional: Number(profile.entityType) === 1,
        trustScore: Number(profile.trustScore),
        entityType: Number(profile.entityType) as EntityType,
        badges,
        isKYC: badges.includes('KYC_L1') || badges.includes('KYC_L2'),
        isAcademic: badges.includes('ACADEMIC_ISSUER'),
        isCreator: badges.includes('VERIFIED_CREATOR'),
      };
    } catch {
      return null;
    }
  }

  /**
   * Check all standard badges for an address
   */
  async checkStandardBadges(address: string): Promise<string[]> {
    if (!this.contract) return [];

    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const contract = this.contract as any;
    const badges: string[] = [];
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
  async hasBadge(address: string, badgeName: BadgeType): Promise<boolean> {
    if (!this.contract) return false;
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const contract = this.contract as any;
    try {
      return await contract.hasBadge(address, badgeName);
    } catch {
      return false;
    }
  }

  /**
   * Get trust score for an address
   */
  async getTrustScore(address: string): Promise<number> {
    if (!this.contract) {
      const profile = await this.getProfile(address);
      return profile.trustScore || 0;
    }
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const contract = this.contract as any;
    try {
      return Number(await contract.getTrustScore(address));
    } catch {
      return 0;
    }
  }

  /**
   * Resolve a handle to an address
   */
  async resolveHandle(handle: string): Promise<string | null> {
    if (!this.contract) return null;
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const contract = this.contract as any;
    try {
      const addr = await contract.resolveHandle(handle);
      return addr === ethers.ZeroAddress ? null : addr;
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
  async getDetailedProfile(address: string): Promise<{
    address: string;
    displayName: string;
    handle: string | null;
    avatarUrl: string | null;
    isVerified: boolean;
    isVerifiedInstitution: boolean;
    trustScore: number;
    badges: Array<{ id: string; name: string; icon: string }>;
    entityType: string;
    profileUrl: string;
  }> {
    // Try to get full identity from contract/API
    const identity = await this.getFullIdentity(address);

    // Badge display mapping
    const badgeDisplay: Record<string, { name: string; icon: string }> = {
      KYC_L1: { name: 'KYC Level 1', icon: '🪪' },
      KYC_L2: { name: 'KYC Level 2', icon: '🛡️' },
      ACADEMIC_ISSUER: { name: 'Academic Issuer', icon: '🎓' },
      VERIFIED_CREATOR: { name: 'Verified Creator', icon: '✨' },
      GOV_AGENCY: { name: 'Government Agency', icon: '🏛️' },
      LEGAL_ENTITY: { name: 'Legal Entity', icon: '⚖️' },
      ISO_9001_CERTIFIED: { name: 'ISO 9001', icon: '📋' },
    };

    // Entity type mapping
    const entityTypes: Record<number, string> = {
      0: 'Individual',
      1: 'Institution',
      2: 'System Admin',
      3: 'Bot',
    };

    if (identity) {
      return {
        address,
        displayName: identity.handle !== 'Anonymous' ? identity.handle : this.truncateAddress(address),
        handle: identity.handle !== 'Anonymous' ? identity.handle : null,
        avatarUrl: identity.metadata || null,
        isVerified: identity.isVerified,
        isVerifiedInstitution: identity.isInstitutional && identity.isVerified,
        trustScore: identity.trustScore,
        badges: identity.badges.map(b => ({
          id: b,
          name: badgeDisplay[b]?.name || b,
          icon: badgeDisplay[b]?.icon || '🏷️',
        })),
        entityType: entityTypes[identity.entityType] || 'Unknown',
        profileUrl: `https://c3rt.org/cert-id?address=${address}`,
      };
    }

    // Fallback for addresses without CertID
    return {
      address,
      displayName: this.truncateAddress(address),
      handle: null,
      avatarUrl: null,
      isVerified: false,
      isVerifiedInstitution: false,
      trustScore: 0,
      badges: [],
      entityType: 'Unknown',
      profileUrl: `https://c3rt.org/cert-id?address=${address}`,
    };
  }

  /**
   * Truncate address for display
   */
  private truncateAddress(address: string): string {
    if (address.length <= 10) return address;
    return `${address.slice(0, 6)}...${address.slice(-4)}`;
  }
}

