# MiroChain Deployment Guide

## Overview

Это руководство описывает процесс развертывания MiroChain блокчейн сети в различных средах.

## System Requirements

### Minimum Requirements
- **CPU**: 2 cores
- **RAM**: 4 GB
- **Storage**: 10 GB SSD
- **Network**: 100 Mbps
- **OS**: Linux (Ubuntu 20.04+), macOS (10.15+), Windows 10+

### Recommended Requirements
- **CPU**: 4+ cores
- **RAM**: 8+ GB
- **Storage**: 50+ GB SSD
- **Network**: 1 Gbps
- **OS**: Linux (Ubuntu 22.04 LTS)

## Prerequisites

### Required Software
- Go 1.25+
- Git
- Make (optional)

### Installation

#### Ubuntu/Debian
```bash
# Update system
sudo apt update && sudo apt upgrade -y

# Install Go
wget https://go.dev/dl/go1.25.1.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.25.1.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc

# Install Git
sudo apt install git -y

# Install Make
sudo apt install build-essential -y
```

#### macOS
```bash
# Install Homebrew (if not installed)
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"

# Install Go
brew install go

# Install Git
brew install git

# Install Make
brew install make
```

#### Windows
1. Download and install Go from https://golang.org/dl/
2. Download and install Git from https://git-scm.com/download/win
3. Download and install Make from https://www.gnu.org/software/make/

## Quick Start

### 1. Clone Repository
```bash
git clone https://github.com/your-org/mirochain.git
cd mirochain
```

### 2. Install Dependencies
```bash
go mod tidy
```

### 3. Build
```bash
# Build all binaries
make build

# Or build individually
go build -o bin/node cmd/node/main.go
go build -o bin/wallet cmd/wallet/main.go
```

### 4. Run
```bash
# Start node
./bin/node

# Or with custom port
./bin/node -port=8080

# Create wallet
./bin/wallet -create
```

## Configuration

### Environment Variables
```bash
# Node configuration
export MIROCHAIN_PORT=8080
export MIROCHAIN_DATA_DIR=./data
export MIROCHAIN_LOG_LEVEL=info
export MIROCHAIN_MINING_ENABLED=true
export MIROCHAIN_DIFFICULTY=4

# Network configuration
export MIROCHAIN_PEERS=127.0.0.1:8081,127.0.0.1:8082
export MIROCHAIN_NODE_ID=node_001
```

### Configuration File
Create `config.yaml`:
```yaml
node:
  port: 8080
  data_dir: "./data"
  log_level: "info"
  mining_enabled: true
  difficulty: 4

network:
  peers:
    - "127.0.0.1:8081"
    - "127.0.0.1:8082"
  node_id: "node_001"
  max_peers: 50

mining:
  enabled: true
  difficulty: 4
  reward: 50
  mempool_size: 1000

api:
  enabled: true
  port: 8080
  cors_enabled: true
```

## Deployment Scenarios

### Single Node Deployment

#### Docker
```dockerfile
FROM golang:1.25-alpine AS builder

WORKDIR /app
COPY . .
RUN go mod tidy
RUN go build -o mirochain cmd/node/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/mirochain .
COPY --from=builder /app/configs/config.yaml .

EXPOSE 8080
CMD ["./mirochain"]
```

```bash
# Build and run
docker build -t mirochain .
docker run -p 8080:8080 mirochain
```

#### Systemd Service
Create `/etc/systemd/system/mirochain.service`:
```ini
[Unit]
Description=MiroChain Node
After=network.target

[Service]
Type=simple
User=mirochain
WorkingDirectory=/opt/mirochain
ExecStart=/opt/mirochain/bin/node
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

```bash
# Create user
sudo useradd -r -s /bin/false mirochain

# Create directory
sudo mkdir -p /opt/mirochain
sudo cp -r . /opt/mirochain/
sudo chown -R mirochain:mirochain /opt/mirochain

# Enable and start service
sudo systemctl enable mirochain
sudo systemctl start mirochain
sudo systemctl status mirochain
```

### Multi-Node Deployment

#### Docker Compose
Create `docker-compose.yml`:
```yaml
version: '3.8'

services:
  node1:
    build: .
    ports:
      - "8080:8080"
    environment:
      - MIROCHAIN_PORT=8080
      - MIROCHAIN_NODE_ID=node_001
      - MIROCHAIN_PEERS=node2:8080,node3:8080
    volumes:
      - ./data/node1:/data

  node2:
    build: .
    ports:
      - "8081:8080"
    environment:
      - MIROCHAIN_PORT=8080
      - MIROCHAIN_NODE_ID=node_002
      - MIROCHAIN_PEERS=node1:8080,node3:8080
    volumes:
      - ./data/node2:/data

  node3:
    build: .
    ports:
      - "8082:8080"
    environment:
      - MIROCHAIN_PORT=8080
      - MIROCHAIN_NODE_ID=node_003
      - MIROCHAIN_PEERS=node1:8080,node2:8080
    volumes:
      - ./data/node3:/data
```

```bash
# Start cluster
docker-compose up -d

# Check status
docker-compose ps

# View logs
docker-compose logs -f node1
```

#### Kubernetes
Create `k8s-deployment.yaml`:
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: mirochain-node
spec:
  replicas: 3
  selector:
    matchLabels:
      app: mirochain
  template:
    metadata:
      labels:
        app: mirochain
    spec:
      containers:
      - name: mirochain
        image: mirochain:latest
        ports:
        - containerPort: 8080
        env:
        - name: MIROCHAIN_PORT
          value: "8080"
        - name: MIROCHAIN_NODE_ID
          valueFrom:
            fieldRef:
              fieldPath: metadata.name
        volumeMounts:
        - name: data
          mountPath: /data
      volumes:
      - name: data
        persistentVolumeClaim:
          claimName: mirochain-data
---
apiVersion: v1
kind: Service
metadata:
  name: mirochain-service
spec:
  selector:
    app: mirochain
  ports:
  - port: 8080
    targetPort: 8080
  type: LoadBalancer
---
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: mirochain-data
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 10Gi
```

```bash
# Deploy to Kubernetes
kubectl apply -f k8s-deployment.yaml

# Check status
kubectl get pods
kubectl get services
```

### Cloud Deployment

#### AWS EC2
```bash
# Launch EC2 instance
aws ec2 run-instances \
  --image-id ami-0c02fb55956c7d316 \
  --instance-type t3.medium \
  --key-name your-key \
  --security-groups mirochain-sg

# Connect to instance
ssh -i your-key.pem ec2-user@your-instance-ip

# Install and run MiroChain
# ... (follow single node deployment steps)
```

#### Google Cloud Platform
```bash
# Create instance
gcloud compute instances create mirochain-node \
  --image-family ubuntu-2004-lts \
  --image-project ubuntu-os-cloud \
  --machine-type e2-medium \
  --zone us-central1-a

# Connect to instance
gcloud compute ssh mirochain-node --zone us-central1-a

# Install and run MiroChain
# ... (follow single node deployment steps)
```

#### Azure
```bash
# Create resource group
az group create --name mirochain-rg --location eastus

# Create VM
az vm create \
  --resource-group mirochain-rg \
  --name mirochain-node \
  --image UbuntuLTS \
  --size Standard_B2s \
  --admin-username azureuser \
  --generate-ssh-keys

# Connect to VM
ssh azureuser@your-vm-ip

# Install and run MiroChain
# ... (follow single node deployment steps)
```

## Monitoring and Logging

### Logging Configuration
```yaml
# config.yaml
logging:
  level: info
  format: json
  output: file
  file_path: /var/log/mirochain/node.log
  max_size: 100MB
  max_backups: 5
  max_age: 30
```

### Monitoring with Prometheus
Create `prometheus.yml`:
```yaml
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: 'mirochain'
    static_configs:
      - targets: ['localhost:8080']
```

### Health Checks
```bash
# Check node health
curl http://localhost:8080/api/v1/blockchain/status

# Check if node is mining
curl http://localhost:8080/api/v1/mining/status

# Check network connectivity
curl http://localhost:8080/api/v1/network/peers
```

## Security Considerations

### Firewall Configuration
```bash
# Allow only necessary ports
sudo ufw allow 8080/tcp
sudo ufw allow 22/tcp
sudo ufw enable
```

### SSL/TLS Configuration
```yaml
# config.yaml
api:
  tls:
    enabled: true
    cert_file: /path/to/cert.pem
    key_file: /path/to/key.pem
```

### Access Control
```yaml
# config.yaml
api:
  auth:
    enabled: true
    jwt_secret: "your-secret-key"
    token_expiry: "24h"
```

## Backup and Recovery

### Backup Script
```bash
#!/bin/bash
# backup.sh

BACKUP_DIR="/backup/mirochain"
DATA_DIR="/opt/mirochain/data"
DATE=$(date +%Y%m%d_%H%M%S)

# Create backup directory
mkdir -p $BACKUP_DIR

# Backup data
tar -czf $BACKUP_DIR/mirochain_$DATE.tar.gz $DATA_DIR

# Keep only last 7 days of backups
find $BACKUP_DIR -name "mirochain_*.tar.gz" -mtime +7 -delete

echo "Backup completed: mirochain_$DATE.tar.gz"
```

### Recovery
```bash
# Stop node
sudo systemctl stop mirochain

# Restore data
tar -xzf /backup/mirochain/mirochain_20240101_120000.tar.gz -C /

# Start node
sudo systemctl start mirochain
```

## Troubleshooting

### Common Issues

#### Node won't start
```bash
# Check logs
journalctl -u mirochain -f

# Check configuration
./bin/node -config-check

# Check port availability
netstat -tlnp | grep 8080
```

#### Mining not working
```bash
# Check mining status
curl http://localhost:8080/api/v1/mining/status

# Check mempool
curl http://localhost:8080/api/v1/blockchain/status

# Restart mining
curl -X POST http://localhost:8080/api/v1/mining/stop
curl -X POST http://localhost:8080/api/v1/mining/start
```

#### Network connectivity issues
```bash
# Check peer connections
curl http://localhost:8080/api/v1/network/peers

# Check firewall
sudo ufw status

# Test connectivity
telnet peer-ip 8080
```

### Performance Tuning

#### System Limits
```bash
# Increase file descriptor limit
echo "* soft nofile 65536" >> /etc/security/limits.conf
echo "* hard nofile 65536" >> /etc/security/limits.conf

# Increase network buffer sizes
echo 'net.core.rmem_max = 16777216' >> /etc/sysctl.conf
echo 'net.core.wmem_max = 16777216' >> /etc/sysctl.conf
sysctl -p
```

#### Go Runtime Tuning
```bash
# Set Go environment variables
export GOGC=100
export GOMAXPROCS=4
export GOMEMLIMIT=4GiB
```

## Maintenance

### Regular Tasks
- Monitor disk space usage
- Check log file sizes
- Verify network connectivity
- Update software dependencies
- Backup data regularly

### Updates
```bash
# Pull latest changes
git pull origin main

# Update dependencies
go mod tidy

# Rebuild and restart
make build
sudo systemctl restart mirochain
```

## Support

### Getting Help
- GitHub Issues: https://github.com/your-org/mirochain/issues
- Documentation: https://docs.mirochain.com
- Community Forum: https://forum.mirochain.com

### Reporting Issues
When reporting issues, please include:
- MiroChain version
- Operating system
- Configuration files
- Log files
- Steps to reproduce
