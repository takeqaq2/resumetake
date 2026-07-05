#!/bin/bash

echo "========================================="
echo "  ResumeTake AI简历工具 - 部署脚本"
echo "========================================="

# Check if docker is installed
if ! command -v docker &> /dev/null; then
    echo "Docker not found. Installing..."
    curl -fsSL https://get.docker.com | sh
    systemctl start docker
    systemctl enable docker
fi

# Check if docker-compose is installed
if ! command -v docker-compose &> /dev/null; then
    echo "Docker Compose not found. Installing..."
    apt-get update
    apt-get install -y docker-compose
fi

# Create necessary directories
mkdir -p nginx/ssl
mkdir -p data

# Stop existing containers
echo "Stopping existing containers..."
docker-compose down 2>/dev/null || true

# Build and start containers
echo "Building and starting containers..."
docker-compose up -d --build

echo ""
echo "========================================="
echo "  部署完成!"
echo "  访问地址: http://localhost:8088"
echo "========================================="
