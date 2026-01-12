#!/bin/bash
# CERT Discourse - Quick Start Script
# For local testing with docker-compose

set -e

echo "=========================================="
echo "CERT Discourse - Quick Start (Docker Compose)"
echo "=========================================="
echo ""

cd "$(dirname "$0")"

# Check if .env exists
if [ ! -f .env ]; then
    echo "Creating .env file..."
    cp .env.example .env
    
    # Generate secrets
    SSO_SECRET=$(openssl rand -hex 32)
    DB_PASSWORD=$(openssl rand -hex 24)
    
    # Update .env
    sed -i "s/generate_sso_secret_here/${SSO_SECRET}/g" .env
    sed -i "s/generate_secure_password_here/${DB_PASSWORD}/g" .env
    
    echo "✓ Created .env with generated secrets"
    echo ""
    echo "⚠ IMPORTANT: Edit .env and add your SMTP_PASSWORD"
    echo ""
    read -p "Press Enter after updating SMTP credentials..."
fi

# Create network if it doesn't exist
if ! docker network inspect cert-network &> /dev/null; then
    echo "Creating cert-network..."
    docker network create cert-network
fi

echo "Starting Discourse with docker-compose..."
docker-compose up -d

echo ""
echo "✓ Discourse is starting..."
echo ""
echo "Services:"
echo "  - Discourse: http://localhost:8080"
echo "  - PostgreSQL: localhost:5432"
echo "  - Redis: localhost:6379"
echo ""
echo "View logs: docker-compose logs -f discourse"
echo "Stop: docker-compose down"
echo ""

