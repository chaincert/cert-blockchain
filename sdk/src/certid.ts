/**
 * CertID module for CERT Blockchain SDK
 * Implements decentralized identity per Whitepaper CertID Section
 */

import { ethers } from 'ethers';
import type {
  CertIDProfile,
  UpdateProfileRequest,
  VerifySocialRequest,
} from './types';

export class CertID {
  private apiUrl: string;

  constructor(apiUrl: string) {
    this.apiUrl = apiUrl;
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
}

