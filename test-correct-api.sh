#!/bin/bash

# Test Correct Zoho WorkDrive API Endpoints
ACCESS_TOKEN=$(grep -o '"access_token":"[^"]*"' zoho_tokens.json | cut -d'"' -f4)

echo "🔗 Testing Correct Zoho WorkDrive API Endpoints"
echo "=============================================="
echo ""
echo "🔑 Access Token: ${ACCESS_TOKEN:0:20}..."
echo ""

# Test 1: Get account information (correct endpoint)
echo "📋 Test 1: Get account information..."
curl -X GET "https://www.zohoapis.com/workdrive/api/v1/account" \
  -H "Authorization: Zoho-oauthtoken $ACCESS_TOKEN" \
  -H "Accept: application/json" \
  -s | python3 -m json.tool 2>/dev/null || echo "Raw response received"

echo ""
echo "📋 Test 2: List files in root..."
curl -X GET "https://www.zohoapis.com/workdrive/api/v1/files" \
  -H "Authorization: Zoho-oauthtoken $ACCESS_TOKEN" \
  -H "Accept: application/json" \
  -s | python3 -m json.tool 2>/dev/null || echo "Raw response received"

echo ""
echo "📋 Test 3: Get user details..."
curl -X GET "https://www.zohoapis.com/workdrive/api/v1/users" \
  -H "Authorization: Zoho-oauthtoken $ACCESS_TOKEN" \
  -H "Accept: application/json" \
  -s | python3 -m json.tool 2>/dev/null || echo "Raw response received"

echo ""
echo "📋 Test 4: Alternative authorization header format..."
curl -X GET "https://www.zohoapis.com/workdrive/api/v1/account" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Accept: application/json" \
  -s | python3 -m json.tool 2>/dev/null || echo "Raw response received"

echo ""
echo "🎯 Testing completed - check responses above for success!"