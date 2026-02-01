/**
 * CertID Sybil Resistance SDK
 * ESM Entry Point
 */

class CertID {
    constructor(options = {}) {
        this.apiKey = options.apiKey;
        this.rpcUrl = options.rpcUrl || options.baseURL || 'https://api.c3rt.org/api/v1';
    }

    async _request(endpoint, options = {}) {
        const headers = {
            'Content-Type': 'application/json',
            ...options.headers,
        };

        if (this.apiKey) {
            headers['X-API-Key'] = this.apiKey;
        }

        const response = await fetch(`${this.rpcUrl}${endpoint}`, {
            ...options,
            headers,
        });

        if (!response.ok) {
            const error = await response.json().catch(() => ({ error: 'Unknown error' }));
            throw new Error(`CertID API Error: ${error.error || response.statusText}`);
        }

        return response.json();
    }

    async check(address) {
        return this.checkSybil(address);
    }

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

    async getHistory(address) {
        return this._request(`/sybil/history/${address}`);
    }

    async filterReal(addresses, minScore = 50) {
        const results = await this.batchCheck(addresses, { minScore });
        return results.filter(r => r.isLikelyHuman).map(r => r.address);
    }

    async filterSuspicious(addresses, maxScore = 30) {
        const results = await this.batchCheck(addresses, { minScore: maxScore });
        return results.filter(r => !r.isLikelyHuman).map(r => r.address);
    }
}

export { CertID };
export default CertID;
