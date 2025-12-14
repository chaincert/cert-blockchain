# API Reference (Continued)

## Encrypted Attestations

### Create Encrypted Attestation

```
POST /encrypted-attestations
```

Body:

```json
{
  "schema_uid": "0x4c11...",
  "recipient": "cert1abc...",
  "encrypted_data": "base64-encrypted-content",
  "access_list": ["cert1abc...", "cert1def..."],
  "ipfs_cid": "Qm..."
}
```

Response:

```json
{
  "uid": "0x708016f0a8402d423b2cb38b6f0160b66f4f78b6ceff59ae0f2db2ee16799724",
  "schema_uid": "0x4c11...",
  "attester": "cert1ktj...",
  "recipient": "cert1abc...",
  "ipfs_cid": "Qm...",
  "revoked": false
}
```

### Get Encrypted Attestation

```
GET /encrypted-attestations/:uid
```

### List by Attester

```
GET /encrypted-attestations/by-attester/:address
```

### List by Recipient

```
GET /encrypted-attestations/by-recipient/:address
```

---

## RPC Endpoints (CometBFT)

Base URL: http://localhost:26657

### Status

```
GET /status
```

Response includes:

- Node info
- Sync info (latest block height, time)
- Validator info

### Block

```
GET /block?height=<height>
```

### Latest Block

```
GET /block
```

### Transaction

```
GET /tx?hash=<hash>
```

### Broadcast Transaction (Sync)

```
POST /broadcast_tx_sync
```

### Broadcast Transaction (Async)

```
POST /broadcast_tx_async
```

### ABCI Query

```
GET /abci_query?path=<path>&data=<data>
```

---

## gRPC Endpoints

Base URL: localhost:9090

### Attestation Queries

- `cert.attestation.v1.Query/Schema`
- `cert.attestation.v1.Query/Attestation`
- `cert.attestation.v1.Query/AttestationsByAttester`
- `cert.attestation.v1.Query/AttestationsByRecipient`

### CertID Queries

- `cert.certid.v1.Query/Identity`
- `cert.certid.v1.Query/IdentityByAddress`

---

## Error Codes

| Code | Description |
|------|-------------|
| 400 | Bad Request - Invalid parameters |
| 401 | Unauthorized - Missing or invalid JWT |
| 404 | Not Found - Resource doesn't exist |
| 409 | Conflict - Resource already exists |
| 500 | Internal Server Error |

