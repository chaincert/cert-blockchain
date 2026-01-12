#!/bin/bash

# CERT Blockchain - Add New Validator Script
# This script automates the process of adding a new validator to the testnet

set -e

echo "================================================"
echo "  CERT Blockchain - Add New Validator"
echo "================================================"
echo ""

# Configuration
CHAIN_ID="cert-testnet-1"
KEYRING_BACKEND="test"
HOME_DIR="$HOME/.certd"
DENOM="ucert"
PRIMARY_VALIDATOR_IP="172.239.32.74"

# Prompt for validator information
read -p "Enter validator moniker (e.g., validator-2): " MONIKER
read -p "Enter primary validator node ID (run 'certd tendermint show-node-id' on primary): " PRIMARY_NODE_ID

echo ""
echo "Configuration:"
echo "  Chain ID: $CHAIN_ID"
echo "  Moniker: $MONIKER"
echo "  Home Directory: $HOME_DIR"
echo "  Primary Validator: $PRIMARY_NODE_ID@$PRIMARY_VALIDATOR_IP:26656"
echo ""
read -p "Continue? (y/n): " CONFIRM

if [ "$CONFIRM" != "y" ]; then
    echo "Aborted."
    exit 1
fi

# Step 1: Check if certd is installed
echo ""
echo "Step 1: Checking certd installation..."
if ! command -v certd &> /dev/null; then
    echo "❌ certd not found. Please install it first:"
    echo "   cd /opt/cert-blockchain && make install"
    exit 1
fi
echo "✅ certd found: $(which certd)"

# Step 2: Initialize node
echo ""
echo "Step 2: Initializing validator node..."
if [ -d "$HOME_DIR" ]; then
    read -p "⚠️  $HOME_DIR already exists. Remove it? (y/n): " REMOVE
    if [ "$REMOVE" = "y" ]; then
        rm -rf "$HOME_DIR"
    else
        echo "Aborted."
        exit 1
    fi
fi

certd init "$MONIKER" --chain-id "$CHAIN_ID" --home "$HOME_DIR"
echo "✅ Node initialized"

# Step 3: Create validator key
echo ""
echo "Step 3: Creating validator key..."
echo "⚠️  IMPORTANT: Save the mnemonic phrase shown below!"
echo ""
certd keys add validator --keyring-backend "$KEYRING_BACKEND" --home "$HOME_DIR"

VALIDATOR_ADDRESS=$(certd keys show validator -a --keyring-backend "$KEYRING_BACKEND" --home "$HOME_DIR")
echo ""
echo "✅ Validator address: $VALIDATOR_ADDRESS"
echo ""
echo "⚠️  BACKUP YOUR KEYS:"
echo "   Mnemonic: (shown above)"
echo "   Private key: $HOME_DIR/config/priv_validator_key.json"
echo ""
read -p "Press Enter after you've saved the mnemonic..."

# Step 4: Get genesis file
echo ""
echo "Step 4: Fetching genesis file from primary validator..."
echo "Run this command on the primary validator to copy genesis.json:"
echo ""
echo "  scp /opt/cert-blockchain/data/certd/config/genesis.json root@$(hostname -I | awk '{print $1}'):$HOME_DIR/config/genesis.json"
echo ""
read -p "Press Enter after genesis.json has been copied..."

if [ ! -f "$HOME_DIR/config/genesis.json" ]; then
    echo "❌ genesis.json not found at $HOME_DIR/config/genesis.json"
    exit 1
fi

certd genesis validate --home "$HOME_DIR"
echo "✅ Genesis file validated"

# Step 5: Configure persistent peers
echo ""
echo "Step 5: Configuring persistent peers..."
sed -i "s/persistent_peers = \"\"/persistent_peers = \"$PRIMARY_NODE_ID@$PRIMARY_VALIDATOR_IP:26656\"/" "$HOME_DIR/config/config.toml"
echo "✅ Persistent peers configured"

# Step 6: Configure minimum gas prices
echo ""
echo "Step 6: Configuring minimum gas prices..."
sed -i 's/minimum-gas-prices = ""/minimum-gas-prices = "0.0001ucert"/' "$HOME_DIR/config/app.toml"
echo "✅ Gas prices configured"

# Step 7: Fund validator account
echo ""
echo "Step 7: Funding validator account..."
echo "Run this command on the primary validator (172.239.32.74):"
echo ""
echo "  certd tx bank send validator $VALIDATOR_ADDRESS 15000000000ucert \\"
echo "    --chain-id cert-testnet-1 \\"
echo "    --keyring-backend test \\"
echo "    --home /opt/cert-blockchain/data/certd \\"
echo "    --fees 10000ucert \\"
echo "    --yes"
echo ""
read -p "Press Enter after funds have been sent..."

# Step 8: Start node
echo ""
echo "Step 8: Starting validator node..."
echo "The node will now start syncing. This may take a while."
echo ""
read -p "Start node now? (y/n): " START_NOW

if [ "$START_NOW" = "y" ]; then
    echo "Starting certd in background..."
    nohup certd start --home "$HOME_DIR" > certd.log 2>&1 &
    CERTD_PID=$!
    echo "✅ certd started (PID: $CERTD_PID)"
    echo "   Logs: tail -f certd.log"
    echo ""
    echo "Waiting for node to start..."
    sleep 5
    
    echo ""
    echo "Checking sync status..."
    certd status --node tcp://localhost:26657 2>/dev/null | jq .sync_info || echo "Node starting..."
fi

# Step 9: Create validator transaction
echo ""
echo "================================================"
echo "  Next Steps"
echo "================================================"
echo ""
echo "1. Wait for node to fully sync:"
echo "   certd status --node tcp://localhost:26657 | jq .sync_info.catching_up"
echo "   (should show 'false' when synced)"
echo ""
echo "2. Create validator:"
echo ""
echo "   certd tx staking create-validator \\"
echo "     --amount=10000000000ucert \\"
echo "     --pubkey=\$(certd tendermint show-validator --home $HOME_DIR) \\"
echo "     --moniker=\"$MONIKER\" \\"
echo "     --chain-id=cert-testnet-1 \\"
echo "     --commission-rate=\"0.05\" \\"
echo "     --commission-max-rate=\"0.20\" \\"
echo "     --commission-max-change-rate=\"0.01\" \\"
echo "     --min-self-delegation=\"1\" \\"
echo "     --from=validator \\"
echo "     --keyring-backend=test \\"
echo "     --home=$HOME_DIR \\"
echo "     --fees=10000ucert \\"
echo "     --yes"
echo ""
echo "3. Verify validator is active:"
echo "   certd query staking validators --node tcp://localhost:26657"
echo ""
echo "4. Check if signing blocks:"
echo "   certd query slashing signing-info \$(certd tendermint show-validator --home $HOME_DIR) \\"
echo "     --node tcp://localhost:26657"
echo ""
echo "================================================"
echo "  Important Files"
echo "================================================"
echo ""
echo "Validator Address: $VALIDATOR_ADDRESS"
echo "Home Directory: $HOME_DIR"
echo "Private Key: $HOME_DIR/config/priv_validator_key.json"
echo "Node Key: $HOME_DIR/config/node_key.json"
echo "Logs: certd.log"
echo ""
echo "⚠️  BACKUP YOUR KEYS SECURELY!"
echo ""

