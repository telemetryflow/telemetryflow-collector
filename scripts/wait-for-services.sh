#!/bin/bash
# Wait for services to be ready for E2E tests

set -e

TIMEOUT=60
INTERVAL=2

wait_for_service() {
    local service=$1
    local port=$2
    local timeout=$3
    
    echo "Waiting for $service on port $port..."
    
    for ((i=0; i<timeout; i+=interval)); do
        if nc -z localhost $port 2>/dev/null; then
            echo "$service is ready!"
            return 0
        fi
        sleep $INTERVAL
    done
    
    echo "Timeout waiting for $service"
    return 1
}

# Wait for collector health endpoint
wait_for_service "TelemetryFlow Collector" 13133 $TIMEOUT

echo "All services are ready!"