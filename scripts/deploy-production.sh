#!/bin/bash
# CERT Blockchain Production Deployment Script
# Server: 172.239.32.74 (C3rt.org)
# Per Whitepaper Section 8: API Specifications

set -e

# Configuration
DEPLOY_DIR="/opt/cert-blockchain"
DOMAIN="c3rt.org"
SERVER_IP="172.239.32.74"

echo "=============================================="
echo "  CERT Blockchain Production Deployment"
echo "  Server: ${SERVER_IP}"
echo "  Domain: ${DOMAIN}"
echo "=============================================="

# Step 1: Check Docker is installed
echo ""
echo "Step 1: Checking Docker installation..."
if ! command -v docker &> /dev/null; then
    echo "Installing Docker..."
    curl -fsSL https://get.docker.com | sh
    systemctl enable docker
    systemctl start docker
fi
docker --version
docker-compose --version || docker compose version

# Step 2: Create deployment directory
echo ""
echo "Step 2: Setting up deployment directory..."
mkdir -p ${DEPLOY_DIR}
cd ${DEPLOY_DIR}

# Step 3: Generate secure secrets if not exists
echo ""
echo "Step 3: Generating secure secrets..."
if [ ! -f .env ]; then
    echo "Creating production .env file..."
    cat > .env << EOF
# CERT Blockchain Production Configuration
# Generated: $(date -u +"%Y-%m-%dT%H:%M:%SZ")

# Database - CHANGE THIS!
POSTGRES_PASSWORD=$(openssl rand -base64 32 | tr -dc 'a-zA-Z0-9' | head -c 32)

# JWT Secret - CHANGE THIS!
JWT_SECRET=$(openssl rand -base64 64 | tr -dc 'a-zA-Z0-9' | head -c 64)

# IPFS Gateway (local or production)
IPFS_GATEWAY=http://localhost:8080

# Chain Configuration
CERT_CHAIN_ID=951753
CERT_MONIKER=cert-mainnet-1

# Network Configuration
CORS_ORIGINS=https://app.${DOMAIN},https://${DOMAIN}

# Logging
LOG_LEVEL=info
LOG_FORMAT=json
EOF
    chmod 600 .env
    echo "Generated .env with secure random secrets"
else
    echo ".env already exists, skipping..."
fi

# Step 4: Pull/build images
echo ""
echo "Step 4: Building Docker images..."
docker-compose build --no-cache

# Step 5: Stop existing containers (if any)
echo ""
echo "Step 5: Stopping existing containers..."
docker-compose down --remove-orphans || true

# Step 6: Start services
echo ""
echo "Step 6: Starting CERT Blockchain services..."
docker-compose up -d postgres
echo "Waiting for PostgreSQL to be healthy..."
sleep 10

docker-compose up -d certd
echo "Waiting for blockchain node to initialize..."
sleep 30

docker-compose up -d api
echo "Waiting for API server to start..."
sleep 10

docker-compose --profile full up -d ipfs
echo "Waiting for IPFS node to start..."
sleep 15

# Step 7: Verify services
echo ""
echo "Step 7: Verifying services..."
echo ""
echo "--- Docker Containers ---"
docker-compose ps

echo ""
echo "--- Blockchain Status ---"
curl -s http://localhost:26657/status | jq '.result.sync_info.latest_block_height' 2>/dev/null || echo "Checking..."

echo ""
echo "--- API Health ---"
curl -s http://localhost:3000/api/v1/health | jq . 2>/dev/null || echo "Checking..."

echo ""
echo "=============================================="
echo "  Deployment Complete!"
echo "=============================================="
echo ""
echo "Services running:"
echo "  - Blockchain RPC:  http://${SERVER_IP}:26657"
echo "  - REST API:        http://${SERVER_IP}:3000"
echo "  - IPFS Gateway:    http://${SERVER_IP}:8080"
echo "  - PostgreSQL:      ${SERVER_IP}:5432"
echo ""
echo "Next steps:"
echo "  1. Configure firewall (ufw/iptables)"
echo "  2. Set up nginx reverse proxy with SSL"
echo "  3. Configure DNS for ${DOMAIN}"
echo "  4. Update .env with production secrets"
echo ""

