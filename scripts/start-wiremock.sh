#!/bin/bash

# Stop any existing container
docker stop wiremock || true
docker rm wiremock || true

# Get the directory where the script is located
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Change to the project root directory
cd "$PROJECT_ROOT"

# Build the image with our mappings
docker build -t eigenlayer-wiremock -f wiremock/Dockerfile wiremock/

# Run the container
docker run -d --name wiremock -p 8081:8080 eigenlayer-wiremock

# Wait for WireMock to start
sleep 2

echo "WireMock is running in container on http://localhost:8081"
echo "Mappings loaded:"
echo "  - GET /list-operator-releases"
echo "  - GET /list-avs-release-keys" 