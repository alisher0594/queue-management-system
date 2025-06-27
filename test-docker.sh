#!/bin/bash

echo "Testing Docker build..."

# Build the Docker image
docker build -t queue-management-system . || {
    echo "Docker build failed!"
    exit 1
}

echo "Docker build successful!"

# Test the container
echo "Testing container startup..."
docker run -d -p 8080:8080 -e PORT=8080 --name queue-test queue-management-system

# Wait a moment for startup
sleep 5

# Check if container is running
if docker ps | grep -q queue-test; then
    echo "Container is running successfully!"
    
    # Test the health endpoint
    if curl -f http://localhost:8080/ > /dev/null 2>&1; then
        echo "Application is responding to HTTP requests!"
    else
        echo "Application might be starting up..."
    fi
else
    echo "Container failed to start"
    docker logs queue-test
fi

# Cleanup
docker stop queue-test
docker rm queue-test

echo "Test completed!"
