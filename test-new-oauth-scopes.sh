#!/bin/bash

# Test with Extended OAuth Scopes for WorkDrive
CLIENT_ID="1000.Z520MJ3HS00YJEKRHRX0U9KGZTATPX"
CLIENT_SECRET="731702ae155269b29c1997664def3553764face6f8"
REDIRECT_URI="http://localhost:8080/callback"

echo "🔧 Testing with Extended WorkDrive OAuth Scopes"
echo "==============================================="
echo ""

# Extended scopes that might be needed
EXTENDED_SCOPES="WorkDrive.files.ALL,WorkDrive.workspace.READ,WorkDrive.organization.READ,WorkDrive.teamfolder.READ,WorkDrive.privatespace.READ,ZohoFiles.files.ALL"

echo "🚀 STEP 1: New Authorization URL with Extended Scopes"
echo "====================================================="
echo ""
echo "Current scopes might be insufficient. Try this URL with extended permissions:"
echo ""

AUTH_URL="https://accounts.zoho.com/oauth/v2/auth"
EXTENDED_AUTH_URL="${AUTH_URL}?response_type=code&client_id=${CLIENT_ID}&scope=${EXTENDED_SCOPES}&redirect_uri=${REDIRECT_URI}&access_type=offline&prompt=consent"

echo "${EXTENDED_AUTH_URL}"
echo ""
echo "📋 Extended scopes include:"
echo "- WorkDrive.files.ALL"
echo "- WorkDrive.workspace.READ" 
echo "- WorkDrive.organization.READ"
echo "- WorkDrive.teamfolder.READ"
echo "- WorkDrive.privatespace.READ"
echo "- ZohoFiles.files.ALL (legacy compatibility)"
echo ""

echo "🔍 STEP 2: Check Current Token Scopes"
echo "===================================="
echo "Current token scopes:"
grep -o '"scope":"[^"]*"' token_response.json | cut -d'"' -f4
echo ""

echo "🛠️  STEP 3: Alternative - Test ZohoFiles API"
echo "============================================"
echo "WorkDrive might use the older ZohoFiles API endpoints..."

ACCESS_TOKEN=$(grep -o '"access_token":"[^"]*"' zoho_tokens.json | cut -d'"' -f4)

# Test ZohoFiles endpoints (legacy)
echo "Testing ZohoFiles API endpoints:"
files_endpoints=(
    "https://www.zohoapis.com/files/v1/files"
    "https://files.zoho.com/api/v1/files"
    "https://www.zohoapis.com/files/api/v1/files"
)

for endpoint in "${files_endpoints[@]}"; do
    echo "Testing: $endpoint"
    response=$(curl -s "$endpoint" \
        -H "Authorization: Zoho-oauthtoken $ACCESS_TOKEN" \
        -H "Accept: application/json")
    
    if echo "$response" | grep -q '"data"'; then
        echo "  ✅ SUCCESS! ZohoFiles API working"
        echo "$response" | python3 -m json.tool | head -10
        break
    elif echo "$response" | grep -q '"errors"'; then
        error=$(echo "$response" | grep -o '"title":"[^"]*"' | cut -d'"' -f4)
        echo "  ❌ Error: $error"
    else
        echo "  ❓ Unknown response: ${response:0:50}..."
    fi
done

echo ""
echo "🎯 Next Actions:"
echo "==============="
echo "1. If current API still fails, visit the extended OAuth URL above"
echo "2. Re-authorize with additional scopes"
echo "3. Get new tokens with extended permissions"
echo "4. Try ZohoFiles API as alternative"
echo ""
echo "💡 The issue might be scope limitations rather than endpoint problems."