/**
 * CertID Sybil Resistance SDK
 * Official JavaScript client for CertID Trust Score API
 * 
 * @package @certid/sybil-sdk
 * @version 1.0.0
 */

class CertIDSybilClient {
    /**
     * Create a new CertID Sybil client
     * @param {Object} options - Configuration options
     * @param {string} options.apiKey - Your CertID API key
     * @param {string} [options.baseURL='https://api.c3rt.org/api/v1'] - API base URL
     */
    constructor(options = {}) {
        this.apiKey = options.apiKey;
        this.baseURL = options.baseURL || 'https://api.c3rt.org/api/v1';

        if (!this.apiKey) {
            console.warn('CertID: No API key provided. Rate limits will be restricted.');
        }
    }

    /**
     * Make an API request
     * @private
     */
    async _request(endpoint, options = {}) {
        const headers = {
            'Content-Type': 'application/json',
            ...options.headers,
        };

        if (this.apiKey) {
            headers['X-API-Key'] = this.apiKey;
        }

        const response = await fetch(`${this.baseURL}${endpoint}`, {
            ...options,
            headers,
        });

        if (!response.ok) {
            const error = await response.json().catch(() => ({ error: 'Unknown error' }));
            throw new Error(`CertID API Error: ${error.error || response.statusText}`);
        }

        return response.json();
    }

    /**
     * Check trust score for a single address
     * @param {string} address - Blockchain address to check
     * @returns {Promise<Object>} Trust score result
     * @example
     * const result = await client.check('cert1abc...');
     * console.log(result.trust_score); // 85
     * console.log(result.is_likely_human); // true
     */
    async check(address) {
        return this._request(`/sybil/check/${address}`);
    }

    /**
     * Batch check multiple addresses
     * @param {string[]} addresses - Array of addresses to check
     * @param {number} [threshold=50] - Trust score threshold for classification
     * @returns {Promise<Object>} Batch check results with summary
     * @example
     * const results = await client.batchCheck(['cert1...', 'cert2...'], 60);
     * console.log(results.summary.likely_real); // 42
     * console.log(results.summary.suspicious); // 8
     */
    async batchCheck(addresses, threshold = 50) {
        if (!Array.isArray(addresses)) {
            throw new Error('Addresses must be an array');
        }
        if (addresses.length > 100) {
            throw new Error('Batch size limited to 100 addresses');
        }

        return this._request('/sybil/batch', {
            method: 'POST',
            body: JSON.stringify({ addresses, threshold }),
        });
    }

    /**
     * Get trust score history for an address
     * @param {string} address - Blockchain address
     * @returns {Promise<Array>} Historical trust score data
     */
    async getHistory(address) {
        return this._request(`/sybil/history/${address}`);
    }

    /**
     * Filter a list of addresses to only include likely real users
     * @param {string[]} addresses - Addresses to filter
     * @param {number} [minScore=50] - Minimum trust score
     * @returns {Promise<string[]>} Filtered addresses
     * @example
     * // Airdrop protection
     * const eligible = await client.filterReal(allAddresses, 60);
     * await distributeAirdrop(eligible);
     */
    async filterReal(addresses, minScore = 50) {
        const results = await this.batchCheck(addresses, minScore);
        return results.results
            .filter(r => r.is_likely_human)
            .map(r => r.address);
    }

    /**
     * Filter a list to only include suspicious accounts
     * @param {string[]} addresses - Addresses to check
     * @param {number} [maxScore=30] - Maximum trust score for suspicious classification
     * @returns {Promise<string[]>} Suspicious addresses
     */
    async filterSuspicious(addresses, maxScore = 30) {
        const results = await this.batchCheck(addresses, maxScore);
        return results.results
            .filter(r => !r.is_likely_human)
            .map(r => r.address);
    }

    /**
     * Get detailed breakdown for a single address
     * @param {string} address - Address to analyze
     * @returns {Promise<Object>} Detailed trust factors
     * @example
     * const details = await client.getDetails('cert1...');
     * console.log(details.factors.kyc_verified); // true
     * console.log(details.factors.social_verifications); // 3
     */
    async getDetails(address) {
        const result = await this.check(address);
        return {
            address: result.address,
            trust_score: result.trust_score,
            factors: result.factors,
            checked_at: result.checked_at,
        };
    }
}

// Node.js support
if (typeof module !== 'undefined' && module.exports) {
    module.exports = CertIDSybilClient;
}

// Browser support
if (typeof window !== 'undefined') {
    window.CertIDSybilClient = CertIDSybilClient;
}

// Usage Examples (for documentation)
/* 

// Initialize client
const certid = new CertIDSybilClient({
    apiKey: 'cert_live_...'
});

// Single address check
const result = await certid.check('cert1abc...');
if (result.is_likely_human) {
    console.log('Real user with score:', result.trust_score);
}

// Batch check for airdrop
const participants = ['cert1...', 'cert2...', ...];
const eligible = await certid.filterReal(participants, 60);
console.log(`${eligible.length} eligible for airdrop`);

// Detailed analysis
const details = await certid.getDetails('cert1...');
console.log('KYC verified:', details.factors.kyc_verified);
console.log('Social accounts:', details.factors.social_verifications);

*/
