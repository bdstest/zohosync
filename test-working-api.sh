#!/bin/bash

# Test WorkDrive API with Confirmed Access
ACCESS_TOKEN=$(grep -o '"access_token":"[^"]*"' zoho_tokens.json | cut -d'"' -f4)
YOUR_FILE_ID="veysx16db130021d84de08b78167afc76c011"

echo "🎯 Testing WorkDrive API with Confirmed US Data Center Access"
echo "============================================================"
echo ""
echo "✅ Your WorkDrive: https://workdrive.zoho.com"
echo "✅ Your File ID: $YOUR_FILE_ID"
echo "🔑 Access Token: ${ACCESS_TOKEN:0:20}..."
echo ""

# Test 1: Get specific file info (we know this file exists)
echo "📋 Test 1: Get your file information"
echo "-----------------------------------"
curl -s "https://www.zohoapis.com/workdrive/api/v1/files/$YOUR_FILE_ID" \
    -H "Authorization: Zoho-oauthtoken $ACCESS_TOKEN" \
    -H "Accept: application/json" | python3 -m json.tool 2>/dev/null || echo "Response received (checking format)"

echo ""
echo "📋 Test 2: List files in workspace root"
echo "--------------------------------------"
# Try different workspace endpoints
endpoints=(
    "https://www.zohoapis.com/workdrive/api/v1/files"
    "https://www.zohoapis.com/workdrive/api/v1/files?parent_id=root"
    "https://www.zohoapis.com/workdrive/api/v1/workspaces"
    "https://www.zohoapis.com/workdrive/api/v1/privatespace/files"
    "https://www.zohoapis.com/workdrive/api/v1/home"
)

for endpoint in "${endpoints[@]}"; do
    echo "Testing: $endpoint"
    response=$(curl -s "$endpoint" \
        -H "Authorization: Zoho-oauthtoken $ACCESS_TOKEN" \
        -H "Accept: application/json")
    
    if echo "$response" | grep -q '"data"'; then
        echo "  ✅ SUCCESS! Got data response"
        echo "$response" | python3 -m json.tool | head -10
        break
    elif echo "$response" | grep -q '"errors"'; then
        error_id=$(echo "$response" | grep -o '"id":"[^"]*"' | cut -d'"' -f4)
        error_title=$(echo "$response" | grep -o '"title":"[^"]*"' | cut -d'"' -f4)
        echo "  ❌ Error: $error_id - $error_title"
    else
        echo "  ❓ Unexpected response: ${response:0:50}..."
    fi
    echo ""
done

echo "📋 Test 3: Alternative authentication headers"
echo "--------------------------------------------"
auth_formats=(
    "Authorization: Zoho-oauthtoken $ACCESS_TOKEN"
    "Authorization: Bearer $ACCESS_TOKEN"
    "X-ZAPI-AUTH-TOKEN: $ACCESS_TOKEN"
)

for auth in "${auth_formats[@]}"; do
    echo "Testing auth format: ${auth:0:30}..."
    response=$(curl -s "https://www.zohoapis.com/workdrive/api/v1/files/$YOUR_FILE_ID" \
        -H "$auth" \
        -H "Accept: application/json")
    
    if echo "$response" | grep -q '"data"'; then
        echo "  ✅ SUCCESS with this auth format!"
        echo "$response" | python3 -m json.tool | head -5
        break
    else
        error=$(echo "$response" | grep -o '"title":"[^"]*"' | cut -d'"' -f4)
        echo "  ❌ Failed: $error"
    fi
done

echo ""
echo "🎯 Summary: Testing with your confirmed WorkDrive file access..."