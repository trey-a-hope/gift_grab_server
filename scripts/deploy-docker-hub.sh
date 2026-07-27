#!/bin/bash

# Docker Hub deployment script
# Customize the variables below before running

set -e  # Exit on any error

# Configuration - UPDATE THESE
DOCKER_HUB_USERNAME="treycodes"
IMAGE_NAME="gift-grab"
IMAGE_TAG="latest"
DOCKERFILE_PATH=".."

# Full image name with Docker Hub repo
FULL_IMAGE_NAME="$DOCKER_HUB_USERNAME/$IMAGE_NAME:$IMAGE_TAG"

echo "🐳 Building Docker image..."

docker build --platform linux/arm64 -t "$FULL_IMAGE_NAME" -f "$DOCKERFILE_PATH/Dockerfile" "$DOCKERFILE_PATH"

if [ $? -ne 0 ]; then
    echo "❌ Build failed"
    exit 1
fi

echo "✅ Build successful"
echo "📤 Pushing to Docker Hub..."

docker push "$FULL_IMAGE_NAME"

if [ $? -ne 0 ]; then
    echo "❌ Push failed"
    
    exit 1
fi

echo "✅ Successfully deployed to Docker Hub!"
echo "📍 Image available at: $FULL_IMAGE_NAME"