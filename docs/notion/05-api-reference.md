# API Reference

## Base URL

- **Local:** http://localhost:3000/api/v1
- **Production:** https://api.c3rt.org/api/v1

---

## Authentication

Protected endpoints require JWT authentication obtained via EIP-191 challenge/response flow.

### How to Obtain a JWT Token

#### Step 1: Request Challenge

```bash
GET /auth/challenge?address=cert1your_address
```

**Response:**

```json
{
  "challenge": "CERT Authentication\n\nAddress: cert1...\nNonce: a1b2c3...\nIssued At: 2025-01-09T12:00:00Z\n\nBy signing, you authorize this app to obtain a short-lived JWT for CERT APIs.",
  "nonce": "a1b2c3d4e5f6...",
  "expiresAt": 1736427600
}
```

**Note:** Challenge expires in 5 minutes.

#### Step 2: Sign Challenge with Wallet

Sign the challenge message using your wallet (Keplr, MetaMask, etc.) with EIP-191 signature standard.

**Example with Keplr:**

```javascript
const signature = await window.keplr.signArbitrary(
  "951753",  // Chain ID
  address,
  challenge
);
```

**Example with MetaMask (EVM addresses):**

```javascript
const signature = await ethereum.request({
  method: 'personal_sign',
  params: [challenge, address]
});
```

#### Step 3: Verify Signature and Get JWT

```bash
POST /auth/verify
Content-Type: application/json

{
  "address": "cert1your_address",
  "nonce": "a1b2c3d4e5f6...",
  "signature": "0x1234abcd..."
}
```

**Response:**

```json
{
  "ok": true,
  "address": "cert1your_address",
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expiresAt": 1736470800
}
```

**JWT is valid for 12 hours.**

#### Step 4: Use JWT in API Requests

Include the JWT token in the Authorization header for all protected endpoints:

```bash
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

### JWT Claims

```json
{
  "address": "cert1...",
  "nonce": "a1b2c3...",
  "iat": 1736427600,
  "exp": 1736470800
}
```

### Complete Example (cURL)

```bash
# 1. Get challenge
CHALLENGE=$(curl -s "https://api.c3rt.org/api/v1/auth/challenge?address=cert1abc..." | jq -r '.challenge')

# 2. Sign challenge (use your wallet - this is pseudocode)
SIGNATURE=$(sign_with_wallet "$CHALLENGE")

# 3. Get JWT token
TOKEN=$(curl -s -X POST https://api.c3rt.org/api/v1/auth/verify \
  -H "Content-Type: application/json" \
  -d "{\"address\":\"cert1abc...\",\"nonce\":\"...\",\"signature\":\"$SIGNATURE\"}" \
  | jq -r '.token')

# 4. Use token in API calls
curl -H "Authorization: Bearer $TOKEN" \
  https://api.c3rt.org/api/v1/attestations
```

### Complete Example (JavaScript)

```javascript
async function authenticate(address, signMessage) {
  // 1. Get challenge from server
  const challengeRes = await fetch(
    `https://api.c3rt.org/api/v1/auth/challenge?address=${address}`
  );
  const { challenge, nonce } = await challengeRes.json();

  // 2. Sign the challenge with wallet
  const signature = await signMessage(challenge);

  // 3. Verify and get JWT
  const verifyRes = await fetch('https://api.c3rt.org/api/v1/auth/verify', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ address, nonce, signature })
  });

  const { token, expiresAt } = await verifyRes.json();

  // 4. Store token for future requests
  localStorage.setItem('jwt_token', token);

  return token;
}

// Use the token
const token = await authenticate(myAddress, mySignFunction);

// Make authenticated API call
const response = await fetch('https://api.c3rt.org/api/v1/attestations', {
  headers: {
    'Authorization': `Bearer ${token}`,
    'Content-Type': 'application/json'
  },
  method: 'POST',
  body: JSON.stringify({ schema_uid: '0x...', data: '...' })
});
```

---

## Endpoints

### Health Check

```
GET /health
```

Response:

```json
{
  "status": "healthy",
  "timestamp": 1765532963,
  "version": "1.0.0"
}
```

---

### Stats

```
GET /stats
```

Response:

```json
{
  "total_attestations": 0,
  "total_encrypted_attestations": 0,
  "total_profiles": 0,
  "total_schemas": 0
}
```

---

### Schemas

#### Create Schema

```
POST /schemas
```

Body:

```json
{
  "name": "EducationalCredential",
  "description": "Schema for educational credentials",
  "schema": "{\"type\":\"object\",\"properties\":{\"degree\":\"string\"}}"
}
```

Response:

```json
{
  "uid": "0x4c110363178f4ac6b868cf918de62978258f1332fbfd8ad4b29f48abb63ce600",
  "name": "EducationalCredential",
  "resolver": "",
  "revocable": true
}
```

#### Get Schema

```
GET /schemas/:uid
```

#### List Schemas

```
GET /schemas
```

---

### Attestations

#### Create Attestation

```
POST /attestations
```

Body:

```json
{
  "schema_uid": "0x4c11...",
  "recipient": "cert1abc...",
  "data": "{\"degree\":\"PhD\",\"institution\":\"MIT\"}",
  "revocable": true
}
```

Response:

```json
{
  "uid": "0xeb135da7fcadefdecc84a3f6d086a1ccf6bfe7a7f76be6fdd05630edadaf98a0",
  "schema_uid": "0x4c11...",
  "attester": "cert1ktj...",
  "recipient": "cert1abc...",
  "revoked": false
}
```

#### Get Attestation

```
GET /attestations/:uid
```

#### List by Attester

```
GET /attestations/by-attester/:address
```

#### List by Recipient

```
GET /attestations/by-recipient/:address
```

