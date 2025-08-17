#!/bin/bash

echo "Setting up testbench environment..."

HOST="localhost:3000"
BASE_URL="http://$HOST/api"

echo "Testing API connectivity..."
if ! curl -s "$BASE_URL/items" > /dev/null 2>&1; then
    echo "Error: API is not accessible at $BASE_URL. Please make sure the server is running."
    exit 1
fi

echo "Getting admin token..."
ADMIN_TOKEN_RESPONSE=$(curl -s -X POST \
    -H "Content-Type: application/json" \
    -d '{
        "username": "admin",
        "password": "admin"
    }' \
    "$BASE_URL/token")

ADMIN_TOKEN=$(echo "$ADMIN_TOKEN_RESPONSE" | jq -r '.token')

if [ "$ADMIN_TOKEN" = "null" ] || [ -z "$ADMIN_TOKEN" ]; then
    echo "Error: Failed to get admin token. Response: $ADMIN_TOKEN_RESPONSE"
    exit 1
fi

echo "Admin token obtained successfully"

echo "Creating tags..."
TAGS=("appetizer" "main" "dessert" "beverage" "vegetarian" "spicy" "popular" "chef-special")

for tag in "${TAGS[@]}"; do
    TAG_RESPONSE=$(curl -s -X POST \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $ADMIN_TOKEN" \
        -d "{\"name\": \"$tag\"}" \
        "$BASE_URL/tags")
    
    TAG_ID=$(echo "$TAG_RESPONSE" | jq -r '.id')
    if [ "$TAG_ID" != "null" ] && [ -n "$TAG_ID" ]; then
        echo "Tag '$tag' created with ID: $TAG_ID"
    else
        echo "Tag '$tag' creation failed or already exists"
    fi
done

echo "Creating 20 items..."
ITEM_NAMES=(
    "Margherita Pizza" "Caesar Salad" "Chicken Wings" "Beef Burger" "Fish Tacos"
    "Pasta Carbonara" "Greek Salad" "BBQ Ribs" "Sushi Roll" "Chicken Curry"
    "Steak Frites" "Vegetable Soup" "Shrimp Scampi" "Mushroom Risotto" "Lamb Chops"
    "Caprese Salad" "Duck Confit" "Seafood Paella" "Beef Stir Fry" "Vegetable Lasagna"
)

ITEM_DESCRIPTIONS=(
    "Classic tomato and mozzarella pizza" "Fresh romaine lettuce with Caesar dressing" "Crispy chicken wings with hot sauce" "Juicy beef burger with cheese" "Fresh fish tacos with salsa"
    "Creamy pasta with bacon and eggs" "Mediterranean salad with feta" "Tender BBQ ribs with sauce" "Fresh salmon sushi roll" "Spicy chicken curry with rice"
    "Grilled steak with French fries" "Hearty vegetable soup" "Garlic shrimp with pasta" "Creamy mushroom risotto" "Herb-crusted lamb chops"
    "Fresh mozzarella and tomato salad" "Slow-cooked duck with potatoes" "Spanish seafood rice dish" "Wok-fried beef with vegetables" "Layered vegetable pasta"
)

PRICES=(12.99 8.99 10.99 13.99 11.99 14.99 9.99 16.99 15.99 12.99 22.99 7.99 18.99 16.99 24.99 8.99 26.99 19.99 13.99 11.99)

TAG_COMBINATIONS=(
    "main,popular" "appetizer,vegetarian" "main,spicy" "main,popular" "main"
    "main,popular" "appetizer,vegetarian" "main,spicy" "main" "main,spicy"
    "main,chef-special" "appetizer,vegetarian" "main" "main,vegetarian" "main,chef-special"
    "appetizer,vegetarian" "main,chef-special" "main" "main,spicy" "main,vegetarian"
)

for i in {0..19}; do
    ITEM_NAME="${ITEM_NAMES[$i]}"
    ITEM_DESC="${ITEM_DESCRIPTIONS[$i]}"
    ITEM_PRICE="${PRICES[$i]}"
    ITEM_TAGS="${TAG_COMBINATIONS[$i]}"
    
    TAGS_JSON=$(echo "$ITEM_TAGS" | tr ',' '\n' | jq -R . | jq -s .)
    
    ITEM_RESPONSE=$(curl -s -X POST \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $ADMIN_TOKEN" \
        -d "{
            \"name\": \"$ITEM_NAME\",
            \"description\": \"$ITEM_DESC\",
            \"price\": $ITEM_PRICE,
            \"image_url\": \"https://http.cat/200\",
            \"tags\": $TAGS_JSON,
            \"available\": true
        }" \
        "$BASE_URL/items")
    
    ITEM_ID=$(echo "$ITEM_RESPONSE" | jq -r '.id')
    if [ "$ITEM_ID" != "null" ] && [ -n "$ITEM_ID" ]; then
        echo "Item '$ITEM_NAME' created with ID: $ITEM_ID"
    else
        echo "Item '$ITEM_NAME' creation failed or already exists"
    fi
done

echo "Getting user token..."
TOKEN_RESPONSE=$(curl -s -X POST \
    -H "Content-Type: application/json" \
    -d '{
        "username": "admin",
        "password": "admin"
    }' \
    "$BASE_URL/token")

TOKEN=$(echo "$TOKEN_RESPONSE" | jq -r '.token')

if [ "$TOKEN" = "null" ] || [ -z "$TOKEN" ]; then
    echo "Error: Failed to get user token. Response: $TOKEN_RESPONSE"
    exit 1
fi

echo "User token obtained successfully"

echo "Creating order..."
ORDER_RESPONSE=$(curl -s -X POST \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $TOKEN" \
    -d '{
        "table_number": 999
    }' \
    "$BASE_URL/orders")

ORDER_ID=$(echo "$ORDER_RESPONSE" | jq -r '.order_id')

if [ "$ORDER_ID" = "null" ] || [ -z "$ORDER_ID" ]; then
    echo "Error: Failed to create order. Response: $ORDER_RESPONSE"
    exit 1
fi

echo "Order created with ID: $ORDER_ID"

echo "HOST=$HOST" > testbench/credentials.env
echo "BASE_URL=$BASE_URL" >> testbench/credentials.env
echo "USERNAME=admin" >> testbench/credentials.env
echo "PASSWORD=admin" >> testbench/credentials.env
echo "ORDER_ID=$ORDER_ID" >> testbench/credentials.env

echo "Setup completed successfully!"
echo "Credentials saved to testbench/credentials.env"
echo "Order ID: $ORDER_ID"
echo "You can now run ./testbench/bench.sh to start benchmarking"
