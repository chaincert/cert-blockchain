# CERT Blockchain - Validator Quick Start

**For detailed instructions, see [VALIDATOR_SETUP_GUIDE.md](VALIDATOR_SETUP_GUIDE.md)**

---

## 🚀 Quick Setup (5 Minutes)

### On New Validator Server

```bash
# 1. Clone repository
cd /opt
git clone https://github.com/chaincertify/cert-blockchain.git
cd cert-blockchain

# 2. Install certd
make install

# 3. Run automated setup
./scripts/add-validator.sh
```

The script will guide you through:
- ✅ Node initialization
- ✅ Key generation
- ✅ Genesis file setup
- ✅ Peer configuration
- ✅ Node startup

---

## 📋 Manual Setup (Step-by-Step)

### 1. Initialize Node

```bash
export MONIKER="validator-2"  # Change for each validator
certd init $MONIKER --chain-id cert-testnet-1
```

### 2. Create Validator Key

```bash
certd keys add validator --keyring-backend test
# ⚠️ SAVE THE MNEMONIC!
```

### 3. Get Genesis File

**On primary validator (172.239.32.74):**
```bash
cat /opt/cert-blockchain/data/certd/config/genesis.json
```

**On new validator:**
```bash
# Copy genesis.json to ~/.certd/config/genesis.json
certd genesis validate
```

### 4. Configure Peers

**Get node ID from primary validator:**
```bash
# On 172.239.32.74
certd tendermint show-node-id --home /opt/cert-blockchain/data/certd
```

**On new validator:**
```bash
nano ~/.certd/config/config.toml
# Update: persistent_peers = "<NODE_ID>@172.239.32.74:26656"
```

### 5. Fund Validator Account

**On primary validator:**
```bash
certd tx bank send validator <NEW_VALIDATOR_ADDRESS> 15000000000ucert \
  --chain-id cert-testnet-1 \
  --keyring-backend test \
  --home /opt/cert-blockchain/data/certd \
  --fees 10000ucert \
  --yes
```

### 6. Start Node

```bash
certd start
# Wait for sync to complete
```

### 7. Create Validator

```bash
certd tx staking create-validator \
  --amount=10000000000ucert \
  --pubkey=$(certd tendermint show-validator) \
  --moniker="validator-2" \
  --chain-id=cert-testnet-1 \
  --commission-rate="0.05" \
  --commission-max-rate="0.20" \
  --commission-max-change-rate="0.01" \
  --min-self-delegation="1" \
  --from=validator \
  --keyring-backend=test \
  --fees=10000ucert \
  --yes
```

### 8. Verify

```bash
certd query staking validators
```

---

## 🔍 Quick Commands

### Check Node Status
```bash
certd status | jq .sync_info
```

### Check Validator Info
```bash
certd query staking validators
```

### Check Balance
```bash
certd query bank balances $(certd keys show validator -a --keyring-backend test)
```

### Unjail Validator
```bash
certd tx slashing unjail --from validator --fees 10000ucert --yes
```

---

## 📊 Network Info

| Parameter | Value |
|-----------|-------|
| **Chain ID** | cert-testnet-1 |
| **Denom** | ucert (1 CERT = 1,000,000 ucert) |
| **Min Stake** | 10,000 CERT (10000000000 ucert) |
| **Block Time** | ~2 seconds |
| **Unbonding** | 21 days |
| **Primary Validator** | 172.239.32.74:26656 |

---

## ⚠️ Important

### Backup These Files
```bash
~/.certd/config/priv_validator_key.json  # Validator signing key
~/.certd/config/node_key.json            # P2P identity
# Mnemonic phrase (from key creation)
```

### Security
```bash
# Firewall
ufw allow 22/tcp    # SSH
ufw allow 26656/tcp # P2P
ufw enable

# Backup keys offline
cp ~/.certd/config/priv_validator_key.json ~/validator_backup.json
```

---

## 🆘 Troubleshooting

### Node Won't Sync
```bash
# Check peers
certd status | jq .sync_info.peers

# Reset and resync
certd tendermint unsafe-reset-all
certd start
```

### Validator Not Signing
```bash
# Check if jailed
certd query staking validator <VALIDATOR_ADDRESS>

# Unjail
certd tx slashing unjail --from validator --fees 10000ucert --yes
```

### Need More Funds
```bash
# Request from primary validator
# Contact: dev@c3rt.org
```

---

## 📚 Resources

- **Full Guide:** [VALIDATOR_SETUP_GUIDE.md](VALIDATOR_SETUP_GUIDE.md)
- **Deployment:** [DEPLOYMENT.md](DEPLOYMENT.md)
- **Website:** https://c3rt.org
- **Docs:** https://c3rt.org/docs
- **Support:** dev@c3rt.org

