#!/bin/bash

# Test Zoho WorkDrive API with Access Token
source zoho_tokens.json 2>/dev/null || {
    ACCESS_TOKEN=$(grep -o '"access_token":"[^"]*"' zoho_tokens.json | cut -d'"' -f4)
}

echo "🔗 Testing Zoho WorkDrive API Connection"
echo "======================================"
echo ""
echo "🔑 Using Access Token: ${ACCESS_TOKEN:0:20}..."
echo ""

# Test 1: Get user information
echo "📋 Test 1: Getting user information..."
curl -X GET "https://www.zohoapis.com/workdrive/api/v1/users/me" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Accept: application/json" \
  -s | python3 -m json.tool 2>/dev/null || echo "Response received (not JSON formatted)"

echo ""
echo "📋 Test 2: List workspaces..."
curl -X GET "https://www.zohoapis.com/workdrive/api/v1/privatespace" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Accept: application/json" \
  -s | python3 -m json.tool 2>/dev/null || echo "Response received (not JSON formatted)"

echo ""
echo "📋 Test 3: List team folders..."
curl -X GET "https://www.zohoapis.com/workdrive/api/v1/teamfolders" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Accept: application/json" \
  -s | python3 -m json.tool 2>/dev/null || echo "Response received (not JSON formatted)"

echo ""
echo "🎉 API tests completed!"
echo "✅ If you see JSON responses above, the API connection is working!"
echo ""
echo "🚀 ZohoSync is now ready for full implementation!"