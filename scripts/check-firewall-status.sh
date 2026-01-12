#!/bin/bash
# Check firewall status on remote server
# Server: 172.239.32.74 (C3rt.org)

SERVER="172.239.32.74"
USER="root"

echo "=============================================="
echo "  CERT Blockchain Firewall Status Check"
echo "  Server: ${SERVER}"
echo "=============================================="
echo ""

# Function to test port
test_port() {
    local port=$1
    local service=$2
    local should_be=$3
    
    echo -n "Port ${port} (${service}): "
    
    if timeout 3 nc -zv ${SERVER} ${port} 2>&1 | grep -q "succeeded"; then
        if [ "$should_be" = "open" ]; then
            echo "✅ OPEN (correct)"
        else
            echo "⚠️  OPEN (should be blocked!)"
        fi
    else
        if [ "$should_be" = "blocked" ]; then
            echo "✅ BLOCKED (correct)"
        else
            echo "❌ BLOCKED (should be open!)"
        fi
    fi
}

echo "Testing external port accessibility..."
echo ""

# Public ports (should be open)
echo "--- Public Ports (should be OPEN) ---"
test_port 80 "HTTP" "open"
test_port 443 "HTTPS" "open"
test_port 26656 "P2P" "open"

echo ""
echo "--- Internal Ports (should be BLOCKED) ---"
test_port 3000 "API" "blocked"
test_port 5432 "PostgreSQL" "blocked"
test_port 26657 "RPC" "blocked"
test_port 8080 "IPFS" "blocked"
test_port 8545 "EVM RPC" "blocked"
test_port 8546 "EVM WS" "blocked"
test_port 1317 "Cosmos REST" "blocked"
test_port 9090 "gRPC" "blocked"

echo ""
echo "--- Service Endpoints (via HTTPS) ---"

# Test HTTPS endpoints
echo -n "https://c3rt.org: "
if curl -s --connect-timeout 5 https://c3rt.org > /dev/null 2>&1; then
    echo "✅ OK"
else
    echo "❌ FAILED"
fi

echo -n "https://api.c3rt.org/api/v1/health: "
if curl -s --connect-timeout 5 https://api.c3rt.org/api/v1/health | grep -q "healthy"; then
    echo "✅ OK"
else
    echo "❌ FAILED"
fi

echo -n "https://rpc.c3rt.org/status: "
if curl -s --connect-timeout 5 https://rpc.c3rt.org/status 2>&1 | grep -q "502"; then
    echo "⚠️  Backend down (502)"
elif curl -s --connect-timeout 5 https://rpc.c3rt.org/status > /dev/null 2>&1; then
    echo "✅ OK"
else
    echo "❌ FAILED"
fi

echo ""
echo "--- Remote Server Status ---"
echo ""

# Check UFW status on remote server
echo "UFW Status:"
ssh ${USER}@${SERVER} "ufw status numbered" 2>/dev/null || echo "Cannot connect to server"

echo ""
echo "fail2ban Status:"
ssh ${USER}@${SERVER} "fail2ban-client status sshd 2>/dev/null || echo 'fail2ban not installed'" 2>/dev/null || echo "Cannot connect to server"

echo ""
echo "=============================================="
echo "  Status Check Complete"
echo "=============================================="

