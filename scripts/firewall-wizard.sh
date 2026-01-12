#!/bin/bash
# CERT Blockchain Firewall Configuration Wizard
# Interactive script to guide through firewall setup

set -e

SERVER="172.239.32.74"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}=============================================="
echo "  CERT Blockchain Firewall Wizard"
echo "  Server: ${SERVER}"
echo -e "==============================================${NC}"
echo ""

# Function to test port
test_port() {
    local port=$1
    timeout 2 nc -zv ${SERVER} ${port} 2>&1 | grep -q succeeded
}

# Step 1: Check current status
echo -e "${BLUE}[Step 1/4] Checking current firewall status...${NC}"
echo ""

echo -n "Port 80 (HTTP): "
if test_port 80; then
    echo -e "${GREEN}✅ OPEN${NC}"
else
    echo -e "${RED}❌ CLOSED${NC}"
fi

echo -n "Port 443 (HTTPS): "
if test_port 443; then
    echo -e "${GREEN}✅ OPEN${NC}"
else
    echo -e "${RED}❌ CLOSED${NC}"
fi

echo -n "Port 3000 (API - should be blocked): "
if test_port 3000; then
    echo -e "${RED}⚠️  OPEN (SECURITY ISSUE!)${NC}"
    ISSUE_3000=1
else
    echo -e "${GREEN}✅ BLOCKED${NC}"
    ISSUE_3000=0
fi

echo -n "Port 26656 (P2P - should be open): "
if test_port 26656; then
    echo -e "${GREEN}✅ OPEN${NC}"
    ISSUE_26656=0
else
    echo -e "${RED}❌ BLOCKED (CRITICAL!)${NC}"
    ISSUE_26656=1
fi

echo -n "Port 5432 (PostgreSQL - should be blocked): "
if test_port 5432; then
    echo -e "${RED}⚠️  OPEN (SECURITY ISSUE!)${NC}"
    ISSUE_5432=1
else
    echo -e "${GREEN}✅ BLOCKED${NC}"
    ISSUE_5432=0
fi

echo ""

# Step 2: Analyze issues
echo -e "${BLUE}[Step 2/4] Analyzing configuration...${NC}"
echo ""

ISSUES_FOUND=0

if [ $ISSUE_3000 -eq 1 ]; then
    echo -e "${RED}❌ Issue: Port 3000 (API) is publicly accessible${NC}"
    echo "   Risk: Bypasses nginx security, rate limiting, SSL"
    echo "   Action: Block port 3000"
    echo ""
    ISSUES_FOUND=$((ISSUES_FOUND + 1))
fi

if [ $ISSUE_26656 -eq 1 ]; then
    echo -e "${RED}❌ Issue: Port 26656 (P2P) is blocked${NC}"
    echo "   Risk: Validators cannot connect, blockchain cannot sync"
    echo "   Action: Open port 26656"
    echo ""
    ISSUES_FOUND=$((ISSUES_FOUND + 1))
fi

if [ $ISSUE_5432 -eq 1 ]; then
    echo -e "${RED}❌ Issue: Port 5432 (PostgreSQL) is publicly accessible${NC}"
    echo "   Risk: Database exposed to internet"
    echo "   Action: Block port 5432"
    echo ""
    ISSUES_FOUND=$((ISSUES_FOUND + 1))
fi

if [ $ISSUES_FOUND -eq 0 ]; then
    echo -e "${GREEN}✅ No issues found! Firewall is properly configured.${NC}"
    echo ""
    exit 0
fi

echo -e "${YELLOW}Total issues found: ${ISSUES_FOUND}${NC}"
echo ""

# Step 3: Offer solutions
echo -e "${BLUE}[Step 3/4] Available solutions:${NC}"
echo ""
echo "1. Automated fix (recommended) - Run configuration script on server"
echo "2. Manual fix - Show commands to run manually"
echo "3. View documentation - Open firewall setup guide"
echo "4. Exit - I'll fix this later"
echo ""

read -p "Choose option (1-4): " choice

case $choice in
    1)
        echo ""
        echo -e "${BLUE}Running automated firewall configuration...${NC}"
        echo ""
        
        if [ ! -f "${SCRIPT_DIR}/apply-firewall-remote.sh" ]; then
            echo -e "${RED}Error: apply-firewall-remote.sh not found${NC}"
            exit 1
        fi
        
        bash "${SCRIPT_DIR}/apply-firewall-remote.sh"
        ;;
    
    2)
        echo ""
        echo -e "${BLUE}Manual Fix Instructions:${NC}"
        echo ""
        echo "SSH to the server and run these commands:"
        echo ""
        echo -e "${YELLOW}ssh root@${SERVER}${NC}"
        echo ""
        echo "# Install fail2ban"
        echo "apt update && apt install -y fail2ban"
        echo ""
        echo "# Configure firewall"
        echo "ufw allow 22/tcp"
        echo "ufw allow 80/tcp"
        echo "ufw allow 443/tcp"
        echo "ufw allow 26656/tcp"
        echo "ufw deny 3000/tcp"
        echo "ufw deny 5432/tcp"
        echo "ufw --force enable"
        echo ""
        echo "# Verify"
        echo "ufw status numbered"
        echo ""
        ;;
    
    3)
        echo ""
        echo -e "${BLUE}Documentation:${NC}"
        echo ""
        echo "📖 Complete Guide: cert-blockchain/FIREWALL-SETUP.md"
        echo "📖 Action Required: cert-blockchain/FIREWALL-ACTION-REQUIRED.md"
        echo "📖 Deployment Guide: cert-blockchain/DEPLOYMENT.md"
        echo ""
        ;;
    
    4)
        echo ""
        echo "Exiting. Please fix firewall issues as soon as possible."
        echo ""
        exit 0
        ;;
    
    *)
        echo "Invalid option"
        exit 1
        ;;
esac

echo ""
echo -e "${BLUE}[Step 4/4] Next steps:${NC}"
echo ""
echo "1. Verify configuration with: ./scripts/check-firewall-status.sh"
echo "2. Test endpoints: curl https://api.c3rt.org/api/v1/health"
echo "3. Monitor logs for 24 hours"
echo ""
echo -e "${GREEN}Done!${NC}"

