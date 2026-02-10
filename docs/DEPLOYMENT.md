# Deployment Guide

This guide covers various deployment scenarios for Jabber Bot.

## Table of Contents
- [Prerequisites](#prerequisites)
- [Configuration](#configuration)
- [Docker Deployment](#docker-deployment)
- [Kubernetes Deployment](#kubernetes-deployment)
- [Systemd Service](#systemd-service)
- [Environment Variables](#environment-variables)
- [Monitoring](#monitoring)
- [Security](#security)
- [Troubleshooting](#troubleshooting)

## Prerequisites

### System Requirements
- **CPU**: 1 core minimum, 2 cores recommended
- **Memory**: 256MB minimum, 512MB recommended
- **Storage**: 1GB minimum (for logs and data)
- **Network**: Outbound HTTPS (for webhooks)
- **OS**: Linux, macOS, or Windows

### Software Requirements
- **Go**: 1.21+ (for source deployment)
- **Docker**: 20.10+ (for container deployment)
- **Docker Compose**: 2.0+ (for multi-service deployment)

## Configuration

### Basic Configuration

1. **Copy example configuration:**
   ```bash
   cp .env.example .env
   ```

2. **Edit `.env` file:**
   ```bash
   # Required XMPP settings
   JABBER_BOT_XMPP_JID=your-bot@your-xmpp-server.com
   JABBER_BOT_XMPP_PASSWORD=your-secure-password
   JABBER_BOT_XMPP_SERVER=your-xmpp-server.com:5222
   
   # Required webhook URL
   JABBER_BOT_WEBHOOK_URL=https://your-webhook-endpoint.com/receive
   ```

3. **Optional settings:**
   ```bash
   # Logging level
   JABBER_BOT_LOG_LEVEL=info
   
   # API port
   JABBER_BOT_API_PORT=8080
   
   # Reconnection settings
   JABBER_BOT_RECONNECTION_ENABLED=true
   JABBER_BOT_RECONNECTION_MAX_ATTEMPTS=5
   ```

### Advanced Configuration

For production use, edit `configs/config.yaml`:

```yaml
# XMPP Configuration
xmpp:
  jid: "bot@company.com"
  password: "${JABBER_BOT_PASSWORD}"  # Use env var
  server: "xmpp.company.com:5222"
  resource: "production-bot"

# API Configuration  
api:
  port: 8080
  host: "0.0.0.0"

# Webhook Configuration
webhook:
  url: "https://api.company.com/webhooks/jabber"
  timeout: 30s
  retry_attempts: 5

# Logging Configuration
logging:
  level: "info"
  output: "file"
  file_path: "/var/log/jabber-bot/production.log"

# Reconnection Configuration
reconnection:
  enabled: true
  max_attempts: 10
  backoff: "5s"
```

## Docker Deployment

### Quick Start (Development)

```bash
# Clone the repository
git clone https://github.com/your-org/jabber-bot.git
cd jabber-bot

# Copy and edit configuration
cp .env.example .env
# Edit .env with your settings

# Start development environment
./scripts/deploy.sh dev
```

### Production Deployment

```bash
# Production environment with monitoring
./scripts/deploy.sh prod
```

### Custom Docker Build

```bash
# Build image
docker build -t jabber-bot:custom .

# Run with configuration
docker run -d \
  --name jabber-bot \
  -p 8080:8080 \
  --env-file .env \
  -v $(pwd)/logs:/app/logs \
  jabber-bot:custom
```

### Docker Compose Custom

Create `docker-compose.custom.yml`:

```yaml
version: '3.8'

services:
  jabber-bot:
    image: jabber-bot:latest
    restart: unless-stopped
    ports:
      - "8080:8080"
    environment:
      - JABBER_BOT_XMPP_JID=${JABBER_BOT_XMPP_JID}
      - JABBER_BOT_XMPP_PASSWORD=${JABBER_BOT_XMPP_PASSWORD}
      - JABBER_BOT_XMPP_SERVER=${JABBER_BOT_XMPP_SERVER}
      - JABBER_BOT_WEBHOOK_URL=${JABBER_BOT_WEBHOOK_URL}
      - JABBER_BOT_LOG_LEVEL=info
    volumes:
      - ./logs:/app/logs
      - ./configs/docker-config.yaml:/app/configs/config.yaml:ro
    healthcheck:
      test: ["CMD", "wget", "--spider", "http://localhost:8080/api/v1/health"]
      interval: 30s
      timeout: 10s
      retries: 3
```

Deploy:
```bash
docker-compose -f docker-compose.custom.yml up -d
```

## Kubernetes Deployment

### Namespace

```yaml
# namespace.yaml
apiVersion: v1
kind: Namespace
metadata:
  name: jabber-bot
```

### ConfigMap

```yaml
# configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: jabber-bot-config
  namespace: jabber-bot
data:
  config.yaml: |
    xmpp:
      jid: "bot@company.com"
      server: "xmpp.company.com:5222"
      resource: "k8s-bot"
    api:
      port: 8080
      host: "0.0.0.0"
    webhook:
      url: "https://api.company.com/webhooks/jabber"
      timeout: 30s
      retry_attempts: 5
    logging:
      level: "info"
      output: "stdout"
    reconnection:
      enabled: true
      max_attempts: 5
      backoff: "5s"
```

### Secret

```yaml
# secret.yaml
apiVersion: v1
kind: Secret
metadata:
  name: jabber-bot-secrets
  namespace: jabber-bot
type: Opaque
data:
  xmpp-password: <base64-encoded-password>
```

### Deployment

```yaml
# deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: jabber-bot
  namespace: jabber-bot
  labels:
    app: jabber-bot
spec:
  replicas: 2
  selector:
    matchLabels:
      app: jabber-bot
  template:
    metadata:
      labels:
        app: jabber-bot
    spec:
      containers:
      - name: jabber-bot
        image: jabber-bot:latest
        ports:
        - containerPort: 8080
        env:
        - name: JABBER_BOT_XMPP_PASSWORD
          valueFrom:
            secretKeyRef:
              name: jabber-bot-secrets
              key: xmpp-password
        volumeMounts:
        - name: config
          mountPath: /app/configs
        resources:
          requests:
            memory: "128Mi"
            cpu: "100m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /api/v1/health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /api/v1/health
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
      volumes:
      - name: config
        configMap:
          name: jabber-bot-config
```

### Service

```yaml
# service.yaml
apiVersion: v1
kind: Service
metadata:
  name: jabber-bot-service
  namespace: jabber-bot
spec:
  selector:
    app: jabber-bot
  ports:
  - protocol: TCP
    port: 80
    targetPort: 8080
  type: ClusterIP
```

### Ingress

```yaml
# ingress.yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: jabber-bot-ingress
  namespace: jabber-bot
  annotations:
    nginx.ingress.kubernetes.io/rewrite-target: /
    cert-manager.io/cluster-issuer: "letsencrypt-prod"
spec:
  tls:
  - hosts:
    - jabber-bot.company.com
    secretName: jabber-bot-tls
  rules:
  - host: jabber-bot.company.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: jabber-bot-service
            port:
              number: 80
```

### Deploy to Kubernetes

```bash
# Apply all manifests
kubectl apply -f k8s/

# Check deployment
kubectl get pods -n jabber-bot
kubectl logs -f deployment/jabber-bot -n jabber-bot

# Check service
kubectl get svc -n jabber-bot
```

## Systemd Service

### Service File

Create `/etc/systemd/system/jabber-bot.service`:

```ini
[Unit]
Description=Jabber Bot
After=network.target
Wants=network.target

[Service]
Type=simple
User=jabber-bot
Group=jabber-bot
WorkingDirectory=/opt/jabber-bot
ExecStart=/opt/jabber-bot/bin/jabber-bot -config /opt/jabber-bot/configs/config.yaml
Restart=always
RestartSec=5
StandardOutput=journal
StandardError=journal
SyslogIdentifier=jabber-bot

# Security settings
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ReadWritePaths=/opt/jabber-bot/logs /opt/jabber-bot/data

[Install]
WantedBy=multi-user.target
```

### Setup

```bash
# Create user
sudo useradd --system --home /opt/jabber-bot --shell /bin/false jabber-bot

# Create directories
sudo mkdir -p /opt/jabber-bot/{bin,configs,logs,data}
sudo chown -R jabber-bot:jabber-bot /opt/jabber-bot

# Copy binary and config
sudo cp bin/jabber-bot /opt/jabber-bot/bin/
sudo cp configs/production-config.yaml /opt/jabber-bot/configs/config.yaml
sudo chown -R jabber-bot:jabber-bot /opt/jabber-bot

# Install service file
sudo cp jabber-bot.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable jabber-bot

# Start service
sudo systemctl start jabber-bot
sudo systemctl status jabber-bot
```

## Environment Variables

| Variable | Required | Default | Description |
|-----------|----------|---------|-------------|
| `JABBER_BOT_XMPP_JID` | Yes | - | XMPP JID (e.g., bot@server.com) |
| `JABBER_BOT_XMPP_PASSWORD` | Yes | - | XMPP password |
| `JABBER_BOT_XMPP_SERVER` | Yes | - | XMPP server (e.g., server.com:5222) |
| `JABBER_BOT_XMPP_RESOURCE` | No | bot | XMPP resource |
| `JABBER_BOT_WEBHOOK_URL` | Yes | - | Webhook endpoint URL |
| `JABBER_BOT_WEBHOOK_TIMEOUT` | No | 30s | Webhook timeout |
| `JABBER_BOT_WEBHOOK_RETRY_ATTEMPTS` | No | 3 | Webhook retry attempts |
| `JABBER_BOT_API_HOST` | No | 0.0.0.0 | API bind host |
| `JABBER_BOT_API_PORT` | No | 8080 | API port |
| `JABBER_BOT_LOG_LEVEL` | No | info | Log level (debug, info, warn, error) |
| `JABBER_BOT_LOG_OUTPUT` | No | stdout | Log output (stdout, stderr, file) |
| `JABBER_BOT_LOG_FILE_PATH` | No | - | Log file path (if output=file) |
| `JABBER_BOT_RECONNECTION_ENABLED` | No | true | Enable reconnection |
| `JABBER_BOT_RECONNECTION_MAX_ATTEMPTS` | No | 5 | Max reconnection attempts |
| `JABBER_BOT_RECONNECTION_BACKOFF` | No | 5s | Reconnection backoff |

## Monitoring

### Health Checks

```bash
# Basic health check
curl http://localhost:8080/api/v1/health

# Detailed status
curl http://localhost:8080/api/v1/status

# Webhook status
curl http://localhost:8080/api/v1/webhook/status
```

### Log Monitoring

```bash
# Docker logs
docker logs -f jabber-bot

# Systemd logs
sudo journalctl -u jabber-bot -f

# Kubernetes logs
kubectl logs -f deployment/jabber-bot -n jabber-bot
```

### Prometheus Metrics

If using the Docker Compose monitoring stack:

- **Prometheus**: http://localhost:9090
- **Grafana**: http://localhost:3000 (admin/admin)

### Alerting

Create alert rules for:

- Bot down (health check failures)
- High webhook failure rate
- XMPP connection issues
- High memory usage

## Security

### Network Security
- Use TLS for webhook URLs
- Restrict API access with firewall
- Consider API authentication for production

### XMPP Security
- Use TLS connections (STARTTLS)
- Store password securely (use environment variables)
- Consider OAuth2 or SASL external auth

### Container Security
- Run as non-root user
- Use minimal base image
- Regular security scans
- Resource limits

### Data Security
- Encrypt sensitive configuration
- Secure log storage
- Regular credential rotation

## Troubleshooting

### Common Issues

#### XMPP Connection Failed
```bash
# Check XMPP server connectivity
telnet your-xmpp-server.com 5222

# Verify JID format
# Should be: username@domain.com
# Not: username or domain.com only

# Check logs
docker logs jabber-bot | grep -i xmpp
```

#### Webhook Not Receiving Messages
```bash
# Check webhook URL accessibility
curl -X POST https://your-webhook-url.com/test \
  -H "Content-Type: application/json" \
  -d '{"test": true}'

# Check webhook service status
curl http://localhost:8080/api/v1/webhook/status

# Verify webhook queue
curl http://localhost:8080/api/v1/webhook/status | jq '.queue_length'
```

#### High Memory Usage
```bash
# Check memory usage
docker stats jabber-bot

# Monitor memory over time
while true; do
  echo "$(date): $(docker stats --no-stream --format 'table {{.MemUsage}}' jabber-bot)"
  sleep 60
done
```

#### API Not Responding
```bash
# Check if process is running
ps aux | grep jabber-bot

# Check port binding
netstat -tlnp | grep 8080

# Test API endpoint
curl -v http://localhost:8080/api/v1/health
```

### Debug Mode

Enable debug logging:

```bash
# Environment variable
export JABBER_BOT_LOG_LEVEL=debug

# In configuration file
logging:
  level: "debug"
```

### Performance Tuning

#### XMPP Connection
```yaml
reconnection:
  backoff: "2s"        # Faster reconnect in dev
  max_attempts: 10      # More attempts for stability
```

#### Webhook Performance
```yaml
webhook:
  timeout: "60s"        # Longer timeout for slow endpoints
  retry_attempts: 2      # Fewer retries to reduce load
```

#### API Performance
```yaml
api:
  host: "127.0.0.1"    # Bind locally if behind proxy
  port: 8080
```

### Getting Help

1. **Check logs**: Look for error messages
2. **Verify configuration**: Check JID, server, webhook URL
3. **Test connectivity**: Ensure network access to XMPP server and webhook
4. **Check resources**: Monitor CPU, memory, disk usage
5. **Review documentation**: Check API docs and examples
6. **Open issue**: Report problems with logs and configuration

For production deployments, consider:
- Log aggregation (ELK stack)
- Automated backups
- Disaster recovery plan
- Regular security updates