# CERT Blockchain Validator Setup Guide

**Chain ID:** cert-testnet-1  
**Minimum Stake:** 10,000 CERT (10000000000 ucert)  
**Current Validators:** 1  
**Target:** Add 1-2 more validators

---

## Prerequisites

### Hardware Requirements
- **CPU:** 4+ cores
- **RAM:** 8GB minimum, 16GB recommended
- **Storage:** 100GB+ SSD
- **Network:** 100 Mbps+ connection
- **OS:** Ubuntu 22.04 LTS (recommended)

### Software Requirements
- Docker & Docker Compose
- Open ports: 26656 (P2P), 26657 (RPC)

---

## Option 1: Add Validator to Existing Network (Recommended)

This method adds a new validator to the running testnet without restarting the chain.

### Step 1: Prepare New Validator Server

```bash
# SSH into new validator server
ssh root@<NEW_VALIDATOR_IP>

# Install dependencies
apt update && apt install -y docker.io docker-compose git

# Clone repository
cd /opt
git clone https://github.com/chaincertify/cert-blockchain.git
cd cert-blockchain
```

### Step 2: Build the Validator Node

```bash
# Build the certd binary
docker build -t cert-blockchain:latest .

# Or build locally
make install
```

### Step 3: Initialize New Validator

```bash
# Set environment variables
export CHAIN_ID="cert-testnet-1"
export MONIKER="validator-2"  # Change this for each validator
export HOME_DIR="$HOME/.certd"

# Initialize node
certd init $MONIKER --chain-id $CHAIN_ID --home $HOME_DIR

# Create validator key
certd keys add validator --keyring-backend test --home $HOME_DIR

# Save the mnemonic and address!
# Example output:
# - address: cert1abc123...
# - mnemonic: word1 word2 word3 ...
```

### Step 4: Get Genesis File from Existing Validator

```bash
# Copy genesis.json from the primary validator
scp root@172.239.32.74:/opt/cert-blockchain/data/certd/config/genesis.json \
    $HOME_DIR/config/genesis.json

# Verify genesis file
certd genesis validate --home $HOME_DIR
```

### Step 5: Configure Persistent Peers

```bash
# Get node ID from primary validator
# On primary validator (172.239.32.74):
certd tendermint show-node-id --home /opt/cert-blockchain/data/certd

# Example output: a1b2c3d4e5f6...

# On new validator, edit config.toml
nano $HOME_DIR/config/config.toml

# Find and update:
persistent_peers = "<NODE_ID>@172.239.32.74:26656"

# Example:
# persistent_peers = "a1b2c3d4e5f6@172.239.32.74:26656"
```

### Step 6: Fund Validator Account

**On the primary validator server (172.239.32.74):**

```bash
# Get new validator address (from Step 3)
NEW_VALIDATOR_ADDRESS="cert1abc123..."  # Replace with actual address

# Send 15,000 CERT (10k stake + 5k for fees)
certd tx bank send validator $NEW_VALIDATOR_ADDRESS 15000000000ucert \
  --chain-id cert-testnet-1 \
  --keyring-backend test \
  --home /opt/cert-blockchain/data/certd \
  --fees 10000ucert \
  --yes

# Verify balance
certd query bank balances $NEW_VALIDATOR_ADDRESS \
  --node tcp://localhost:26657
```

### Step 7: Start the New Validator Node

```bash
# On new validator server
certd start --home $HOME_DIR

# Or run in background
nohup certd start --home $HOME_DIR > certd.log 2>&1 &

# Check logs
tail -f certd.log

# Wait for node to sync (check latest block)
certd status --node tcp://localhost:26657 | jq .sync_info
```

### Step 8: Create Validator Transaction

**After node is fully synced:**

```bash
# Create validator
certd tx staking create-validator \
  --amount=10000000000ucert \
  --pubkey=$(certd tendermint show-validator --home $HOME_DIR) \
  --moniker="validator-2" \
  --chain-id=cert-testnet-1 \
  --commission-rate="0.05" \
  --commission-max-rate="0.20" \
  --commission-max-change-rate="0.01" \
  --min-self-delegation="1" \
  --from=validator \
  --keyring-backend=test \
  --home=$HOME_DIR \
  --fees=10000ucert \
  --yes

# Verify validator is active
certd query staking validators --node tcp://localhost:26657
```

### Step 9: Verify Validator is Signing Blocks

```bash
# Check validator status
certd query staking validator $(certd keys show validator --bech val -a \
  --keyring-backend test --home $HOME_DIR) \
  --node tcp://localhost:26657

# Check if signing blocks
certd query slashing signing-info $(certd tendermint show-validator --home $HOME_DIR) \
  --node tcp://localhost:26657
```

---

## Option 2: Add Validator at Genesis (Requires Chain Restart)

**⚠️ WARNING:** This method requires restarting the entire chain and will cause downtime.

### When to Use
- Setting up a new testnet from scratch
- Adding multiple validators before launch
- Testing validator setup in development

### Steps

1. **Stop all existing validators**
2. **Collect gentx files from all validators**
3. **Merge genesis files**
4. **Restart all validators simultaneously**

See `scripts/init.sh` for genesis validator setup.

---

## Validator Management

### Check Validator Status

```bash
# Get validator address
certd keys show validator --bech val -a --keyring-backend test

# Query validator info
certd query staking validator <VALIDATOR_ADDRESS>

# Check signing info
certd query slashing signing-info $(certd tendermint show-validator)
```

### Delegate More Stake

```bash
certd tx staking delegate <VALIDATOR_ADDRESS> 5000000000ucert \
  --from validator \
  --chain-id cert-testnet-1 \
  --keyring-backend test \
  --fees 10000ucert \
  --yes
```

### Unjail Validator (if slashed)

```bash
certd tx slashing unjail \
  --from validator \
  --chain-id cert-testnet-1 \
  --keyring-backend test \
  --fees 10000ucert \
  --yes
```

---

## Security Best Practices

### 1. Secure Private Keys

```bash
# Backup validator key
cp $HOME_DIR/config/priv_validator_key.json ~/validator_key_backup.json

# Store securely offline!
# Never share this file
```

### 2. Configure Firewall

```bash
# Allow only necessary ports
ufw allow 22/tcp    # SSH
ufw allow 26656/tcp # P2P
ufw allow 26657/tcp # RPC (optional, for monitoring)
ufw enable
```

### 3. Set Up Monitoring

```bash
# Install Prometheus node exporter
docker run -d --name node-exporter \
  -p 9100:9100 \
  prom/node-exporter

# Monitor validator uptime
# Use Grafana + Prometheus for dashboards
```

### 4. Enable Sentry Nodes (Production)

For production, use sentry node architecture:
- Public sentry nodes handle P2P connections
- Validator node only connects to sentries
- Protects validator from DDoS attacks

---

## Troubleshooting

### Node Won't Sync

```bash
# Check peers
certd status | jq .sync_info.peers

# Add more peers
nano $HOME_DIR/config/config.toml
# Update persistent_peers

# Reset and resync
certd tendermint unsafe-reset-all --home $HOME_DIR
certd start --home $HOME_DIR
```

### Validator Not Signing

```bash
# Check if jailed
certd query staking validator <VALIDATOR_ADDRESS>

# Check signing info
certd query slashing signing-info $(certd tendermint show-validator)

# Unjail if needed
certd tx slashing unjail --from validator --fees 10000ucert --yes
```

### Insufficient Funds

```bash
# Check balance
certd query bank balances $(certd keys show validator -a --keyring-backend test)

# Request funds from faucet or existing validator
```

---

## Quick Reference

### Important Commands

```bash
# Node status
certd status

# Validator info
certd query staking validators

# Account balance
certd query bank balances <ADDRESS>

# Send tokens
certd tx bank send <FROM> <TO> <AMOUNT>ucert --fees 10000ucert

# Delegate
certd tx staking delegate <VALIDATOR> <AMOUNT>ucert --from <KEY>

# Check logs
journalctl -u certd -f  # If running as systemd service
tail -f certd.log       # If running in background
```

### Network Information

- **Chain ID:** cert-testnet-1
- **Denom:** ucert (1 CERT = 1,000,000 ucert)
- **Block Time:** ~2 seconds
- **Unbonding Period:** 21 days
- **Min Stake:** 10,000 CERT
- **Max Validators:** 80

---

## Next Steps

After setting up your validator:

1. ✅ Monitor validator uptime (should be >99%)
2. ✅ Set up automated backups of validator keys
3. ✅ Configure monitoring and alerting
4. ✅ Join validator community channels
5. ✅ Participate in governance proposals

---

## Support

- **Documentation:** https://c3rt.org/docs
- **Discord:** https://discord.gg/cert
- **Email:** dev@c3rt.org

