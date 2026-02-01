/**
 * CertID Sybil Resistance SDK
 * Official JavaScript client for CertID Trust Score API
 * 
 * @package @certid/sybil-sdk
 * @version 1.0.0
 * @license Apache-2.0
 */

export interface TrustScoreResult {
    address: string;
    trustScore: number;
    isLikelyHuman: boolean;
    factors?: {
        kycVerified?: boolean;
        socialVerifications?: number;
        onChainActivity?: number;
        crossChainPresence?: number;
    };
    checkedAt?: string;
}

export interface BatchCheckOptions {
    minScore?: number;
    threshold?: number;
    includeBreakdown?: boolean;
}

export interface CertIDOptions {
    apiKey?: string;
    rpcUrl?: string;
    baseURL?: string;
}

export declare class CertID {
    constructor(options?: CertIDOptions);

    /**
     * Check trust score for a single address
     */
    check(address: string): Promise<TrustScoreResult>;

    /**
     * Check trust score for a single address
     */
    checkSybil(address: string): Promise<TrustScoreResult>;

    /**
     * Batch check multiple addresses
     */
    batchCheck(addresses: string[], options?: BatchCheckOptions): Promise<TrustScoreResult[]>;

    /**
     * Get trust score history for an address
     */
    getHistory(address: string): Promise<TrustScoreResult[]>;

    /**
     * Filter addresses to only include likely real users
     */
    filterReal(addresses: string[], minScore?: number): Promise<string[]>;

    /**
     * Filter addresses to only include suspicious accounts
     */
    filterSuspicious(addresses: string[], maxScore?: number): Promise<string[]>;
}

export default CertID;
