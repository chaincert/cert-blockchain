# Deployment Guide

## Production Server

- **IP:** 172.239.32.74
- **Domain:** C3rt.org

---

## Quick Deployment

### 1. Transfer Package

```bash
scp cert-blockchain-deploy.zip root@172.239.32.74:/opt/
```

### 2. Extract & Deploy

```bash
ssh root@172.239.32.74
cd /opt
unzip cert-blockchain-deploy.zip
cd cert-blockchain
chmod +x scripts/deploy-production.sh
./scripts/deploy-production.sh
```

---

## Docker Services

### Start All Services

```bash
docker-compose --profile full up -d
```

### Service Ports

| Service | Internal Port | External Port |
|---------|---------------|---------------|
| certd (RPC) | 26657 | 26657 |
| certd (P2P) | 26656 | 26656 |
| certd (REST) | 1317 | 1317 |
| certd (gRPC) | 9090 | 9090 |
| PostgreSQL | 5432 | 5432 |
| API | 3000 | 3000 |
| IPFS Gateway | 8080 | 8080 |
| IPFS API | 5001 | 5001 |
| IPFS P2P | 4001 | 4001 |

---

## Nginx Configuration (Production)

Subdomains:

- **api.c3rt.org** → localhost:3000 (REST API)
- **rpc.c3rt.org** → localhost:26657 (CometBFT RPC)
- **ipfs.c3rt.org** → localhost:8080 (IPFS Gateway)

---

## Environment Variables

```bash
# .env file
POSTGRES_USER=cert
POSTGRES_PASSWORD=<generated-secure-password>
POSTGRES_DB=certdb

JWT_SECRET=<generated-secure-secret>

CHAIN_ID=951753
MONIKER=cert-validator

IPFS_GATEWAY=http://ipfs:8080
```

---

## SSL Certificates (Let's Encrypt)

```bash
# Install certbot
apt install certbot python3-certbot-nginx

# Generate certificates
certbot --nginx -d api.c3rt.org -d rpc.c3rt.org -d ipfs.c3rt.org

# Auto-renewal
certbot renew --dry-run
```

---

## Health Checks

```bash
# API Health
curl http://localhost:3000/api/v1/health

# Blockchain Status
curl http://localhost:26657/status

# IPFS Status
docker exec cert-ipfs ipfs id
```

---

## Troubleshooting

### Container won't start

```bash
docker-compose logs <service-name>
docker-compose down && docker-compose up -d
```

### Reset blockchain data

```bash
docker-compose down
docker volume rm cert-blockchain_certd-data
docker-compose up -d
```

### Database issues

```bash
docker exec -it cert-postgres psql -U cert -d certdb
```

