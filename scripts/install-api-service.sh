#!/bin/bash
# CERT API Service Installation Script
# Installs the cert-api systemd service with proper configuration

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

echo "=============================================="
echo "  CERT API Service Installation"
echo "=============================================="

# Check if running as root
if [ "$EUID" -ne 0 ]; then
    echo "Error: Please run as root (sudo)"
    exit 1
fi

# Check if binary exists
if [ ! -f "$PROJECT_DIR/cert-api-linux" ]; then
    echo "Building API server..."
    cd "$PROJECT_DIR"
    go build -o cert-api-linux ./cmd/api/
fi

# Get database password from docker if postgres is running
DB_PASSWORD=""
if docker ps | grep -q cert-postgres; then
    DB_PASSWORD=$(docker exec cert-postgres env | grep POSTGRES_PASSWORD | cut -d'=' -f2)
    echo "Found PostgreSQL password from running container"
fi

# Create environment file if it doesn't exist
if [ ! -f "$PROJECT_DIR/.env.api" ]; then
    echo "Creating .env.api configuration file..."
    cp "$PROJECT_DIR/.env.api.example" "$PROJECT_DIR/.env.api"
    
    if [ -n "$DB_PASSWORD" ]; then
        sed -i "s/YOUR_PASSWORD_HERE/$DB_PASSWORD/" "$PROJECT_DIR/.env.api"
        echo "Updated database password in .env.api"
    else
        echo "WARNING: Update DATABASE_URL password in $PROJECT_DIR/.env.api"
    fi
    
    # Generate random JWT secret
    JWT_SECRET=$(openssl rand -hex 32)
    sed -i "s/your-secure-jwt-secret-here/$JWT_SECRET/" "$PROJECT_DIR/.env.api"
    echo "Generated new JWT secret"
fi

# Stop existing service if running
if systemctl is-active --quiet cert-api; then
    echo "Stopping existing cert-api service..."
    systemctl stop cert-api
fi

# Kill any existing process
pkill -f cert-api-linux 2>/dev/null || true

# Copy service file
echo "Installing systemd service..."
cp "$PROJECT_DIR/deploy-package/services/cert-api.service" /etc/systemd/system/

# Reload systemd
systemctl daemon-reload

# Enable and start service
echo "Enabling and starting cert-api service..."
systemctl enable cert-api
systemctl start cert-api

# Wait for startup
sleep 3

# Check status
if systemctl is-active --quiet cert-api; then
    echo ""
    echo "=============================================="
    echo "  ✅ CERT API Service installed successfully!"
    echo "=============================================="
    echo ""
    echo "Service Status:"
    systemctl status cert-api --no-pager | head -15
    echo ""
    echo "Health Check:"
    curl -s http://localhost:3000/api/v1/health | jq . 2>/dev/null || curl -s http://localhost:3000/api/v1/health
    echo ""
    echo ""
    echo "Commands:"
    echo "  View logs:    journalctl -u cert-api -f"
    echo "  Restart:      systemctl restart cert-api"
    echo "  Stop:         systemctl stop cert-api"
    echo "  Status:       systemctl status cert-api"
else
    echo ""
    echo "=============================================="
    echo "  ❌ Service failed to start"
    echo "=============================================="
    echo ""
    echo "Check logs with: journalctl -u cert-api -n 50"
    systemctl status cert-api --no-pager
    exit 1
fi

