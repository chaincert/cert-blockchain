# CERT Blockchain Firewall Configuration Guide

**Server:** 172.239.32.74 (c3rt.org)  
**Date:** 2025-12-27

## Current Status

### ✅ Currently Open Ports
- **80** (HTTP) - ✅ Correct
- **443** (HTTPS) - ✅ Correct  
- **3000** (API) - ⚠️ **SHOULD BE CLOSED**

### ❌ Currently Closed Ports
- **26656** (P2P) - ❌ **SHOULD BE OPEN**
- **5432** (PostgreSQL) - ✅ Correct (internal only)
- **26657** (RPC) - ✅ Correct (nginx proxy only)

---

## Required Configuration

### Ports That MUST Be Open
1. **22** - SSH (admin access)
2. **80** - HTTP (redirects to HTTPS)
3. **443** - HTTPS (nginx reverse proxy)
4. **26656** - Tendermint P2P (validator communication)

### Ports That MUST Be Blocked
1. **3000** - Direct API access (use nginx proxy)
2. **5432** - PostgreSQL (internal only)
3. **26657** - Direct RPC access (use nginx proxy)
4. **8080** - Direct IPFS access (use nginx proxy)
5. **8545** - Direct EVM RPC (use nginx proxy)
6. **8546** - Direct EVM WebSocket (use nginx proxy)
7. **1317** - Direct Cosmos REST (use nginx proxy)
8. **9090** - Direct gRPC (use nginx proxy)

---

## Automated Setup (Recommended)

### Option 1: Run from Local Machine

```bash
# From your local machine
cd /opt/cert-blockchain/scripts

# Apply firewall configuration to remote server
./apply-firewall-remote.sh
```

This will:
1. Copy the firewall script to the server
2. Install and configure fail2ban
3. Configure UFW with proper rules
4. Test the configuration

---

## Manual Setup (If SSH Access Available)

### Step 1: SSH to Server

```bash
ssh root@172.239.32.74
```

### Step 2: Install fail2ban

```bash
apt update
apt install -y fail2ban

# Configure fail2ban for SSH protection
cat > /etc/fail2ban/jail.local << 'EOF'
[DEFAULT]
bantime = 3600
findtime = 600
maxretry = 5

[sshd]
enabled = true
port = 22
maxretry = 3
bantime = 7200
EOF

systemctl enable fail2ban
systemctl restart fail2ban
```

### Step 3: Configure UFW Firewall

```bash
# Reset UFW (clean slate)
ufw --force reset

# Set default policies
ufw default deny incoming
ufw default allow outgoing

# Allow essential services
ufw allow 22/tcp comment 'SSH'
ufw allow 80/tcp comment 'HTTP'
ufw allow 443/tcp comment 'HTTPS'
ufw allow 26656/tcp comment 'P2P'

# Block direct access to internal services
ufw deny 3000/tcp comment 'Block direct API'
ufw deny 5432/tcp comment 'Block PostgreSQL'
ufw deny 26657/tcp comment 'Block direct RPC'
ufw deny 8080/tcp comment 'Block direct IPFS'
ufw deny 8545/tcp comment 'Block EVM RPC'
ufw deny 8546/tcp comment 'Block EVM WS'
ufw deny 1317/tcp comment 'Block Cosmos REST'
ufw deny 9090/tcp comment 'Block gRPC'

# Enable firewall
ufw --force enable
```

### Step 4: Verify Configuration

```bash
# Check UFW status
ufw status numbered

# Check fail2ban status
fail2ban-client status sshd

# Test from another terminal (don't close current SSH session!)
```

---

## Verification Tests

### From External Machine

```bash
# Test public ports (should be OPEN)
nc -zv 172.239.32.74 80      # Should succeed
nc -zv 172.239.32.74 443     # Should succeed
nc -zv 172.239.32.74 26656   # Should succeed

# Test blocked ports (should TIMEOUT)
nc -zv 172.239.32.74 3000    # Should timeout
nc -zv 172.239.32.74 5432    # Should timeout
nc -zv 172.239.32.74 26657   # Should timeout

# Test HTTPS endpoints (should work)
curl https://c3rt.org
curl https://api.c3rt.org/api/v1/health
curl https://rpc.c3rt.org/status
```

---

## Security Best Practices

### ✅ Implemented
- [x] SSH protected by fail2ban (max 3 attempts, 2-hour ban)
- [x] All public services behind nginx reverse proxy
- [x] Database not accessible from internet
- [x] Internal services blocked from direct access
- [x] P2P port open for validator communication

### 🔄 Recommended Additional Steps
- [ ] Configure SSH key-only authentication (disable password)
- [ ] Set up automated backups
- [ ] Configure log monitoring (Prometheus/Grafana)
- [ ] Set up SSL certificate auto-renewal
- [ ] Configure rate limiting in nginx
- [ ] Set up intrusion detection (OSSEC/Wazuh)

---

## Troubleshooting

### If You Get Locked Out

If UFW blocks your SSH connection:

1. **Access via console** (VPS provider's web console)
2. Disable UFW temporarily:
   ```bash
   ufw disable
   ```
3. Fix the rules and re-enable:
   ```bash
   ufw allow 22/tcp
   ufw enable
   ```

### Check Firewall Logs

```bash
# View UFW logs
tail -f /var/log/ufw.log

# View fail2ban logs
tail -f /var/log/fail2ban.log

# Check blocked IPs
fail2ban-client status sshd
```

### Unblock an IP from fail2ban

```bash
# Unban an IP
fail2ban-client set sshd unbanip <IP_ADDRESS>
```

---

## Quick Reference

| Service | Port | Public Access | Access Method |
|---------|------|---------------|---------------|
| Web | 443 | ✅ Yes | https://c3rt.org |
| API | 443 | ✅ Yes | https://api.c3rt.org |
| RPC | 443 | ✅ Yes | https://rpc.c3rt.org |
| IPFS | 443 | ✅ Yes | https://ipfs.c3rt.org |
| P2P | 26656 | ✅ Yes | Direct TCP |
| SSH | 22 | ⚠️ Admin | Direct TCP |
| PostgreSQL | 5432 | ❌ No | Internal only |
| Direct API | 3000 | ❌ No | Blocked |

---

## Support

For issues or questions:
- Check logs: `docker-compose logs -f`
- View firewall: `ufw status verbose`
- Test connectivity: Run `check-firewall-status.sh`

