#!/bin/bash
# AWS Build Script
# This script builds Docker images for AWS deployment

echo "Building backend..."
cd backend
docker build -t shiftplanner-backend:latest -f Dockerfile .
cd ..

echo "Building frontend..."
cd frontend
docker build -t shiftplanner-frontend:latest -f Dockerfile --build-arg VITE_API_URL= .
cd ..

echo "Build complete!"

