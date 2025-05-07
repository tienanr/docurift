#!/bin/sh

# Exit on error
set -e

echo "Waiting for services to be ready..."

# Wait for backend
until curl -s "${BACKEND_URL}/health" > /dev/null; do
    echo "Waiting for backend..."
    sleep 1
done

# Wait for docurift
until curl -s "${ANALYZER_URL}/openapi.json" > /dev/null; do
    echo "Waiting for docurift..."
    sleep 1
done

echo "Running tests..."

# Run the test suite
go test -v ./test/...

# Fetch and validate OpenAPI spec
curl -s "${ANALYZER_URL}/openapi.json" > /tmp/openapi.json

# Basic validation
if ! jq -e '.openapi' /tmp/openapi.json > /dev/null; then
    echo "Error: Invalid OpenAPI spec"
    exit 1
fi

if ! jq -e '.paths' /tmp/openapi.json > /dev/null; then
    echo "Error: No paths found in OpenAPI spec"
    exit 1
fi

echo "Tests completed successfully!" 