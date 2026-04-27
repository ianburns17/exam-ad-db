#!/bin/bash
# Additional API handler tests
# Usage: bash test_extra_api.sh

API_URL="http://localhost:4000"

# 1. Healthcheck

echo "Testing healthcheck endpoint:"
curl -i "$API_URL/v1/healthcheck"
echo

# 2. List all customers

echo "Testing list customers endpoint:"
curl -i "$API_URL/v1/customers"
echo

# 3. List all locations

echo "Testing list locations endpoint:"
curl -i "$API_URL/v1/locations"
echo

# 4. List all rentals

echo "Testing list rentals endpoint:"
curl -i "$API_URL/v1/rentals"
echo

echo "Done."
