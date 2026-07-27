#!/bin/bash

# Docker Hub deployment script with environment selection
# Customize the variables below before running

set -e  # Exit on any error

# Configuration - UPDATE THESE
DOCKER_HUB_USERNAME="treycodes"
IMAGE_NAME="gift-grab"
IMAGE_TAG="latest"
DOCKERFILE_PATH=".."
LOCAL_YML_PATH="$DOCKERFILE_PATH/local.yml"

# Full image name with Docker Hub repo
FULL_IMAGE_NAME="$DOCKER_HUB_USERNAME/$IMAGE_NAME:$IMAGE_TAG"

# Prompt for environment
echo "🌍 Select environment:"
echo "1) production"
echo "2) staging"
echo "3) development"
echo ""
read -p "Enter choice (1-3): " env_choice

case $env_choice in
    1)
        ENVIRONMENT="production"
        ;;
    2)
        ENVIRONMENT="staging"
        ;;
    3)
        ENVIRONMENT="development"
        ;;
    *)
        echo "❌ Invalid choice"
        exit 1
        ;;
esac

echo "✅ Using environment: $ENVIRONMENT"

# Update local.yml with selected environment
echo "🔧 Updating local.yml..."
sed -i.bak "s/ENVIRONMENT=[^\"]*\"/ENVIRONMENT=$ENVIRONMENT\"/" "$LOCAL_YML_PATH"

if [ $? -ne 0 ]; then
    echo "❌ Failed to update local.yml"
    exit 1
fi

echo "✅ Updated local.yml"

# Build Docker image
echo "🐳 Building Docker image..."
docker build --platform linux/arm64 -t "$FULL_IMAGE_NAME" -f "$DOCKERFILE_PATH/Dockerfile" "$DOCKERFILE_PATH"

if [ $? -ne 0 ]; then
    echo "❌ Build failed"
    exit 1
fi

echo "✅ Build successful"

# Confirm before pushing
read -p "Deploy $ENVIRONMENT build to Docker Hub? (y/n) " confirm
if [[ ! $confirm =~ ^[Yy]$ ]]; then
    echo "Cancelled"
    exit 0
fi

# Push to Docker Hub
echo "📤 Pushing to Docker Hub..."
docker push "$FULL_IMAGE_NAME"

if [ $? -ne 0 ]; then
    echo "❌ Push failed"
    exit 1
fi

echo "✅ Successfully deployed to Docker Hub!"
echo "📍 Image available at: $FULL_IMAGE_NAME"
echo "🌍 Environment: $ENVIRONMENT"