#!/bin/bash

# CERT Blockchain Initialization Script
# Per Whitepaper Section 4 and 12

set -e

# Allow overriding the daemon binary; default to whatever is on PATH.
DAEMON_BIN="${DAEMON_BIN:-certd}"

CHAIN_ID="${CHAIN_ID:-cert-testnet-1}"
MONIKER="${MONIKER:-cert-validator}"
KEYRING_BACKEND="${KEYRING_BACKEND:-test}"
HOME_DIR="${HOME_DIR:-$HOME/.certd}"
DENOM="ucert"

# Remove existing configuration
rm -rf $HOME_DIR

echo "================================================"
echo "  CERT Blockchain Node Initialization"
echo "================================================"
echo "Chain ID: $CHAIN_ID"
echo "Moniker: $MONIKER"
echo "Home Directory: $HOME_DIR"
echo ""

# Initialize the chain
"$DAEMON_BIN" init "$MONIKER" --chain-id "$CHAIN_ID" --home "$HOME_DIR"

# Update genesis with CERT token parameters (Whitepaper Section 5)
# Total Supply: 1 Billion CERT
# Base denomination: ucert (1,000,000 ucert = 1 CERT)

# Configure minimum gas prices
sed -i 's/minimum-gas-prices = ""/minimum-gas-prices = "0.0001ucert"/' $HOME_DIR/config/app.toml

# Configure consensus parameters (Whitepaper Section 4.1)
# Block time: ~2 seconds
sed -i 's/timeout_commit = "5s"/timeout_commit = "2s"/' $HOME_DIR/config/config.toml
sed -i 's/timeout_propose = "3s"/timeout_propose = "3s"/' $HOME_DIR/config/config.toml

# JSON-RPC for EVM compatibility (Whitepaper Section 8) is now configured
# directly via the app config template in cmd/certd/cmd/root.go (initAppConfig).
# No additional manual app.toml patching is required here.

# Create validator account
"$DAEMON_BIN" keys add validator --keyring-backend "$KEYRING_BACKEND" --home "$HOME_DIR"

# Get validator address
VALIDATOR_ADDRESS=$("$DAEMON_BIN" keys show validator -a --keyring-backend "$KEYRING_BACKEND" --home "$HOME_DIR")

echo ""
echo "Validator address: $VALIDATOR_ADDRESS"

# Fund validator account for self-delegation and fees
# 110,000 CERT = 110,000 * 1,000,000 ucert = 110000000000ucert
"$DAEMON_BIN" genesis add-genesis-account "$VALIDATOR_ADDRESS" 110000000000"$DENOM" \
	--home "$HOME_DIR"

echo "Using custom genesis state with pre-configured accounts plus funded validator..."

# Create gentx for validator
# Minimum stake: 10,000 CERT (Whitepaper Section 10)
"$DAEMON_BIN" genesis gentx validator 10000000000"$DENOM" \
  --chain-id "$CHAIN_ID" \
  --moniker "$MONIKER" \
  --keyring-backend "$KEYRING_BACKEND" \
  --home "$HOME_DIR" \
  --commission-rate 0.05 \
  --commission-max-rate 0.20 \
  --commission-max-change-rate 0.01

# Collect gentxs
"$DAEMON_BIN" genesis collect-gentxs --home "$HOME_DIR"

# Validate genesis
"$DAEMON_BIN" genesis validate --home "$HOME_DIR"

echo ""
echo "================================================"
echo "  CERT Blockchain Node Initialized Successfully"
echo "================================================"
echo ""
echo "To start the node, run:"
echo "  $DAEMON_BIN start --home $HOME_DIR"
echo ""
echo "JSON-RPC endpoint will be available at:"
echo "  HTTP: http://localhost:8545"
echo "  WebSocket: ws://localhost:8546"
echo ""
echo "Cosmos REST API will be available at:"
echo "  http://localhost:1317"
echo ""

