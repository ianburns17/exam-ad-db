#!/bin/bash
# Test script for API rate limiting, CORS, compression, and /metrics endpoint
# Usage: bash test_api.sh

API_URL="http://localhost:4000"  # Change this if your API runs elsewhere


echo
echo "Testing CORS preflight (OPTIONS):"
curl -i -X OPTIONS "$API_URL/v1/vehicles" \
  -H "Origin: http://example.com" \
  -H "Access-Control-Request-Method: GET"
echo

echo "Testing compressed response:"
curl -i -H "Accept-Encoding: gzip" "$API_URL/v1/vehicles"
echo

echo "Testing uncompressed response:"
curl -i -H "Accept-Encoding: identity" "$API_URL/v1/vehicles"
echo

echo "Testing /metrics endpoint:"
curl -i "$API_URL/metrics"
echo

echo "Testing rate limit (expecting 429):"
for i in {1..10}
do
  http_code=$(curl -s -o /dev/null -w "%{http_code}" "$API_URL/v1/vehicles")
  echo "Request $i: HTTP $http_code"
  if [ "$http_code" = "429" ]; then
    echo "Rate limit triggered on request $i"
    break
  fi
done

echo "Done."
