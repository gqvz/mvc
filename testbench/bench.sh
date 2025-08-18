#!/bin/bash

source ./testbench/setup.sh

echo "Starting API benchmarking process..."

source testbench/credentials.env

BASE_URL="http://localhost:3000/api"

echo "Testing API connectivity..."
if ! curl -s "$BASE_URL/items" > /dev/null 2>&1; then
    echo "Error: API is not accessible at $BASE_URL. Please make sure the server is running."
    exit 1
fi

echo "Getting JWT token for user: $USERNAME"
TOKEN_RESPONSE=$(curl -s -X POST \
    -H "Content-Type: application/json" \
    -d "{
        \"username\": \"$USERNAME\",
        \"password\": \"$PASSWORD\"
    }" \
    "$BASE_URL/token")

JWT_TOKEN=$(echo "$TOKEN_RESPONSE" | jq -r '.token')

if [ -z "$JWT_TOKEN" ] || [ "$JWT_TOKEN" = "null" ]; then
    echo "Error: Failed to retrieve JWT token. Please check your user credentials and the API endpoint."
    echo "Response: $TOKEN_RESPONSE"
    exit 1
fi

echo "Successfully retrieved JWT token."

JSON_FILE=$(mktemp)
echo '{
  "custom_instructions": "Extra spicy please",
  "item_id": 1,
  "quantity": 2
}' > "$JSON_FILE"

DURATION=10
CONCURRENCY=1000

echo "Starting benchmarks..."

echo "Benchmarking GET /api/items"
echo "Duration: ${DURATION}s, Concurrency: $CONCURRENCY"
ab -t $DURATION -c $CONCURRENCY \
    -H "Authorization: Bearer $JWT_TOKEN" \
    -H "Content-Type: application/json" \
    "$BASE_URL/items" 2>/dev/null

echo ""
echo "Results for GET /api/items completed"
echo ""

echo "Benchmarking POST /api/orders/$ORDER_ID/items"
echo "Duration: ${DURATION}s, Concurrency: $CONCURRENCY"
echo "Using order ID: $ORDER_ID"

ab -l -t $DURATION -c $CONCURRENCY \
    -H "Authorization: Bearer $JWT_TOKEN" \
    -H "Content-Type: application/json" \
    -T "application/json" \
    -p "$JSON_FILE" \
    "$BASE_URL/orders/$ORDER_ID/items" 2>/dev/null

echo ""
echo "Results for POST /api/orders/$ORDER_ID/items completed"
echo ""

rm -f "$JSON_FILE"

echo "Closing order $ORDER_ID..."
CLOSE_RESPONSE=$(curl -s -X POST \
    -H "Authorization: Bearer $JWT_TOKEN" \
    -H "Content-Type: application/json" \
    "$BASE_URL/orders/$ORDER_ID/close")

if echo "$CLOSE_RESPONSE" | jq -e '.message' > /dev/null 2>&1; then
    echo "Order closed successfully"
else
    echo "Order close response: $CLOSE_RESPONSE"
fi

echo "Benchmarking completed!"
echo "All tests finished. Check the results above for performance metrics."
