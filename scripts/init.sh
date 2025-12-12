#!/bin/bash

# CERT Blockchain Initialization Script
# Per Whitepaper Section 4 and 12

set -e

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
certd init $MONIKER --chain-id $CHAIN_ID --home $HOME_DIR

# Update genesis with CERT token parameters (Whitepaper Section 5)
# Total Supply: 1 Billion CERT
# Base denomination: ucert (1,000,000 ucert = 1 CERT)

# Configure minimum gas prices
sed -i 's/minimum-gas-prices = ""/minimum-gas-prices = "0.0001ucert"/' $HOME_DIR/config/app.toml

# Configure consensus parameters (Whitepaper Section 4.1)
# Block time: ~2 seconds
sed -i 's/timeout_commit = "5s"/timeout_commit = "2s"/' $HOME_DIR/config/config.toml
sed -i 's/timeout_propose = "3s"/timeout_propose = "3s"/' $HOME_DIR/config/config.toml

# Enable JSON-RPC for EVM compatibility (Whitepaper Section 8)
cat >> $HOME_DIR/config/app.toml << EOF

###############################################################################
###                          JSON-RPC Configuration                         ###
###############################################################################

[json-rpc]
# Enable defines if the JSON-RPC server should be enabled.
enable = true

# Address defines the EVM RPC HTTP server address to bind to.
address = "0.0.0.0:8545"

# WsAddress defines the EVM RPC WebSocket server address to bind to.
ws-address = "0.0.0.0:8546"

# API defines a list of JSON-RPC namespaces that should be enabled
api = "eth,txpool,personal,net,debug,web3"

# GasCap sets a cap on gas that can be used in eth_call/estimateGas
gas-cap = 25000000

# EVMTimeout is the global timeout for eth_call.
evm-timeout = "5s"

# TxFeeCap is the global tx-fee cap for send transactions
txfee-cap = 1

# FilterCap sets the global cap for total number of filters
filter-cap = 200

# FeeHistoryCap sets the global cap for total number of blocks for feeHistory
feehistory-cap = 100

# LogsCap defines the max number of results can be returned from eth_getLogs
logs-cap = 10000

# BlockRangeCap defines the max block range for eth_getLogs
block-range-cap = 10000

# HTTPTimeout is the timeout for HTTP requests
http-timeout = "30s"

# HTTPIdleTimeout is the idle timeout for HTTP connections
http-idle-timeout = "120s"

# MaxOpenConnections sets the maximum number of simultaneous connections
max-open-connections = 0
EOF

# Create validator account
certd keys add validator --keyring-backend $KEYRING_BACKEND --home $HOME_DIR

# Get validator address
VALIDATOR_ADDRESS=$(certd keys show validator -a --keyring-backend $KEYRING_BACKEND --home $HOME_DIR)

echo ""
echo "Validator address: $VALIDATOR_ADDRESS"

# Add genesis account with initial supply
# 1 Billion CERT = 1,000,000,000,000,000 ucert
certd genesis add-genesis-account $VALIDATOR_ADDRESS 1000000000000000$DENOM --home $HOME_DIR

# Create gentx for validator
# Minimum stake: 10,000 CERT (Whitepaper Section 10)
certd genesis gentx validator 10000000000$DENOM \
  --chain-id $CHAIN_ID \
  --moniker $MONIKER \
  --keyring-backend $KEYRING_BACKEND \
  --home $HOME_DIR \
  --commission-rate 0.05 \
  --commission-max-rate 0.20 \
  --commission-max-change-rate 0.01

# Collect gentxs
certd genesis collect-gentxs --home $HOME_DIR

# Validate genesis
certd genesis validate --home $HOME_DIR

echo ""
echo "================================================"
echo "  CERT Blockchain Node Initialized Successfully"
echo "================================================"
echo ""
echo "To start the node, run:"
echo "  certd start --home $HOME_DIR"
echo ""
echo "JSON-RPC endpoint will be available at:"
echo "  HTTP: http://localhost:8545"
echo "  WebSocket: ws://localhost:8546"
echo ""
echo "Cosmos REST API will be available at:"
echo "  http://localhost:1317"
echo ""

