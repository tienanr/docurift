#!/bin/sh

# Exit on error
set -e

# Set default values for environment variables
BACKEND_URL=${BACKEND_URL:-"http://backend:8080"}
ANALYZER_URL=${ANALYZER_URL:-"http://docurift:9877"}

echo "Waiting for services to be ready..."

# Wait for backend
until curl -s "${BACKEND_URL}/health" > /dev/null; do
    echo "Waiting for backend..."
    sleep 1
done

# Wait for docurift
until curl -s "${ANALYZER_URL}/api/openapi.json" > /dev/null; do
    echo "Waiting for docurift..."
    sleep 1
done

echo "Running tests..."

# Run the test suite
go test -v ./shop

echo "Fetching OpenAPI specification..."

# Fetch OpenAPI spec
curl -s "${ANALYZER_URL}/api/openapi.json" > /tmp/openapi.json

# Basic validation
if ! jq -e '.openapi' /tmp/openapi.json > /dev/null; then
    echo "Error: Invalid OpenAPI spec"
    exit 1
fi

if ! jq -e '.paths' /tmp/openapi.json > /dev/null; then
    echo "Error: No paths found in OpenAPI spec"
    exit 1
fi

echo "Comparing OpenAPI specification with expected output..."

# Replace timestamps with "TIMESTAMP", sort lists, and replace all numbers with 999 in both JSON files
jq 'walk(if type == "string" and test("^\\d{4}-\\d{2}-\\d{2}T\\d{2}:\\d{2}:\\d{2}") then "TIMESTAMP" elif type == "number" then 999 else . end) | (.. | arrays) |= sort' /tmp/openapi.json > /tmp/openapi_processed.json
jq 'walk(if type == "string" and test("^\\d{4}-\\d{2}-\\d{2}T\\d{2}:\\d{2}:\\d{2}") then "TIMESTAMP" elif type == "number" then 999 else . end) | (.. | arrays) |= sort' /app/expected_openapi.json > /tmp/expected_processed.json

# Compare the processed JSON files
if ! diff -u /tmp/expected_processed.json /tmp/openapi_processed.json; then
    echo "Error: OpenAPI specification does not match expected output"
    exit 1
fi

echo "OpenAPI specification matches expected output!"