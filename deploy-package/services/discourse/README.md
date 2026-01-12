# CERT Community Forum (Discourse)

Self-hosted Discourse forum integrated with CERT Blockchain via DiscourseConnect SSO.

**🌐 Live:** https://forum.c3rt.org
**📚 Full Guide:** [DEPLOYMENT_GUIDE.md](./DEPLOYMENT_GUIDE.md)

## 🚀 Quick Start

### Automated Deployment (Recommended)

```bash
cd /opt/cert-blockchain/deploy-package/services/discourse
sudo ./deploy.sh
```

This script handles everything automatically. See [DEPLOYMENT_GUIDE.md](./DEPLOYMENT_GUIDE.md) for details.

### Prerequisites
- Ubuntu 22.04 LTS (or higher)
- 2GB RAM minimum (4GB recommended)
- Docker installed
- Domain pointing to your server (e.g., `forum.c3rt.org`)
- SMTP credentials (Mailgun, SendGrid, etc.)

### 1. Clone Discourse Docker

```bash
sudo -s
git clone https://github.com/discourse/discourse_docker.git /var/discourse
cd /var/discourse
```

### 2. Copy Configuration

```bash
cp /path/to/deploy-package/services/discourse/app.yml containers/app.yml
```

### 3. Configure Secrets

Edit `containers/app.yml` and replace:
- `REPLACE_WITH_SMTP_PASSWORD` - Your SMTP password
- `REPLACE_WITH_SSO_SECRET` - Generate with `openssl rand -hex 32`
- `REPLACE_WITH_DB_PASSWORD` - Generate with `openssl rand -hex 24`

### 4. Set API Environment Variable

On your CERT API server, set the same SSO secret:
```bash
export DISCOURSE_SSO_SECRET="your_sso_secret_here"
```

### 5. Bootstrap & Launch

```bash
./launcher rebuild app
```

This takes 5-15 minutes. When complete, Discourse is running on port 8080.

### 6. Install Custom Theme

1. Go to `forum.c3rt.org/admin/customize/themes`
2. Click "Install" → "From Git Repository"
3. Upload the theme files from `theme/` directory
4. Set as default theme

## Architecture

```
┌─────────────────┐     ┌──────────────────┐     ┌─────────────────┐
│   cert-web      │────▶│   CERT API       │────▶│   Discourse     │
│   (React)       │     │   /discourse/sso │     │   (forum.c3rt)  │
└─────────────────┘     └──────────────────┘     └─────────────────┘
        │                        │
        │                        ▼
        │               ┌──────────────────┐
        └──────────────▶│   CertID DB      │
                        │   (profiles)     │
                        └──────────────────┘
```

## SSO Flow

1. User clicks "Login" on Discourse
2. Discourse redirects to `api.c3rt.org/api/v1/discourse/sso`
3. API checks for authenticated wallet session
4. If not logged in → redirect to CERT login page
5. If logged in → fetch CertID profile, build SSO payload
6. Redirect back to Discourse with signed payload
7. Discourse creates/updates user account

## Theme Colors

The custom theme matches the cert-web design system:
- **Background**: `#050508` (ink)
- **Surface**: `#0A0A0F`
- **Mint accent**: `#00FFA3`
- **Electric**: `#4D9FFF`
- **Cyber**: `#9D00FF`

## Nginx Reverse Proxy

If running behind nginx with the main site:

```nginx
server {
    listen 443 ssl http2;
    server_name forum.c3rt.org;

    ssl_certificate /etc/letsencrypt/live/forum.c3rt.org/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/forum.c3rt.org/privkey.pem;

    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

## Troubleshooting

### Logs
```bash
cd /var/discourse
./launcher logs app
```

### Enter Container
```bash
./launcher enter app
```

### Rebuild After Config Changes
```bash
./launcher rebuild app
```

## Plugins Included

- **docker_manager** - Update management
- **discourse-voting** - Feature voting
- **discourse-solved** - Q&A mark-as-solved (bundled with Discourse)
- **discourse-assign** - Topic assignment (bundled with Discourse)
- **discourse-oauth2-basic** - OAuth integration (bundled with Discourse)

## Website Integration

The CERT website has been updated to integrate the forum:

### Community Page Updates
- ✅ New "Join the Forum" card (first position)
- ✅ Prominent hero section with forum CTA
- ✅ External link with icon
- ✅ Forum categories preview

### Navigation Updates
- ✅ Footer: Community section includes "Forum" link
- ✅ External link handling in footer component

### Files Modified
- `cert-web/src/pages/Community.jsx` - Added forum card and hero
- `cert-web/src/config/nav.js` - Added forum to Community section
- `cert-web/src/components/SiteFooter.jsx` - External link support

Visit https://c3rt.org/community to see the integration.

## Scripts

- **deploy.sh** - Automated production deployment
- **quick-start.sh** - Local development with docker-compose
- **install-theme.sh** - Theme packaging helper

## Resources

- [Full Deployment Guide](./DEPLOYMENT_GUIDE.md)
- [Discourse Docker](https://github.com/discourse/discourse_docker)
- [DiscourseConnect SSO](https://meta.discourse.org/t/discourseconnect-official-single-sign-on-for-discourse-sso/13045)
- [CERT Blockchain](https://c3rt.org)
