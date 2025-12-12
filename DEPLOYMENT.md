# CERT Blockchain Production Deployment Guide

**Server:** 172.239.32.74  
**Domain:** C3rt.org  
**Chain ID:** 951753

## Prerequisites

- Ubuntu 22.04 LTS (or similar)
- Docker & Docker Compose
- 4+ CPU cores, 8GB+ RAM, 100GB+ SSD
- Open ports: 22, 80, 443, 26656, 26657

## Quick Deployment

### 1. Transfer Files to Server

```bash
# From local machine
scp -r cert-blockchain root@172.239.32.74:/opt/
```

### 2. Run Deployment Script

```bash
ssh root@172.239.32.74
cd /opt/cert-blockchain
chmod +x scripts/deploy-production.sh
./scripts/deploy-production.sh
```

### 3. Configure Nginx & SSL

```bash
# Install nginx
apt update && apt install -y nginx certbot python3-certbot-nginx

# Copy nginx config
cp scripts/nginx-production.conf /etc/nginx/sites-available/cert-blockchain
ln -s /etc/nginx/sites-available/cert-blockchain /etc/nginx/sites-enabled/
rm /etc/nginx/sites-enabled/default

# Get SSL certificates
certbot --nginx -d c3rt.org -d www.c3rt.org -d api.c3rt.org -d rpc.c3rt.org -d ipfs.c3rt.org

# Restart nginx
systemctl restart nginx
```

### 4. Configure Firewall

```bash
ufw allow 22/tcp    # SSH
ufw allow 80/tcp    # HTTP (redirect)
ufw allow 443/tcp   # HTTPS
ufw allow 26656/tcp # P2P
ufw enable
```

## Services

| Service | Port | Endpoint |
|---------|------|----------|
| REST API | 3000 | https://api.c3rt.org |
| RPC | 26657 | https://rpc.c3rt.org |
| P2P | 26656 | 172.239.32.74:26656 |
| IPFS Gateway | 8080 | https://ipfs.c3rt.org |
| PostgreSQL | 5432 | Internal only |

## Verify Deployment

```bash
# Check blockchain status
curl https://rpc.c3rt.org/status

# Check API health
curl https://api.c3rt.org/api/v1/health

# Check all containers
docker-compose ps
```

## Security Checklist

- [ ] Change POSTGRES_PASSWORD in .env
- [ ] Change JWT_SECRET in .env  
- [ ] Enable UFW firewall
- [ ] Configure fail2ban for SSH
- [ ] Set up automated backups
- [ ] Configure log rotation
- [ ] Set up monitoring (Prometheus/Grafana)

## Maintenance

```bash
# View logs
docker-compose logs -f certd
docker-compose logs -f api

# Restart services
docker-compose restart

# Update deployment
git pull
docker-compose build --no-cache
docker-compose up -d

# Backup blockchain data
docker-compose exec certd tar -czf /tmp/backup.tar.gz /root/.certd
docker cp certd:/tmp/backup.tar.gz ./backups/
```

## Troubleshooting

### Container won't start
```bash
docker-compose logs certd
docker-compose down && docker-compose up -d
```

### Database connection issues
```bash
docker-compose exec postgres psql -U cert -d certid -c "SELECT 1"
```

### Reset blockchain (CAUTION: deletes all data)
```bash
docker-compose down -v
docker-compose up -d
```

