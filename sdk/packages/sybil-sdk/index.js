/**
 * CertID Sybil Resistance SDK
 * Official JavaScript client for CertID Trust Score API
 * 
 * @package @certid/sybil-sdk
 * @version 1.0.0
 * @license Apache-2.0
 */

class CertID {
    /**
     * Create a new CertID Sybil client
     * @param {Object} options - Configuration options
     * @param {string} [options.apiKey] - Your CertID API key (optional for basic usage)
     * @param {string} [options.rpcUrl='https://api.c3rt.org/api/v1'] - API base URL
     */
    constructor(options = {}) {
        this.apiKey = options.apiKey;
        this.rpcUrl = options.rpcUrl || options.baseURL || 'https://api.c3rt.org/api/v1';
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

        const url = `${this.rpcUrl}${endpoint}`;

        // Use fetch if available (browser/Node 18+), otherwise try node-fetch
        const fetchFn = typeof fetch !== 'undefined' ? fetch : require('node-fetch');

        const response = await fetchFn(url, {
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
     * Check trust score for a single address (alias for checkSybil)
     * @param {string} address - Blockchain address to check
     * @returns {Promise<Object>} Trust score result
     */
    async check(address) {
        return this.checkSybil(address);
    }

    /**
     * Check trust score for a single address
     * @param {string} address - Blockchain address to check
     * @returns {Promise<Object>} Trust score result
     * @example
     * const result = await certid.checkSybil('cert1abc...');
     * console.log(result.trustScore); // 85
     * console.log(result.isLikelyHuman); // true
     */
    async checkSybil(address) {
        const result = await this._request(`/sybil/check/${address}`);
        return {
            address: result.address,
            trustScore: result.trust_score,
            isLikelyHuman: result.is_likely_human,
            factors: result.factors,
            checkedAt: result.checked_at,
        };
    }

    /**
     * Batch check multiple addresses
     * @param {string[]} addresses - Array of addresses to check
     * @param {Object} [options] - Options
     * @param {number} [options.minScore=50] - Trust score threshold for classification
     * @param {boolean} [options.includeBreakdown=false] - Include detailed breakdown
     * @returns {Promise<Array>} Array of results
     * @example
     * const results = await certid.batchCheck(['cert1...', 'cert2...']);
     * const eligible = results.filter(r => r.trustScore >= 50);
     */
    async batchCheck(addresses, options = {}) {
        if (!Array.isArray(addresses)) {
            throw new Error('Addresses must be an array');
        }
        if (addresses.length > 100) {
            throw new Error('Batch size limited to 100 addresses');
        }

        const threshold = options.minScore || options.threshold || 50;

        const response = await this._request('/sybil/batch', {
            method: 'POST',
            body: JSON.stringify({ addresses, threshold }),
        });

        return response.results.map(r => ({
            address: r.address,
            trustScore: r.trust_score,
            isLikelyHuman: r.is_likely_human,
            factors: r.factors,
        }));
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
     * const eligible = await certid.filterReal(allAddresses, 60);
     * await distributeAirdrop(eligible);
     */
    async filterReal(addresses, minScore = 50) {
        const results = await this.batchCheck(addresses, { minScore });
        return results
            .filter(r => r.isLikelyHuman)
            .map(r => r.address);
    }

    /**
     * Filter a list to only include suspicious accounts
     * @param {string[]} addresses - Addresses to check
     * @param {number} [maxScore=30] - Maximum trust score for suspicious classification
     * @returns {Promise<string[]>} Suspicious addresses
     */
    async filterSuspicious(addresses, maxScore = 30) {
        const results = await this.batchCheck(addresses, { minScore: maxScore });
        return results
            .filter(r => !r.isLikelyHuman)
            .map(r => r.address);
    }
}

// Export for CommonJS
module.exports = { CertID };
module.exports.CertID = CertID;
module.exports.default = CertID;
