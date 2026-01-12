# 🔥 FIREWALL CONFIGURATION - ACTION REQUIRED

**Server:** 172.239.32.74 (c3rt.org)  
**Status:** ⚠️ **SECURITY ISSUES DETECTED**  
**Priority:** 🔴 **HIGH**

---

## 🚨 Critical Issues

### 1. Port 3000 (API) is Publicly Accessible ❌
**Risk:** Direct API access bypasses nginx security, rate limiting, and SSL  
**Impact:** HIGH - Security vulnerability  
**Action:** Block port 3000

### 2. Port 26656 (P2P) is Blocked ❌
**Risk:** Validators cannot connect, network cannot sync  
**Impact:** HIGH - Blockchain functionality broken  
**Action:** Open port 26656

### 3. fail2ban Not Installed ❌
**Risk:** SSH brute force attacks  
**Impact:** MEDIUM - Server security  
**Action:** Install and configure fail2ban

---

## ✅ Quick Fix (5 Minutes)

### Option A: Automated (Recommended)

```bash
# From your local machine at /opt
cd cert-blockchain/scripts
./apply-firewall-remote.sh
```

This script will:
1. ✅ Install fail2ban
2. ✅ Open port 26656 (P2P)
3. ✅ Block port 3000 (direct API)
4. ✅ Block all other internal ports
5. ✅ Verify configuration

---

### Option B: Manual (If SSH Available)

```bash
# SSH to server
ssh root@172.239.32.74

# Install fail2ban
apt update && apt install -y fail2ban

# Configure firewall
ufw allow 22/tcp
ufw allow 80/tcp
ufw allow 443/tcp
ufw allow 26656/tcp
ufw deny 3000/tcp
ufw deny 5432/tcp
ufw deny 26657/tcp
ufw --force enable

# Verify
ufw status numbered
```

---

## 📊 Current vs Required State

| Port | Service | Current | Required | Action |
|------|---------|---------|----------|--------|
| 22 | SSH | ✅ Open | ✅ Open | None |
| 80 | HTTP | ✅ Open | ✅ Open | None |
| 443 | HTTPS | ✅ Open | ✅ Open | None |
| **3000** | **API** | ⚠️ **Open** | ❌ **Blocked** | **CLOSE** |
| **26656** | **P2P** | ❌ **Blocked** | ✅ **Open** | **OPEN** |
| 5432 | PostgreSQL | ✅ Blocked | ✅ Blocked | None |
| 26657 | RPC | ✅ Blocked | ✅ Blocked | None |

---

## 🔍 Verification

After applying the fix, verify:

```bash
# Test from external machine
nc -zv 172.239.32.74 26656   # Should succeed (P2P open)
nc -zv 172.239.32.74 3000    # Should timeout (API blocked)

# Test HTTPS endpoints (should work)
curl https://api.c3rt.org/api/v1/health   # Should return {"status":"healthy"}
curl https://c3rt.org                      # Should return HTML
```

---

## 📖 Documentation

- **Complete Guide:** [FIREWALL-SETUP.md](FIREWALL-SETUP.md)
- **Deployment Guide:** [DEPLOYMENT.md](DEPLOYMENT.md)
- **Scripts:**
  - `scripts/configure-firewall.sh` - Firewall configuration
  - `scripts/apply-firewall-remote.sh` - Remote deployment
  - `scripts/check-firewall-status.sh` - Status verification

---

## ⏱️ Timeline

**Estimated Time:** 5-10 minutes  
**Downtime:** None (services continue running)  
**Risk:** Low (can revert via console if needed)

---

## 🆘 Emergency Rollback

If something goes wrong:

```bash
# Via console access (VPS provider)
ufw disable

# Or reset to open state
ufw --force reset
ufw allow 22/tcp
ufw allow 80/tcp
ufw allow 443/tcp
ufw enable
```

---

## ✅ Post-Configuration Checklist

After applying firewall rules:

- [ ] Port 26656 is accessible (test with `nc -zv 172.239.32.74 26656`)
- [ ] Port 3000 is blocked (test with `nc -zv 172.239.32.74 3000`)
- [ ] HTTPS endpoints work (test `https://api.c3rt.org/api/v1/health`)
- [ ] SSH still accessible (test from another terminal)
- [ ] fail2ban is running (`systemctl status fail2ban`)
- [ ] UFW is enabled (`ufw status`)

---

## 📞 Next Steps

1. **Apply firewall configuration** (choose Option A or B above)
2. **Verify configuration** (run verification tests)
3. **Monitor logs** for 24 hours
4. **Update documentation** if any issues found

---

## 🔐 Security Impact

**Before:**
- ⚠️ Direct API access possible (bypasses nginx)
- ❌ P2P network cannot sync
- ⚠️ No SSH brute force protection

**After:**
- ✅ All services behind nginx reverse proxy
- ✅ P2P network fully functional
- ✅ SSH protected by fail2ban
- ✅ Rate limiting enforced
- ✅ SSL/TLS for all public endpoints

