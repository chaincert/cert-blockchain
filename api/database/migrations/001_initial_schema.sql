-- CERT Blockchain CertID Database Schema
-- Per CertID Section 2.2: user_profiles table

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- User profiles table
-- Primary table for CertID decentralized identity
CREATE TABLE IF NOT EXISTS user_profiles (
    -- Address is the primary key (Cosmos/EVM address)
    address VARCHAR(64) PRIMARY KEY,
    
    -- Basic profile information
    name VARCHAR(100),
    bio TEXT CHECK (char_length(bio) <= 500),
    avatar_url VARCHAR(512),
    
    -- Social links stored as JSONB for flexibility
    -- Format: {"twitter": "@handle", "github": "username", ...}
    social_links JSONB DEFAULT '{}',
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create index on address for fast lookups
CREATE INDEX IF NOT EXISTS idx_user_profiles_address ON user_profiles(address);

-- Credentials table
-- Links attestations to user profiles
CREATE TABLE IF NOT EXISTS credentials (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    
    -- Foreign key to user profile
    user_address VARCHAR(64) NOT NULL REFERENCES user_profiles(address) ON DELETE CASCADE,
    
    -- Credential details
    credential_type VARCHAR(50) NOT NULL,
    attestation_uid VARCHAR(66) NOT NULL,
    issuer VARCHAR(64) NOT NULL,
    
    -- Verification status
    verified BOOLEAN DEFAULT false,
    
    -- Timestamps
    issued_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for credentials
CREATE INDEX IF NOT EXISTS idx_credentials_user_address ON credentials(user_address);
CREATE INDEX IF NOT EXISTS idx_credentials_attestation_uid ON credentials(attestation_uid);

-- Social verifications table
-- Tracks verified social accounts
CREATE TABLE IF NOT EXISTS social_verifications (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    
    -- Foreign key to user profile
    user_address VARCHAR(64) NOT NULL REFERENCES user_profiles(address) ON DELETE CASCADE,
    
    -- Social platform details
    platform VARCHAR(50) NOT NULL,
    handle VARCHAR(100) NOT NULL,
    proof_url VARCHAR(512),
    
    -- Verification status
    verified BOOLEAN DEFAULT false,
    verified_at TIMESTAMP WITH TIME ZONE,
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    -- Unique constraint: one verification per platform per user
    UNIQUE(user_address, platform)
);

-- Create index for social verifications
CREATE INDEX IF NOT EXISTS idx_social_verifications_user_address ON social_verifications(user_address);

-- Attestation cache table
-- Caches frequently accessed attestation metadata
CREATE TABLE IF NOT EXISTS attestation_cache (
    uid VARCHAR(66) PRIMARY KEY,
    schema_uid VARCHAR(66) NOT NULL,
    attester VARCHAR(64) NOT NULL,
    recipient VARCHAR(64),
    
    -- Attestation metadata
    data_hash VARCHAR(66),
    ipfs_cid VARCHAR(100),
    is_encrypted BOOLEAN DEFAULT false,
    revocable BOOLEAN DEFAULT true,
    revoked BOOLEAN DEFAULT false,
    
    -- Timestamps
    attestation_time TIMESTAMP WITH TIME ZONE NOT NULL,
    expiration_time TIMESTAMP WITH TIME ZONE,
    revocation_time TIMESTAMP WITH TIME ZONE,
    
    -- Cache metadata
    cached_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for attestation cache
CREATE INDEX IF NOT EXISTS idx_attestation_cache_attester ON attestation_cache(attester);
CREATE INDEX IF NOT EXISTS idx_attestation_cache_recipient ON attestation_cache(recipient);
CREATE INDEX IF NOT EXISTS idx_attestation_cache_schema ON attestation_cache(schema_uid);

-- Function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Trigger for user_profiles updated_at
CREATE TRIGGER update_user_profiles_updated_at
    BEFORE UPDATE ON user_profiles
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

