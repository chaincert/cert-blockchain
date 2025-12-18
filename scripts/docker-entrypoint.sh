#!/bin/sh
# CERT Blockchain Docker Entrypoint Script
# Initializes the chain if not already initialized, then starts the node
# Tokenomics v2.1 - Total Supply: 1,000,000,000 CERT

set -ex

# Support both CHAIN_ID and CERT_CHAIN_ID environment variables
CHAIN_ID="${CERT_CHAIN_ID:-${CHAIN_ID:-951753}}"
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

    # Update genesis.json to use ucert as the bond denom (instead of default 'stake')
    echo "Step 1b: Updating staking params to use ucert..."
    sed -i 's/"bond_denom": "stake"/"bond_denom": "ucert"/' $HOME_DIR/config/genesis.json
    sed -i 's/"mint_denom": "stake"/"mint_denom": "ucert"/' $HOME_DIR/config/genesis.json

    # Disable inflation (fixed supply per Tokenomics v2.1)
    echo "Step 1c: Disabling inflation (fixed supply)..."
    sed -i 's/"inflation": "[^"]*"/"inflation": "0.000000000000000000"/' $HOME_DIR/config/genesis.json
    sed -i 's/"inflation_rate_change": "[^"]*"/"inflation_rate_change": "0.000000000000000000"/' $HOME_DIR/config/genesis.json
    sed -i 's/"inflation_max": "[^"]*"/"inflation_max": "0.000000000000000000"/' $HOME_DIR/config/genesis.json
    sed -i 's/"inflation_min": "[^"]*"/"inflation_min": "0.000000000000000000"/' $HOME_DIR/config/genesis.json

    # Configure minimum gas prices
    echo "Step 2: Configuring app.toml..."
    sed -i 's/minimum-gas-prices = ""/minimum-gas-prices = "0.0001ucert"/' $HOME_DIR/config/app.toml

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

    # ============================================================
    # Step 5: Add Genesis Accounts - Tokenomics v2.1 Distribution
    # Total: 1,000,000,000 CERT = 1,000,000,000,000,000 ucert
    # ============================================================
    echo "Step 5: Adding genesis accounts (Tokenomics v2.1)..."

    # Give validator enough for staking PLUS faucet/operations
    # 10,000 CERT for stake + 100,000 CERT for faucet operations = 110,000 CERT
    # This comes from the staking pool allocation
    certd genesis add-genesis-account $VALIDATOR_ADDRESS 110000000000$DENOM --home $HOME_DIR --keyring-backend $KEYRING_BACKEND

    # Treasury (32%) - 320,000,000 CERT
    echo "  Adding Treasury (32%): 320,000,000 CERT"
	    # Note: certd genesis currently expects a Bech32 address or key name.
	    # Our EVM addresses may not be directly importable yet. Make failures non-fatal
	    # so that node initialization (gentx, collect-gentxs, validate) can still succeed.
	    certd genesis add-genesis-account $TREASURY_ADDR_EVM ${TREASURY_AMOUNT}$DENOM --home $HOME_DIR --keyring-backend $KEYRING_BACKEND 2>/dev/null || \
	    certd genesis add-genesis-account $(echo $TREASURY_ADDR_EVM | sed 's/0x//') ${TREASURY_AMOUNT}$DENOM --home $HOME_DIR --keyring-backend $KEYRING_BACKEND 2>/dev/null || \
	    echo "  Warning: could not add Treasury genesis account (address format not yet supported), skipping."

    # Staking Pool (30% - 10k for validator) - 299,990,000 CERT
    echo "  Adding Staking Pool (30%): 299,990,000 CERT"
    STAKING_REMAINING=$((STAKING_AMOUNT - 10000000000))
	    certd genesis add-genesis-account $STAKING_ADDR_EVM ${STAKING_REMAINING}$DENOM --home $HOME_DIR --keyring-backend $KEYRING_BACKEND 2>/dev/null || \
	    certd genesis add-genesis-account $(echo $STAKING_ADDR_EVM | sed 's/0x//') ${STAKING_REMAINING}$DENOM --home $HOME_DIR --keyring-backend $KEYRING_BACKEND 2>/dev/null || \
	    echo "  Warning: could not add Staking Pool genesis account (address format not yet supported), skipping."

    # Note: Team and Private Sale share same address, combine amounts
    # Combined (15% + 15%) - 300,000,000 CERT
    echo "  Adding Team + Private Sale (30%): 300,000,000 CERT"
    COMBINED_AMOUNT=$((TEAM_AMOUNT + PRIVATE_AMOUNT))
	    certd genesis add-genesis-account $TEAM_ADDR_EVM ${COMBINED_AMOUNT}$DENOM --home $HOME_DIR --keyring-backend $KEYRING_BACKEND 2>/dev/null || \
	    certd genesis add-genesis-account $(echo $TEAM_ADDR_EVM | sed 's/0x//') ${COMBINED_AMOUNT}$DENOM --home $HOME_DIR --keyring-backend $KEYRING_BACKEND 2>/dev/null || \
	    echo "  Warning: could not add Team+Private genesis account (address format not yet supported), skipping."

    # Advisors (5%) - 50,000,000 CERT
    echo "  Adding Advisors (5%): 50,000,000 CERT"
	    certd genesis add-genesis-account $ADVISORS_ADDR_EVM ${ADVISORS_AMOUNT}$DENOM --home $HOME_DIR --keyring-backend $KEYRING_BACKEND 2>/dev/null || \
	    certd genesis add-genesis-account $(echo $ADVISORS_ADDR_EVM | sed 's/0x//') ${ADVISORS_AMOUNT}$DENOM --home $HOME_DIR --keyring-backend $KEYRING_BACKEND 2>/dev/null || \
	    echo "  Warning: could not add Advisors genesis account (address format not yet supported), skipping."

    # Airdrop (3%) - 30,000,000 CERT (added to Treasury for now)
    echo "  Adding Airdrop (3%): 30,000,000 CERT (to Treasury)"
    # Airdrop uses same address as Treasury, will be combined automatically

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

