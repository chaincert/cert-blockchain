#!/bin/bash
# Deploy and apply firewall configuration to remote server
# Server: 172.239.32.74 (C3rt.org)

set -e

SERVER="172.239.32.74"
USER="root"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo "=============================================="
echo "  Deploying Firewall Configuration"
echo "  Server: ${SERVER}"
echo "=============================================="
echo ""

# Check if we can connect
echo "[1/3] Testing SSH connection..."
if ! ssh -o ConnectTimeout=10 ${USER}@${SERVER} "echo 'Connection successful'"; then
    echo "ERROR: Cannot connect to ${SERVER}"
    echo "Please ensure:"
    echo "  1. SSH key is configured"
    echo "  2. Server is accessible"
    echo "  3. You have root access"
    exit 1
fi
echo "✅ SSH connection successful"
echo ""

# Copy firewall script to server
echo "[2/3] Copying firewall configuration script..."
scp ${SCRIPT_DIR}/configure-firewall.sh ${USER}@${SERVER}:/tmp/configure-firewall.sh
ssh ${USER}@${SERVER} "chmod +x /tmp/configure-firewall.sh"
echo "✅ Script copied to server"
echo ""

# Execute firewall configuration
echo "[3/3] Executing firewall configuration on remote server..."
echo ""
echo "⚠️  WARNING: This will modify firewall rules!"
echo "⚠️  Make sure you have console access in case SSH is blocked."
echo ""
read -p "Continue? (yes/no): " confirm

if [ "$confirm" != "yes" ]; then
    echo "Aborted."
    exit 0
fi

echo ""
echo "Applying firewall configuration..."
echo "=============================================="
ssh ${USER}@${SERVER} "bash /tmp/configure-firewall.sh"

echo ""
echo "=============================================="
echo "  Deployment Complete!"
echo "=============================================="
echo ""
echo "Testing connectivity..."
echo ""

# Test HTTPS
echo -n "Testing HTTPS (api.c3rt.org): "
if curl -s --connect-timeout 5 https://api.c3rt.org/api/v1/health > /dev/null 2>&1; then
    echo "✅ OK"
else
    echo "❌ FAILED"
fi

# Test direct API access (should be blocked)
echo -n "Testing direct API access (should be blocked): "
if timeout 3 nc -zv ${SERVER} 3000 2>&1 | grep -q "succeeded"; then
    echo "⚠️  WARNING: Port 3000 is still accessible!"
else
    echo "✅ Blocked (correct)"
fi

# Test P2P port (should be open)
echo -n "Testing P2P port 26656 (should be open): "
if timeout 3 nc -zv ${SERVER} 26656 2>&1 | grep -q "succeeded"; then
    echo "✅ Open (correct)"
else
    echo "❌ Blocked (incorrect)"
fi

# Test PostgreSQL (should be blocked)
echo -n "Testing PostgreSQL 5432 (should be blocked): "
if timeout 3 nc -zv ${SERVER} 5432 2>&1 | grep -q "succeeded"; then
    echo "⚠️  WARNING: PostgreSQL is accessible!"
else
    echo "✅ Blocked (correct)"
fi

echo ""
echo "=============================================="
echo "  Firewall configuration applied successfully!"
echo "=============================================="

