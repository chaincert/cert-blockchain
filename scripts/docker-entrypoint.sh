#!/bin/sh
# CERT Blockchain Docker Entrypoint Script
# Initializes the chain if not already initialized, then starts the node
# Tokenomics v2.1 - Total Supply: 1,000,000,000 CERT

set -ex

# Support both CHAIN_ID and CERT_CHAIN_ID environment variables
CHAIN_ID="${CERT_CHAIN_ID:-${CHAIN_ID:-cert-testnet-1}}"
MONIKER="${CERT_MONIKER:-${MONIKER:-cert-validator}}"
KEYRING_BACKEND="${KEYRING_BACKEND:-test}"
HOME_DIR="${HOME_DIR:-/root/.certd}"
DENOM="ucert"

# ============================================================
# TOKENOMICS v2.1 - Wallet Addresses (EVM format)
# Convert to Bech32 at runtime using certd debug addr
# ============================================================
# Treasury & Ecosystem (32%): 320,000,000 CERT = 320,000,000,000,000 ucert
TREASURY_ADDR_EVM="0xc68a92163f496ADCc7A8502fB2fdc7341fFdF589"
TREASURY_AMOUNT="320000000000000"

# Staking Rewards Pool (30%): 300,000,000 CERT = 300,000,000,000,000 ucert
STAKING_ADDR_EVM="0x5813612e4736cE42FC8582e0dBC7Ef51cAe906b9"
STAKING_AMOUNT="300000000000000"

# Private Sale & Liquidity (15%): 150,000,000 CERT = 150,000,000,000,000 ucert
PRIVATE_ADDR_EVM="0x0D756101183fe368C3364aBDe6Bf063CC3e7fcFD"
PRIVATE_AMOUNT="150000000000000"

# Core Team & Founders (15%): 150,000,000 CERT = 150,000,000,000,000 ucert
TEAM_ADDR_EVM="0x0D756101183fe368C3364aBDe6Bf063CC3e7fcFD"
TEAM_AMOUNT="150000000000000"

# Advisors & Future Hires (5%): 50,000,000 CERT = 50,000,000,000,000 ucert
ADVISORS_ADDR_EVM="0x0547711B2aC90Cead95E010B863a304602b40bF6"
ADVISORS_AMOUNT="50000000000000"

# Community Airdrop (3%): 30,000,000 CERT = 30,000,000,000,000 ucert
# NOTE: Using Treasury address until separate airdrop address provided
AIRDROP_ADDR_EVM="0xc68a92163f496ADCc7A8502fB2fdc7341fFdF589"
AIRDROP_AMOUNT="30000000000000"

# Check if we need to reset data (triggered by RESET_DATA=true env var)
if [ "${RESET_DATA:-false}" = "true" ]; then
    echo "RESET_DATA=true detected. Clearing blockchain data..."
    rm -rf $HOME_DIR/data/*
    rm -rf $HOME_DIR/config/genesis.json
    rm -rf $HOME_DIR/config/gentx
    rm -rf $HOME_DIR/keyring-test/*
fi

# Auto-detect corrupted data and reset if necessary
# Check if genesis exists but data is missing/corrupted (common after abrupt shutdown)
if [ -f "$HOME_DIR/config/genesis.json" ] && [ -d "$HOME_DIR/data" ]; then
    # If application.db is empty or missing key files, reset
    if [ ! -f "$HOME_DIR/data/application.db/CURRENT" ] || [ ! -s "$HOME_DIR/data/priv_validator_state.json" ]; then
        echo "WARNING: Detected corrupted or incomplete data directory. Resetting..."
        rm -rf $HOME_DIR/data/*
        rm -rf $HOME_DIR/config/genesis.json
        rm -rf $HOME_DIR/config/gentx
        rm -rf $HOME_DIR/keyring-test/*
    fi
fi

# Additional check: try a quick status query to detect store corruption
# This catches cases where files exist but are in an inconsistent state
if [ -f "$HOME_DIR/config/genesis.json" ] && [ -f "$HOME_DIR/data/application.db/CURRENT" ]; then
    echo "Validating chain state..."
    # Use timeout to prevent hanging; if query fails, data is corrupted
    if ! timeout 5 certd query block 1 --home $HOME_DIR > /dev/null 2>&1; then
        echo "WARNING: Chain state validation failed. Performing full reset..."
        rm -rf $HOME_DIR/data/*
        rm -rf $HOME_DIR/config/genesis.json
        rm -rf $HOME_DIR/config/gentx
        rm -rf $HOME_DIR/keyring-test/*
    fi
fi

# Check if chain is already initialized
if [ ! -f "$HOME_DIR/config/genesis.json" ]; then
    echo "================================================"
    echo "  CERT Blockchain Node Initialization"
    echo "  Tokenomics v2.1 - 1 Billion CERT"
    echo "================================================"
    echo "Chain ID: $CHAIN_ID"
    echo "Moniker: $MONIKER"
    echo "Home Directory: $HOME_DIR"
    echo ""

    # Initialize the chain
    echo "Step 1: Initializing chain..."
    certd init $MONIKER --chain-id $CHAIN_ID --home $HOME_DIR

    # Note: Genesis parameters are now set by NewDefaultGenesisState()
    # No manual sed patches needed for bond denom or inflation
    echo "Step 1b: Using custom genesis state with CERT parameters..."

    # Configure minimum gas prices
    echo "Step 2: Configuring app.toml..."
    sed -i 's/minimum-gas-prices = ""/minimum-gas-prices = "0.0001ucert"/' $HOME_DIR/config/app.toml

    # Disable IAVL fast node to avoid "version does not exist" errors on fresh chains
    # The fastnode feature has issues with IAVL v1.x when starting from genesis
    # Setting iavl-disable-fastnode = true disables fastnode (slightly slower queries but more stable)
    echo "Disabling IAVL FastNode for stability..."
    sed -i 's/iavl-disable-fastnode = false/iavl-disable-fastnode = true/' $HOME_DIR/config/app.toml
    grep "iavl-disable-fastnode" $HOME_DIR/config/app.toml

    # Set pruning to nothing to keep all state versions (required for REST API queries)
    awk '{gsub(/pruning = "default"/, "pruning = \"nothing\"")}1' $HOME_DIR/config/app.toml > $HOME_DIR/config/app.toml.tmp && mv $HOME_DIR/config/app.toml.tmp $HOME_DIR/config/app.toml

    # Set app-db-backend to goleveldb for proper store versioning
    sed -i 's/app-db-backend = ""/app-db-backend = "goleveldb"/' $HOME_DIR/config/app.toml

    # Set min-retain-blocks to 0 to keep all blocks (prevents pruning version issues)
    sed -i 's/min-retain-blocks = 0/min-retain-blocks = 0/' $HOME_DIR/config/app.toml

    # Disable state-sync snapshots to avoid IAVL v1.x bug with empty module stores
    # The bug causes "version does not exist" errors when exporting empty stores like crisis module
    # See: https://github.com/cosmos/iavl/issues/939
    sed -i 's/snapshot-interval = 100/snapshot-interval = 0/' $HOME_DIR/config/app.toml

    # Configure consensus parameters (Whitepaper Section 4.1)
    echo "Step 3: Configuring config.toml..."
    sed -i 's/timeout_commit = "5s"/timeout_commit = "2s"/' $HOME_DIR/config/config.toml
    sed -i 's/timeout_propose = "3s"/timeout_propose = "3s"/' $HOME_DIR/config/config.toml

    # Enable external RPC access
    sed -i 's/laddr = "tcp:\/\/127.0.0.1:26657"/laddr = "tcp:\/\/0.0.0.0:26657"/' $HOME_DIR/config/config.toml

    # Enable CORS for API
    sed -i 's/cors_allowed_origins = \[\]/cors_allowed_origins = ["*"]/' $HOME_DIR/config/config.toml

    # Enable API, gRPC, and gRPC-Web - using awk for more reliable section-based editing
    # Alpine's busybox sed has limited support for range patterns
    awk '
    /^\[api\]/ { in_api=1 }
    /^\[grpc\]$/ { in_api=0; in_grpc=1 }
    /^\[grpc-web\]/ { in_grpc=0; in_grpc_web=1 }
    /^\[/ && !/^\[api\]/ && !/^\[grpc\]$/ && !/^\[grpc-web\]/ { in_api=0; in_grpc=0; in_grpc_web=0 }

    in_api && /^enable = false/ { $0="enable = true" }
    in_api && /^swagger = false/ { $0="swagger = true" }
    in_api && /^address = "tcp:\/\/localhost:1317"/ { $0="address = \"tcp://0.0.0.0:1317\"" }

    in_grpc && /^enable = false/ { $0="enable = true" }
    in_grpc && /^address = "localhost:9090"/ { $0="address = \"0.0.0.0:9090\"" }

    in_grpc_web && /^enable = false/ { $0="enable = true" }

    { print }
    ' $HOME_DIR/config/app.toml > $HOME_DIR/config/app.toml.tmp && mv $HOME_DIR/config/app.toml.tmp $HOME_DIR/config/app.toml

    # Create validator account
    echo "Step 4: Creating validator key..."
    certd keys add validator --keyring-backend $KEYRING_BACKEND --home $HOME_DIR 2>&1 || true

    # Get validator address
    VALIDATOR_ADDRESS=$(certd keys show validator -a --keyring-backend $KEYRING_BACKEND --home $HOME_DIR)
    echo "Validator address: $VALIDATOR_ADDRESS"

	    # Fund validator account in genesis so gentx can succeed
	    echo "Step 5: Funding validator account in genesis..."
	    certd genesis add-genesis-account "$VALIDATOR_ADDRESS" 110000000000$DENOM \
	      --home $HOME_DIR \
	      --keyring-backend $KEYRING_BACKEND

	    # Ensure priv_validator_state.json exists (required by `certd genesis gentx`)
	    if [ ! -s "$HOME_DIR/data/priv_validator_state.json" ]; then
	        echo "priv_validator_state.json missing, creating default state file..."
	        mkdir -p "$HOME_DIR/data"
	        cat > "$HOME_DIR/data/priv_validator_state.json" <<EOF
{
  "height": "0",
  "round": 0,
  "step": 0
}
EOF
	    fi

	    # Create gentx for validator
	    echo "Step 6: Creating gentx (10,000 CERT stake)..."
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
    echo "  Tokenomics v2.1 Distribution Complete"
    echo "================================================"
    echo "Token Distribution:"
    echo "  Treasury:     320,000,000 CERT (32%)"
    echo "  Staking:      300,000,000 CERT (30%)"
    echo "  Team+Private: 300,000,000 CERT (30%)"
    echo "  Advisors:      50,000,000 CERT (5%)"
    echo "  Airdrop:       30,000,000 CERT (3%)*"
    echo "  --------------------------------"
    echo "  Total:      1,000,000,000 CERT"
    echo ""
    echo "* Airdrop held in Treasury until distribution"
    echo "================================================"
fi

# Start the node
echo "Starting CERT Blockchain Node..."
exec certd start --home $HOME_DIR "$@"

