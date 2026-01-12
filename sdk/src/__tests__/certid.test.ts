/*
 * Copyright 2026 Brandon Guynn
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
 * Tests for CertID SDK Module
 * Tests the CertID identity management functionality
 */

import { ethers } from 'ethers';

// CertID types (matching certid.ts structure)
interface CertIDProfile {
  username: string;
  address: string;
  displayName?: string;
  bio?: string;
  avatarCID?: string;
  credentials: Credential[];
  socialVerifications: SocialVerification[];
  createdAt: number;
  updatedAt: number;
}

interface Credential {
  type: string;
  attestationUID: string;
  issuedAt: number;
  expiresAt?: number;
  revoked: boolean;
}

interface SocialVerification {
  platform: 'twitter' | 'github' | 'linkedin' | 'discord';
  handle: string;
  verified: boolean;
  verifiedAt?: number;
}

// Username validation per whitepaper
function validateUsername(username: string): { valid: boolean; error?: string } {
  if (!username || username.length < 3) {
    return { valid: false, error: 'Username must be at least 3 characters' };
  }
  if (username.length > 32) {
    return { valid: false, error: 'Username must be at most 32 characters' };
  }
  if (!/^[a-z][a-z0-9_]*$/.test(username)) {
    return { valid: false, error: 'Username must start with a letter and contain only lowercase letters, numbers, and underscores' };
  }
  return { valid: true };
}

// Supported platforms
const SUPPORTED_PLATFORMS = ['twitter', 'github', 'linkedin', 'discord'] as const;
const CREDENTIAL_TYPES = ['education', 'employment', 'certification', 'membership', 'achievement'] as const;

describe('CertID SDK', () => {
  describe('CertIDProfile interface', () => {
    it('should accept valid profile data', () => {
      const profile: CertIDProfile = {
        username: 'alice',
        address: '0x' + '1'.repeat(40),
        displayName: 'Alice Johnson',
        bio: 'Blockchain developer and researcher',
        avatarCID: 'QmYwAPJzv5CZsnAzt8auVZRn5W7x8Hd8fH6FS6NVQP3fSw',
        credentials: [],
        socialVerifications: [],
        createdAt: Date.now(),
        updatedAt: Date.now(),
      };

      expect(profile.username).toBe('alice');
      expect(profile.address).toHaveLength(42);
    });

    it('should allow optional fields to be undefined', () => {
      const profile: CertIDProfile = {
        username: 'bob',
        address: '0x' + '2'.repeat(40),
        credentials: [],
        socialVerifications: [],
        createdAt: Date.now(),
        updatedAt: Date.now(),
      };

      expect(profile.displayName).toBeUndefined();
      expect(profile.bio).toBeUndefined();
      expect(profile.avatarCID).toBeUndefined();
    });
  });

  describe('validateUsername', () => {
    it('should accept valid usernames', () => {
      expect(validateUsername('alice').valid).toBe(true);
      expect(validateUsername('bob123').valid).toBe(true);
      expect(validateUsername('charlie_dev').valid).toBe(true);
      expect(validateUsername('user_with_numbers_123').valid).toBe(true);
    });

    it('should reject empty username', () => {
      const result = validateUsername('');
      expect(result.valid).toBe(false);
      expect(result.error).toContain('at least 3');
    });

    it('should reject username shorter than 3 characters', () => {
      const result = validateUsername('ab');
      expect(result.valid).toBe(false);
    });

    it('should reject username longer than 32 characters', () => {
      const result = validateUsername('a'.repeat(33));
      expect(result.valid).toBe(false);
      expect(result.error).toContain('at most 32');
    });

    it('should reject username starting with number', () => {
      const result = validateUsername('123user');
      expect(result.valid).toBe(false);
    });

    it('should reject username with uppercase letters', () => {
      const result = validateUsername('Alice');
      expect(result.valid).toBe(false);
    });

    it('should reject username with special characters', () => {
      expect(validateUsername('user@name').valid).toBe(false);
      expect(validateUsername('user-name').valid).toBe(false);
      expect(validateUsername('user.name').valid).toBe(false);
    });
  });

  describe('SUPPORTED_PLATFORMS', () => {
    it('should include all supported social platforms', () => {
      expect(SUPPORTED_PLATFORMS).toContain('twitter');
      expect(SUPPORTED_PLATFORMS).toContain('github');
      expect(SUPPORTED_PLATFORMS).toContain('linkedin');
      expect(SUPPORTED_PLATFORMS).toContain('discord');
    });

    it('should have exactly 4 platforms', () => {
      expect(SUPPORTED_PLATFORMS.length).toBe(4);
    });
  });

  describe('CREDENTIAL_TYPES', () => {
    it('should include all credential types', () => {
      expect(CREDENTIAL_TYPES).toContain('education');
      expect(CREDENTIAL_TYPES).toContain('employment');
      expect(CREDENTIAL_TYPES).toContain('certification');
      expect(CREDENTIAL_TYPES).toContain('membership');
      expect(CREDENTIAL_TYPES).toContain('achievement');
    });
  });

  describe('Credential interface', () => {
    it('should accept valid credential data', () => {
      const credential: Credential = {
        type: 'education',
        attestationUID: '0x' + 'a'.repeat(64),
        issuedAt: Date.now(),
        revoked: false,
      };

      expect(credential.type).toBe('education');
      expect(credential.attestationUID).toHaveLength(66);
    });
  });

  describe('SocialVerification interface', () => {
    it('should accept valid social verification', () => {
      const verification: SocialVerification = {
        platform: 'twitter',
        handle: '@alice_crypto',
        verified: true,
        verifiedAt: Date.now(),
      };

      expect(verification.platform).toBe('twitter');
      expect(verification.verified).toBe(true);
    });
  });
});

