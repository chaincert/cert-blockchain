#!/bin/sh
# CERT Blockchain Docker Entrypoint Script
# Initializes the chain if not already initialized, then starts the node

set -ex

# Support both CHAIN_ID and CERT_CHAIN_ID environment variables
CHAIN_ID="${CERT_CHAIN_ID:-${CHAIN_ID:-951753}}"
MONIKER="${CERT_MONIKER:-${MONIKER:-cert-validator}}"
KEYRING_BACKEND="${KEYRING_BACKEND:-test}"
HOME_DIR="${HOME_DIR:-/root/.certd}"
DENOM="ucert"

# Check if chain is already initialized
if [ ! -f "$HOME_DIR/config/genesis.json" ]; then
    echo "================================================"
    echo "  CERT Blockchain Node Initialization"
    echo "================================================"
    echo "Chain ID: $CHAIN_ID"
    echo "Moniker: $MONIKER"
    echo "Home Directory: $HOME_DIR"
    echo ""

    # Initialize the chain
    echo "Step 1: Initializing chain..."
    certd init $MONIKER --chain-id $CHAIN_ID --home $HOME_DIR

    # Update genesis.json to use ucert as the bond denom (instead of default 'stake')
    echo "Step 1b: Updating staking params to use ucert..."
    sed -i 's/"bond_denom": "stake"/"bond_denom": "ucert"/' $HOME_DIR/config/genesis.json
    # Also update mint module to use ucert
    sed -i 's/"mint_denom": "stake"/"mint_denom": "ucert"/' $HOME_DIR/config/genesis.json

    # Configure minimum gas prices
    echo "Step 2: Configuring app.toml..."
    sed -i 's/minimum-gas-prices = ""/minimum-gas-prices = "0.0001ucert"/' $HOME_DIR/config/app.toml

    # Configure consensus parameters (Whitepaper Section 4.1)
    # Block time: ~2 seconds
    echo "Step 3: Configuring config.toml..."
    sed -i 's/timeout_commit = "5s"/timeout_commit = "2s"/' $HOME_DIR/config/config.toml
    sed -i 's/timeout_propose = "3s"/timeout_propose = "3s"/' $HOME_DIR/config/config.toml

    # Enable external RPC access
    sed -i 's/laddr = "tcp:\/\/127.0.0.1:26657"/laddr = "tcp:\/\/0.0.0.0:26657"/' $HOME_DIR/config/config.toml

    # Enable CORS for API
    sed -i 's/cors_allowed_origins = \[\]/cors_allowed_origins = ["*"]/' $HOME_DIR/config/config.toml

    # Enable API
    sed -i 's/enable = false/enable = true/' $HOME_DIR/config/app.toml
    sed -i 's/swagger = false/swagger = true/' $HOME_DIR/config/app.toml
    sed -i 's/address = "tcp:\/\/localhost:1317"/address = "tcp:\/\/0.0.0.0:1317"/' $HOME_DIR/config/app.toml

    # Enable gRPC
    sed -i 's/address = "localhost:9090"/address = "0.0.0.0:9090"/' $HOME_DIR/config/app.toml

    # Create validator account
    echo "Step 4: Creating validator key..."
    certd keys add validator --keyring-backend $KEYRING_BACKEND --home $HOME_DIR 2>&1 || true

    # Get validator address
    VALIDATOR_ADDRESS=$(certd keys show validator -a --keyring-backend $KEYRING_BACKEND --home $HOME_DIR)
    echo "Validator address: $VALIDATOR_ADDRESS"

    # Add genesis account with initial supply
    # 1 Billion CERT = 1,000,000,000,000,000 ucert
    echo "Step 5: Adding genesis account..."
    certd genesis add-genesis-account $VALIDATOR_ADDRESS 1000000000000000$DENOM --home $HOME_DIR --keyring-backend $KEYRING_BACKEND

    # Create gentx for validator
    # Minimum stake: 10,000 CERT (Whitepaper Section 10)
    echo "Step 6: Creating gentx..."
    certd genesis gentx validator 10000000000$DENOM \
      --chain-id $CHAIN_ID \
      --moniker $MONIKER \
      --keyring-backend $KEYRING_BACKEND \
      --home $HOME_DIR \
      --commission-rate 0.05 \
      --commission-max-rate 0.20 \
      --commission-max-change-rate 0.01

    # Collect gentxs
    echo "Step 7: Collecting gentxs..."
    certd genesis collect-gentxs --home $HOME_DIR

    # Validate genesis
    echo "Step 8: Validating genesis..."
    certd genesis validate --home $HOME_DIR

    echo ""
    echo "================================================"
    echo "  CERT Blockchain Node Initialized Successfully"
    echo "================================================"
fi

# Start the node
echo "Starting CERT Blockchain Node..."
exec certd start --home $HOME_DIR "$@"

