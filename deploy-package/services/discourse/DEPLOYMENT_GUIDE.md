# CERT Discourse Community Forum - Complete Deployment Guide

## 📋 Overview

This guide covers the complete deployment of the CERT Discourse community forum, including:
- Docker container setup
- CertID SSO integration
- Custom theme installation
- Website integration
- Production deployment

**Domain:** `forum.c3rt.org`  
**SSO Provider:** `api.c3rt.org/api/v1/discourse/sso`  
**Main Site:** `c3rt.org`

---

## 🚀 Quick Start (Production)

### Prerequisites

- Ubuntu 22.04 LTS or higher
- Docker installed
- Root/sudo access
- Domain `forum.c3rt.org` pointing to server
- SMTP credentials (Mailgun, SendGrid, etc.)

### One-Command Deployment

```bash
cd /opt/cert-blockchain/deploy-package/services/discourse
sudo ./deploy.sh
```

This automated script will:
1. ✅ Check prerequisites
2. ✅ Generate secure secrets
3. ✅ Clone Discourse Docker
4. ✅ Configure app.yml
5. ✅ Obtain SSL certificate
6. ✅ Configure Nginx
7. ✅ Build Discourse container (5-15 mins)

---

## 📝 Manual Deployment Steps

### Step 1: Generate Secrets

```bash
# Generate SSO secret (32 bytes)
openssl rand -hex 32

# Generate database password (24 bytes)
openssl rand -hex 24
```

Save these securely - you'll need them for configuration.

### Step 2: Configure Environment

Create `.env` file:

```bash
cd /opt/cert-blockchain/deploy-package/services/discourse
cp .env.example .env
nano .env
```

Fill in:
```env
SMTP_PASSWORD=your_mailgun_password
DB_PASSWORD=<generated_password>
SSO_SECRET=<generated_secret>
```

### Step 3: Clone Discourse Docker

```bash
sudo git clone https://github.com/discourse/discourse_docker.git /var/discourse
cd /var/discourse
```

### Step 4: Configure app.yml

```bash
sudo cp /opt/cert-blockchain/deploy-package/services/discourse/app.yml containers/app.yml
sudo nano containers/app.yml
```

Replace placeholders:
- `REPLACE_WITH_SMTP_PASSWORD` → Your SMTP password
- `REPLACE_WITH_SSO_SECRET` → Generated SSO secret
- `REPLACE_WITH_DB_PASSWORD` → Generated DB password

### Step 5: Configure Nginx

```bash
# Copy nginx config
sudo cp /opt/cert-blockchain/deploy-package/services/discourse/nginx-forum.conf \
  /etc/nginx/sites-available/forum.c3rt.org

# Enable site
sudo ln -s /etc/nginx/sites-available/forum.c3rt.org \
  /etc/nginx/sites-enabled/forum.c3rt.org

# Test configuration
sudo nginx -t

# Reload nginx
sudo systemctl reload nginx
```

### Step 6: Obtain SSL Certificate

```bash
sudo certbot certonly --nginx -d forum.c3rt.org \
  --non-interactive --agree-tos --email admin@c3rt.org
```

### Step 7: Configure API Server

Add SSO secret to CERT API environment:

```bash
# Add to /etc/environment
echo 'DISCOURSE_SSO_SECRET="<your_sso_secret>"' | sudo tee -a /etc/environment

# Restart API service
sudo systemctl restart cert-api
```

### Step 8: Build Discourse

```bash
cd /var/discourse
sudo ./launcher rebuild app
```

This takes 5-15 minutes. Monitor progress:
```bash
sudo ./launcher logs app
```

---

## 🎨 Theme Installation

### Automated Theme Packaging

```bash
cd /opt/cert-blockchain/deploy-package/services/discourse
./install-theme.sh
```

### Manual Theme Installation

1. **Login as Admin**
   - Visit `https://forum.c3rt.org`
   - Create account with email: `admin@c3rt.org`
   - First user becomes admin

2. **Navigate to Themes**
   - Go to Admin → Customize → Themes
   - URL: `https://forum.c3rt.org/admin/customize/themes`

3. **Install Theme**
   - Click "Install" → "From a Git Repository"
   - OR manually upload files:
     - `theme/about.json`
     - `theme/common/common.scss`
     - `theme/common/header.html`

4. **Set as Default**
   - Click on "CERT Dark Theme"
   - Click "Set as default"

5. **Verify Colors**
   - Background: `#050508` (ink)
   - Mint: `#00FFA3`
   - Electric: `#4D9FFF`
   - Cyber: `#9D00FF`

---

## 🔗 SSO Integration

### How It Works

```
User clicks "Login" on forum
  ↓
Discourse redirects to: api.c3rt.org/api/v1/discourse/sso?sso=...&sig=...
  ↓
API verifies HMAC signature
  ↓
API checks if user is authenticated (JWT cookie)
  ↓
If NOT authenticated → Redirect to c3rt.org/login
  ↓
User logs in with wallet (MetaMask/Keplr)
  ↓
API fetches CertID profile from database
  ↓
API builds Discourse user payload:
  - external_id: wallet address
  - email: <address>@wallet.c3rt.org
  - username: from CertID or truncated address
  - name: from CertID profile
  - avatar_url: from CertID profile
  ↓
API signs payload with SSO secret
  ↓
Redirect back to forum with signed payload
  ↓
Discourse creates/updates user account
  ↓
User is logged in!
```

### Testing SSO

1. **Clear browser cookies**
2. **Visit** `https://forum.c3rt.org`
3. **Click "Login"**
4. **Should redirect** to `c3rt.org/login`
5. **Connect wallet** (MetaMask/Keplr)
6. **Should redirect back** to forum, logged in
7. **Verify** username and avatar from CertID

---

## 🌐 Website Integration

The website has been updated to integrate the forum:

### Community Page (`/community`)

- ✅ New "Join the Forum" card (first position)
- ✅ Prominent hero section with forum CTA
- ✅ External link with icon
- ✅ Forum categories preview

### Navigation

- ✅ Footer: Community section includes "Forum" link
- ✅ External link handling in footer component

### Changes Made

**Files Modified:**
- `cert-web/src/pages/Community.jsx` - Added forum card and hero
- `cert-web/src/config/nav.js` - Added forum to Community section
- `cert-web/src/components/SiteFooter.jsx` - External link support

---

## 🔧 Configuration Reference

### Discourse Settings (app.yml)

```yaml
DISCOURSE_HOSTNAME: 'forum.c3rt.org'
DISCOURSE_DEVELOPER_EMAILS: 'admin@c3rt.org'

# SMTP (Mailgun)
DISCOURSE_SMTP_ADDRESS: smtp.mailgun.org
DISCOURSE_SMTP_PORT: 587
DISCOURSE_SMTP_USER_NAME: postmaster@mail.c3rt.org

# SSO
DISCOURSE_ENABLE_DISCOURSE_CONNECT: true
DISCOURSE_DISCOURSE_CONNECT_URL: 'https://api.c3rt.org/api/v1/discourse/sso'
DISCOURSE_ENABLE_LOCAL_LOGINS: false  # SSO only
```

### Nginx Configuration

```nginx
server {
    listen 443 ssl http2;
    server_name forum.c3rt.org;
    
    location / {
        proxy_pass http://localhost:8080;
        # WebSocket support for live updates
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
    }
}
```

### Plugins Included

- ✅ **docker_manager** - Update management
- ✅ **discourse-oauth2-basic** - OAuth integration
- ✅ **discourse-solved** - Q&A mark-as-solved
- ✅ **discourse-voting** - Feature voting
- ✅ **discourse-assign** - Topic assignment

---

## 📊 Post-Deployment Checklist

### Initial Setup

- [ ] Forum accessible at `https://forum.c3rt.org`
- [ ] SSL certificate valid
- [ ] Admin account created
- [ ] Custom theme installed and set as default
- [ ] SSO login working
- [ ] Email notifications working

### Configuration

- [ ] Site title: "CERT Community"
- [ ] Logo uploaded
- [ ] Favicon set
- [ ] Contact email configured
- [ ] Categories created:
  - [ ] General Discussion
  - [ ] Development
  - [ ] Governance
  - [ ] Support
  - [ ] Announcements

### Testing

- [ ] Create test topic
- [ ] Reply to topic
- [ ] Test voting plugin
- [ ] Test solved plugin
- [ ] Test email notifications
- [ ] Test SSO logout
- [ ] Test mobile responsiveness

---

## 🐛 Troubleshooting

### Forum Not Accessible

```bash
# Check if container is running
docker ps | grep discourse

# View logs
cd /var/discourse
sudo ./launcher logs app

# Restart container
sudo ./launcher restart app
```

### SSO Not Working

```bash
# Check API logs
sudo journalctl -u cert-api -f

# Verify SSO secret matches
# In Discourse: containers/app.yml
# In API: /etc/environment or docker-compose.yml

# Test SSO endpoint
curl https://api.c3rt.org/api/v1/discourse/sso
```

### Theme Not Applied

1. Go to Admin → Customize → Themes
2. Click on "CERT Dark Theme"
3. Click "Set as default"
4. Clear browser cache
5. Hard refresh (Ctrl+Shift+R)

### Email Not Sending

```bash
# Enter container
cd /var/discourse
sudo ./launcher enter app

# Test email
rails c
> SiteSetting.notification_email
> Email::Sender.new(message, :test).send

# Check SMTP settings in app.yml
```

---

## 📈 Monitoring

### Health Check

```bash
# Container status
docker ps | grep discourse

# Nginx logs
tail -f /var/log/nginx/forum.c3rt.org.access.log
tail -f /var/log/nginx/forum.c3rt.org.error.log

# Discourse logs
cd /var/discourse
sudo ./launcher logs app
```

### Performance

- Monitor CPU/RAM usage
- Check database size
- Review slow queries
- Monitor SMTP delivery rate

---

## 🔄 Updates

### Update Discourse

```bash
cd /var/discourse
sudo ./launcher rebuild app
```

### Update Theme

1. Update theme files in `theme/` directory
2. Re-upload to Discourse admin panel
3. OR use Git repository method for auto-updates

---

## 📞 Support

- **Documentation**: `/opt/cert-blockchain/deploy-package/services/discourse/README.md`
- **Discourse Docs**: https://docs.discourse.org
- **CERT Docs**: https://c3rt.org/docs

---

**Deployment Date**: _To be filled after deployment_  
**Deployed By**: _To be filled_  
**Version**: Discourse 3.x with CERT customizations

