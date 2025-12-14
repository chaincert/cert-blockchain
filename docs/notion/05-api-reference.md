# API Reference

## Base URL

- **Local:** http://localhost:3000/api/v1
- **Production:** https://api.c3rt.org/api/v1

---

## Authentication

All endpoints (except health) require JWT authentication.

### Header

```
Authorization: Bearer <jwt-token>
```

### JWT Claims

```json
{
  "address": "cert1...",
  "exp": 1765619363
}
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

