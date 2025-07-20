#!/bin/bash

# Test Different Zoho WorkDrive API Endpoints
ACCESS_TOKEN=$(grep -o '"access_token":"[^"]*"' zoho_tokens.json | cut -d'"' -f4)

echo "🔗 Testing Various Zoho WorkDrive API Endpoints"
echo "=============================================="
echo ""
echo "🔑 Access Token: ${ACCESS_TOKEN:0:20}..."
echo ""

# Test different authorization header formats
AUTH_HEADERS=(
    "Authorization: Zoho-oauthtoken $ACCESS_TOKEN"
    "Authorization: Bearer $ACCESS_TOKEN"
    "Authorization: OAuth $ACCESS_TOKEN"
)

# Test different API endpoints
ENDPOINTS=(
    "https://www.zohoapis.com/workdrive/api/v1/files"
    "https://www.zohoapis.com/workdrive/api/v1/workspaces"
    "https://www.zohoapis.com/workdrive/api/v1/teamfolders"
    "https://www.zohoapis.com/workdrive/api/v1/privatespace"
    "https://www.zohoapis.com/workdrive/api/v1/account"
    "https://www.zohoapis.com/workdrive/api/v1/users"
    "https://workdrive.zoho.com/api/v1/files"
    "https://www.zohoapis.com/workdrive/v1/files"
)

# Function to test endpoint
test_endpoint() {
    local endpoint="$1"
    local auth_header="$2"
    
    echo "📋 Testing: $endpoint"
    echo "   Auth: ${auth_header:0:50}..."
    
    response=$(curl -s -X GET "$endpoint" -H "$auth_header" -H "Accept: application/json")
    
    if echo "$response" | grep -q '"errors"'; then
        error_id=$(echo "$response" | grep -o '"id":"[^"]*"' | cut -d'"' -f4)
        error_title=$(echo "$response" | grep -o '"title":"[^"]*"' | cut -d'"' -f4)
        echo "   ❌ Error: $error_id - $error_title"
    elif echo "$response" | grep -q '"data"'; then
        echo "   ✅ SUCCESS: Got data response"
        echo "$response" | python3 -m json.tool | head -10
    elif echo "$response" | grep -q '"status"'; then
        echo "   ✅ SUCCESS: Got status response"  
        echo "$response" | python3 -m json.tool | head -10
    else
        echo "   ❓ Unknown response:"
        echo "   ${response:0:100}..."
    fi
    echo ""
}

# Test each combination
for auth_header in "${AUTH_HEADERS[@]}"; do
    echo "🔐 Testing with auth header: ${auth_header:0:30}..."
    echo "=================================================="
    
    for endpoint in "${ENDPOINTS[@]}"; do
        test_endpoint "$endpoint" "$auth_header"
    done
    
    echo ""
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo ""
done

echo "🎯 Testing completed - look for ✅ SUCCESS responses above!"