#!/bin/bash
set -e

# Start time measurement
start_time=$(date +%s)

# ==============================
# Usage:
#   ./build-and-push.sh <version> <docker_username> <docker_password>
# Example:
#   ./build-and-push.sh v0.0.1 zwj666 my-docker-token
# ==============================

if [ $# -ne 3 ]; then
  echo "❌ Usage: $0 <version> <docker_username> <docker_password_or_token>"
  exit 1
fi

# Parameters
TAG="$1"
DOCKER_USER="$2"
DOCKER_PASS="$3"
GIT_SHA=$(git rev-parse --short HEAD || echo "nogit")

# Proxy
HTTP_PROXY=http://172.23.0.1:10808
HTTPS_PROXY=http://172.23.0.1:10808

# Image names
# SERVICE_IMAGE=whalefell/ucdc-service
# WEB_IMAGE=whalefell/ucdc-web
IMAGE_NAME=whalefell/map-server

# ========================
# Login if not already
# ========================
if docker info 2>/dev/null | grep -q "Username: $DOCKER_USER"; then
  echo "🔐 Already logged in as $DOCKER_USER"
else
  echo "🔐 Logging in to Docker Hub as $DOCKER_USER..."
  echo "$DOCKER_PASS" | docker login --username "$DOCKER_USER" --password-stdin
fi

# ========================
# Build Docker Images
# ========================
echo "🔨 Building Docker images with tag: $TAG"
build_start_time=$(date +%s)


docker build \
  --build-arg HTTP_PROXY=${HTTP_PROXY} \
  --build-arg HTTPS_PROXY=${HTTPS_PROXY} \
  -f ./Dockerfile \
  -t ${IMAGE_NAME}:${TAG} \
  -t ${IMAGE_NAME}:latest \
  . # context

build_end_time=$(date +%s)
build_duration=$((build_end_time - build_start_time))
echo "⏱️ Build completed in ${build_duration} seconds"

# ========================
# Push Docker Images
# ========================
echo "📤 Pushing Docker images..."
push_start_time=$(date +%s)

docker push ${IMAGE_NAME}:${TAG}
docker push ${IMAGE_NAME}:latest

push_end_time=$(date +%s)
push_duration=$((push_end_time - push_start_time))
echo "⏱️ Push completed in ${push_duration} seconds"

# Calculate total elapsed time
end_time=$(date +%s)
total_duration=$((end_time - start_time))
# Format duration in minutes and seconds
minutes=$((total_duration / 60))
seconds=$((total_duration % 60))

echo "✅ Build and push completed! Tag: $TAG (Git: $GIT_SHA)"
echo "⏱️ Total elapsed time: ${minutes}m ${seconds}s (${total_duration} seconds)"
