#!/bin/bash
# CERT Discourse Forum - Automated Deployment Script
# Run as root on production server (172.239.32.74)

set -e  # Exit on error

echo "=========================================="
echo "CERT Discourse Forum Deployment"
echo "=========================================="
echo ""

# Color codes
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check if running as root
if [ "$EUID" -ne 0 ]; then 
    echo -e "${RED}Error: Please run as root (sudo)${NC}"
    exit 1
fi

# Configuration
DISCOURSE_DIR="/var/discourse"
DEPLOY_DIR="/opt/cert-blockchain/deploy-package/services/discourse"
NGINX_SITES="/etc/nginx/sites-available"
NGINX_ENABLED="/etc/nginx/sites-enabled"

echo -e "${YELLOW}Step 1: Checking prerequisites...${NC}"

# Check if Docker is installed
if ! command -v docker &> /dev/null; then
    echo -e "${RED}Error: Docker is not installed${NC}"
    echo "Install Docker first: https://docs.docker.com/engine/install/"
    exit 1
fi

# Check if certbot is installed
if ! command -v certbot &> /dev/null; then
    echo -e "${YELLOW}Warning: certbot not found. Installing...${NC}"
    apt-get update
    apt-get install -y certbot python3-certbot-nginx
fi

echo -e "${GREEN}✓ Prerequisites OK${NC}"
echo ""

echo -e "${YELLOW}Step 2: Generating secrets...${NC}"

# Generate secrets if .env doesn't exist
if [ ! -f "$DEPLOY_DIR/.env" ]; then
    echo "Creating .env file with generated secrets..."
    SSO_SECRET=$(openssl rand -hex 32)
    DB_PASSWORD=$(openssl rand -hex 24)
    SMTP_PASSWORD=""
    
    cat > "$DEPLOY_DIR/.env" <<EOF
# CERT Discourse Environment Configuration
# Generated on $(date)

# SMTP Configuration (REQUIRED - Update this!)
SMTP_ADDRESS=smtp.mailgun.org
SMTP_PORT=587
SMTP_USER=postmaster@mail.c3rt.org
SMTP_PASSWORD=${SMTP_PASSWORD}

# Database Password
DB_PASSWORD=${DB_PASSWORD}

# DiscourseConnect SSO Secret
SSO_SECRET=${SSO_SECRET}

# Domain
DISCOURSE_DOMAIN=forum.c3rt.org
EOF
    
    echo -e "${GREEN}✓ Generated .env file${NC}"
    echo -e "${YELLOW}⚠ IMPORTANT: Edit $DEPLOY_DIR/.env and add your SMTP_PASSWORD${NC}"
    echo ""
    read -p "Press Enter after updating SMTP_PASSWORD in .env file..."
else
    echo -e "${GREEN}✓ .env file already exists${NC}"
    source "$DEPLOY_DIR/.env"
fi

echo ""
echo -e "${YELLOW}Step 3: Setting up Discourse Docker...${NC}"

# Clone discourse_docker if not exists
if [ ! -d "$DISCOURSE_DIR" ]; then
    echo "Cloning discourse_docker repository..."
    git clone https://github.com/discourse/discourse_docker.git "$DISCOURSE_DIR"
    echo -e "${GREEN}✓ Cloned discourse_docker${NC}"
else
    echo -e "${GREEN}✓ discourse_docker already exists${NC}"
fi

# Copy and configure app.yml
echo "Configuring app.yml..."
cp "$DEPLOY_DIR/app.yml" "$DISCOURSE_DIR/containers/app.yml"

# Replace placeholders in app.yml
source "$DEPLOY_DIR/.env"
sed -i "s/REPLACE_WITH_SMTP_PASSWORD/${SMTP_PASSWORD}/g" "$DISCOURSE_DIR/containers/app.yml"
sed -i "s/REPLACE_WITH_SSO_SECRET/${SSO_SECRET}/g" "$DISCOURSE_DIR/containers/app.yml"
sed -i "s/REPLACE_WITH_DB_PASSWORD/${DB_PASSWORD}/g" "$DISCOURSE_DIR/containers/app.yml"

echo -e "${GREEN}✓ Configured app.yml${NC}"
echo ""

echo -e "${YELLOW}Step 4: Setting up SSL certificate...${NC}"

# Check if certificate exists
if [ ! -f "/etc/letsencrypt/live/forum.c3rt.org/fullchain.pem" ]; then
    echo "Obtaining SSL certificate for forum.c3rt.org..."
    certbot certonly --nginx -d forum.c3rt.org --non-interactive --agree-tos --email admin@c3rt.org
    echo -e "${GREEN}✓ SSL certificate obtained${NC}"
else
    echo -e "${GREEN}✓ SSL certificate already exists${NC}"
fi

echo ""
echo -e "${YELLOW}Step 5: Configuring Nginx...${NC}"

# Copy nginx config
cp "$DEPLOY_DIR/nginx-forum.conf" "$NGINX_SITES/forum.c3rt.org"

# Enable site
if [ ! -L "$NGINX_ENABLED/forum.c3rt.org" ]; then
    ln -s "$NGINX_SITES/forum.c3rt.org" "$NGINX_ENABLED/forum.c3rt.org"
fi

# Test nginx config
nginx -t

# Reload nginx
systemctl reload nginx

echo -e "${GREEN}✓ Nginx configured and reloaded${NC}"
echo ""

echo -e "${YELLOW}Step 6: Setting API environment variable...${NC}"

# Add SSO secret to API environment
if ! grep -q "DISCOURSE_SSO_SECRET" /etc/environment; then
    echo "DISCOURSE_SSO_SECRET=\"${SSO_SECRET}\"" >> /etc/environment
    echo -e "${GREEN}✓ Added DISCOURSE_SSO_SECRET to /etc/environment${NC}"
    echo -e "${YELLOW}⚠ Restart CERT API service to apply: systemctl restart cert-api${NC}"
else
    echo -e "${GREEN}✓ DISCOURSE_SSO_SECRET already set${NC}"
fi

echo ""
echo -e "${YELLOW}Step 7: Building Discourse container...${NC}"
echo "This will take 5-15 minutes..."
echo ""

cd "$DISCOURSE_DIR"
./launcher rebuild app

echo ""
echo -e "${GREEN}=========================================="
echo "✓ Discourse deployment complete!"
echo "==========================================${NC}"
echo ""
echo "Next steps:"
echo "1. Visit https://forum.c3rt.org"
echo "2. Create admin account with email: admin@c3rt.org"
echo "3. Install custom theme:"
echo "   - Go to Admin → Customize → Themes"
echo "   - Upload theme from: $DEPLOY_DIR/theme/"
echo "4. Restart CERT API: systemctl restart cert-api"
echo ""
echo "SSO Secret: ${SSO_SECRET}"
echo "(Save this - it's needed for API configuration)"
echo ""

