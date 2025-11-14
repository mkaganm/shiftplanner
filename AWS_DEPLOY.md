# AWS EC2 Deployment Guide

## Problem
AWS EC2'de Docker build yaparken `stat /app/cmd/server: directory not found` hatası alınıyor.

## Solution
AWS'de build context root'tan başlatılmalı. İki seçenek var:

### Seçenek 1: docker-compose.aws.yml kullan (ÖNERİLEN)

```bash
# AWS EC2'de
cd ~/shiftplanner
docker compose -f docker-compose.aws.yml up -d --build
```

### Seçenek 2: Manuel build

```bash
# Backend build (root context)
cd ~/shiftplanner
docker build -t shiftplanner-backend:latest -f Dockerfile.backend .

# Frontend build (root context)
docker build -t shiftplanner-frontend:latest -f Dockerfile.frontend --build-arg VITE_API_URL= .

# Run containers
docker run -d --name shiftplanner-backend -p 8080:8080 \
  -e PORT=8080 -e DATA_DIR=/data -e ALLOWED_ORIGINS=http://localhost:3000 \
  -v backend-data:/data shiftplanner-backend:latest

docker run -d --name shiftplanner-frontend -p 3000:3000 \
  --link shiftplanner-backend:backend shiftplanner-frontend:latest
```

### Seçenek 3: docker-compose.yml'i düzelt

Eğer normal `docker-compose.yml` kullanmak istiyorsanız, AWS'de şu şekilde kullanın:

```bash
# Build context'i root olarak ayarla
cd ~/shiftplanner
docker compose build --build-arg BUILD_CONTEXT=.
docker compose up -d
```

## Debug

Eğer hala sorun yaşıyorsanız, debug için:

```bash
# Backend Dockerfile'da debug satırları çalışacak
docker build -t shiftplanner-backend:debug -f backend/Dockerfile ./backend
```

Bu komut hangi dosyaların kopyalandığını gösterecek.

