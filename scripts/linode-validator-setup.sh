#!/bin/bash

# CERT Blockchain - Linode Validator Setup Script
# Run this on a fresh Ubuntu 24.04 Linode instance
# Usage: curl -sSL https://raw.githubusercontent.com/chaincert/cert-blockchain/main/scripts/linode-validator-setup.sh | bash

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Configuration
CHAIN_ID="cert-testnet-1"
KEYRING_BACKEND="test"
HOME_DIR="$HOME/.certd"
DENOM="ucert"
PRIMARY_VALIDATOR_IP="172.239.32.74"
MIN_STAKE="10000000000"  # 10,000 CERT
FUND_AMOUNT="15000000000" # 15,000 CERT

echo -e "${BLUE}"
echo "================================================"
echo "  CERT Blockchain - Linode Validator Setup"
echo "  Chain: $CHAIN_ID"
echo "================================================"
echo -e "${NC}"

# Check if running as root
if [ "$EUID" -ne 0 ]; then
    echo -e "${RED}Please run as root${NC}"
    exit 1
fi

# Prompt for configuration
read -p "Enter validator moniker (e.g., validator-2): " MONIKER
if [ -z "$MONIKER" ]; then
    echo -e "${RED}Moniker is required${NC}"
    exit 1
fi

echo ""
echo -e "${YELLOW}Step 1: Installing dependencies...${NC}"
apt update && apt upgrade -y
apt install -y git make golang-go build-essential curl jq ufw

# Install Go 1.22+ if needed
GO_VERSION=$(go version 2>/dev/null | grep -oP 'go\K[0-9]+\.[0-9]+' || echo "0")
if (( $(echo "$GO_VERSION < 1.22" | bc -l) )); then
    echo "Installing Go 1.22..."
    wget -q https://go.dev/dl/go1.22.5.linux-amd64.tar.gz
    rm -rf /usr/local/go && tar -C /usr/local -xzf go1.22.5.linux-amd64.tar.gz
    rm go1.22.5.linux-amd64.tar.gz
    export PATH=$PATH:/usr/local/go/bin
    echo 'export PATH=$PATH:/usr/local/go/bin:$HOME/go/bin' >> ~/.bashrc
fi
export PATH=$PATH:/usr/local/go/bin:$HOME/go/bin

echo -e "${GREEN}✅ Dependencies installed${NC}"

echo ""
echo -e "${YELLOW}Step 2: Configuring firewall...${NC}"
ufw allow 22/tcp
ufw allow 26656/tcp
ufw allow 26657/tcp
ufw --force enable
echo -e "${GREEN}✅ Firewall configured${NC}"

echo ""
echo -e "${YELLOW}Step 3: Cloning and building CERT blockchain...${NC}"
cd /opt
if [ -d "cert-blockchain" ]; then
    cd cert-blockchain && git pull
else
    git clone https://gitlab.com/cert-dev/cert-blockchain.git
    cd cert-blockchain
fi

make install
echo -e "${GREEN}✅ certd installed: $(certd version)${NC}"

echo ""
echo -e "${YELLOW}Step 4: Initializing node...${NC}"
rm -rf $HOME_DIR
certd init "$MONIKER" --chain-id "$CHAIN_ID" --home "$HOME_DIR"
echo -e "${GREEN}✅ Node initialized${NC}"

echo ""
echo -e "${YELLOW}Step 5: Creating validator key...${NC}"
echo -e "${RED}⚠️  IMPORTANT: Save the mnemonic shown below!${NC}"
echo ""
certd keys add validator --keyring-backend "$KEYRING_BACKEND" --home "$HOME_DIR"

VALIDATOR_ADDRESS=$(certd keys show validator -a --keyring-backend "$KEYRING_BACKEND" --home "$HOME_DIR")
echo ""
echo -e "${GREEN}✅ Validator address: ${BLUE}$VALIDATOR_ADDRESS${NC}"
echo ""
echo -e "${RED}⚠️  BACKUP YOUR KEYS NOW:${NC}"
echo "   - Mnemonic: (shown above)"
echo "   - Private key: $HOME_DIR/config/priv_validator_key.json"
echo ""
read -p "Press Enter after you've saved the mnemonic..."

echo ""
echo -e "${YELLOW}Step 6: Fetching genesis file...${NC}"
echo "Copying genesis.json from primary validator..."
scp -o StrictHostKeyChecking=no root@$PRIMARY_VALIDATOR_IP:/opt/cert-blockchain/data/certd/config/genesis.json \
    $HOME_DIR/config/genesis.json

if [ ! -f "$HOME_DIR/config/genesis.json" ]; then
    echo -e "${RED}❌ Failed to copy genesis.json${NC}"
    echo "Manually run: scp root@$PRIMARY_VALIDATOR_IP:/opt/cert-blockchain/data/certd/config/genesis.json $HOME_DIR/config/genesis.json"
    exit 1
fi
certd genesis validate --home "$HOME_DIR"
echo -e "${GREEN}✅ Genesis file validated${NC}"

echo ""
echo -e "${YELLOW}Step 7: Configuring peers...${NC}"
echo "Fetching node ID from primary validator..."
PRIMARY_NODE_ID=$(ssh -o StrictHostKeyChecking=no root@$PRIMARY_VALIDATOR_IP \
    "certd tendermint show-node-id --home /opt/cert-blockchain/data/certd" 2>/dev/null || echo "")

if [ -z "$PRIMARY_NODE_ID" ]; then
    read -p "Could not auto-fetch. Enter primary validator node ID: " PRIMARY_NODE_ID
fi

PEERS="${PRIMARY_NODE_ID}@${PRIMARY_VALIDATOR_IP}:26656"
sed -i "s/^persistent_peers = .*/persistent_peers = \"$PEERS\"/" $HOME_DIR/config/config.toml
echo -e "${GREEN}✅ Peers configured: $PEERS${NC}"

# Optimize config
sed -i 's/timeout_commit = "5s"/timeout_commit = "2s"/' $HOME_DIR/config/config.toml
sed -i 's/minimum-gas-prices = ""/minimum-gas-prices = "0.0001ucert"/' $HOME_DIR/config/app.toml

echo ""
echo -e "${YELLOW}Step 8: Creating systemd service...${NC}"
cat > /etc/systemd/system/certd.service << EOF
[Unit]
Description=CERT Blockchain Validator
After=network.target

[Service]
Type=simple
User=root
ExecStart=$(which certd) start --home $HOME_DIR
Restart=always
RestartSec=3
LimitNOFILE=65535

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
systemctl enable certd
echo -e "${GREEN}✅ Systemd service created${NC}"

echo ""
echo -e "${BLUE}================================================${NC}"
echo -e "${GREEN}  Setup Complete!${NC}"
echo -e "${BLUE}================================================${NC}"
echo ""
echo -e "${YELLOW}NEXT STEPS:${NC}"
echo ""
echo "1. ${BLUE}Fund your validator${NC} (on primary validator 172.239.32.74):"
echo ""
echo "   certd tx bank send validator $VALIDATOR_ADDRESS ${FUND_AMOUNT}ucert \\"
echo "     --chain-id $CHAIN_ID \\"
echo "     --keyring-backend test \\"
echo "     --home /opt/cert-blockchain/data/certd \\"
echo "     --fees 10000ucert -y"
echo ""
echo "2. ${BLUE}Start the node${NC}:"
echo ""
echo "   systemctl start certd"
echo "   journalctl -u certd -f"
echo ""
echo "3. ${BLUE}Wait for sync${NC} (catching_up should be false):"
echo ""
echo "   certd status | jq .sync_info.catching_up"
echo ""
echo "4. ${BLUE}Create validator${NC} (after sync complete):"
echo ""
echo "   certd tx staking create-validator \\"
echo "     --amount=${MIN_STAKE}ucert \\"
echo "     --pubkey=\$(certd tendermint show-validator --home $HOME_DIR) \\"
echo "     --moniker=\"$MONIKER\" \\"
echo "     --chain-id=$CHAIN_ID \\"
echo "     --commission-rate=\"0.05\" \\"
echo "     --commission-max-rate=\"0.20\" \\"
echo "     --commission-max-change-rate=\"0.01\" \\"
echo "     --min-self-delegation=\"1\" \\"
echo "     --from=validator \\"
echo "     --keyring-backend=test \\"
echo "     --home=$HOME_DIR \\"
echo "     --fees=10000ucert -y"
echo ""
echo "5. ${BLUE}Verify validator${NC}:"
echo ""
echo "   certd query staking validators"
echo ""
echo -e "${GREEN}Validator Address: $VALIDATOR_ADDRESS${NC}"
echo -e "${GREEN}Home Directory: $HOME_DIR${NC}"
echo ""

# Save info to file
cat > /root/validator-info.txt << EOF
CERT Validator Info
===================
Moniker: $MONIKER
Address: $VALIDATOR_ADDRESS
Chain ID: $CHAIN_ID
Home: $HOME_DIR
Primary Peer: $PEERS

Commands:
---------
Start:   systemctl start certd
Stop:    systemctl stop certd
Logs:    journalctl -u certd -f
Status:  certd status | jq .sync_info
EOF

echo -e "${GREEN}Validator info saved to /root/validator-info.txt${NC}"

