#!/bin/bash
# CERT Blockchain Firewall Configuration Script
# Server: 172.239.32.74 (C3rt.org)
# Implements security best practices for production deployment

set -e

echo "=============================================="
echo "  CERT Blockchain Firewall Configuration"
echo "  Server: 172.239.32.74"
echo "=============================================="
echo ""

# Check if running as root
if [ "$EUID" -ne 0 ]; then 
    echo "ERROR: This script must be run as root"
    echo "Please run: sudo $0"
    exit 1
fi

# Step 1: Install fail2ban
echo "[1/5] Installing fail2ban for SSH protection..."
apt update -qq
apt install -y fail2ban

# Configure fail2ban for SSH
cat > /etc/fail2ban/jail.local << 'EOF'
[DEFAULT]
bantime = 3600
findtime = 600
maxretry = 5
destemail = admin@c3rt.org
sendername = Fail2Ban

[sshd]
enabled = true
port = 22
filter = sshd
logpath = /var/log/auth.log
maxretry = 3
bantime = 7200
EOF

systemctl enable fail2ban
systemctl restart fail2ban
echo "✅ fail2ban installed and configured"
echo ""

# Step 2: Reset UFW to clean state
echo "[2/5] Resetting UFW firewall..."
ufw --force reset
echo "✅ UFW reset"
echo ""

# Step 3: Configure UFW rules
echo "[3/5] Configuring firewall rules..."

# Default policies
ufw default deny incoming
ufw default allow outgoing

# Allow SSH (CRITICAL - don't lock yourself out!)
ufw allow 22/tcp comment 'SSH access'

# Allow HTTP/HTTPS (nginx reverse proxy)
ufw allow 80/tcp comment 'HTTP (redirect to HTTPS)'
ufw allow 443/tcp comment 'HTTPS (nginx)'

# Allow Tendermint P2P (validator communication)
ufw allow 26656/tcp comment 'Tendermint P2P'

# DENY direct access to internal services
# These should only be accessible via nginx reverse proxy or internally
ufw deny 3000/tcp comment 'Block direct API access'
ufw deny 5432/tcp comment 'Block PostgreSQL'
ufw deny 26657/tcp comment 'Block direct RPC access'
ufw deny 8080/tcp comment 'Block direct IPFS access'
ufw deny 8545/tcp comment 'Block direct EVM RPC'
ufw deny 8546/tcp comment 'Block direct EVM WebSocket'
ufw deny 1317/tcp comment 'Block direct Cosmos REST'
ufw deny 9090/tcp comment 'Block direct gRPC'

echo "✅ Firewall rules configured"
echo ""

# Step 4: Enable UFW
echo "[4/5] Enabling UFW firewall..."
ufw --force enable
echo "✅ UFW enabled"
echo ""

# Step 5: Display configuration
echo "[5/5] Firewall configuration complete!"
echo ""
echo "=============================================="
echo "  Current Firewall Status"
echo "=============================================="
ufw status numbered
echo ""

echo "=============================================="
echo "  fail2ban Status"
echo "=============================================="
fail2ban-client status sshd
echo ""

echo "=============================================="
echo "  Security Summary"
echo "=============================================="
echo "✅ SSH protected by fail2ban (max 3 attempts)"
echo "✅ HTTP/HTTPS open (ports 80, 443)"
echo "✅ P2P open for validators (port 26656)"
echo "✅ Direct API access blocked (port 3000)"
echo "✅ PostgreSQL blocked (port 5432)"
echo "✅ Direct RPC access blocked (port 26657)"
echo "✅ All services accessible via nginx reverse proxy"
echo ""

echo "=============================================="
echo "  Public Endpoints"
echo "=============================================="
echo "Web:  https://c3rt.org"
echo "API:  https://api.c3rt.org"
echo "RPC:  https://rpc.c3rt.org"
echo "IPFS: https://ipfs.c3rt.org"
echo "P2P:  172.239.32.74:26656"
echo ""

echo "=============================================="
echo "  Next Steps"
echo "=============================================="
echo "1. Test SSH access from another terminal"
echo "2. Verify services: curl https://api.c3rt.org/api/v1/health"
echo "3. Monitor fail2ban: sudo fail2ban-client status sshd"
echo "4. View logs: sudo tail -f /var/log/ufw.log"
echo ""

