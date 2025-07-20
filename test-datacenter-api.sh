#!/bin/bash

# Test Different Zoho Data Centers and API Configurations
ACCESS_TOKEN=$(grep -o '"access_token":"[^"]*"' zoho_tokens.json | cut -d'"' -f4)

echo "🌍 Testing Different Zoho Data Centers & API Configurations"
echo "=========================================================="
echo ""

# Different data center domains
DOMAINS=(
    "https://www.zohoapis.com"
    "https://www.zohoapis.eu" 
    "https://www.zohoapis.in"
    "https://www.zohoapis.com.au"
    "https://www.zohoapis.jp"
)

# Test basic connectivity to each domain
echo "📡 Testing data center connectivity..."
for domain in "${DOMAINS[@]}"; do
    echo "Testing $domain..."
    response=$(curl -s --connect-timeout 5 "$domain/workdrive/api/v1/files" \
        -H "Authorization: Zoho-oauthtoken $ACCESS_TOKEN" \
        -H "Accept: application/json")
    
    if echo "$response" | grep -q "INVALID_TICKET"; then
        echo "  ✅ Connected (auth error expected)"
    elif echo "$response" | grep -q "errors"; then
        error=$(echo "$response" | grep -o '"title":"[^"]*"' | cut -d'"' -f4)
        echo "  ⚠️  Connected but error: $error"
    else
        echo "  ❌ No connection or unexpected response"
    fi
done

echo ""
echo "🔍 Testing with different token formats..."

# Test token refresh to get a new one
echo "📋 Attempting token refresh..."
CLIENT_ID="1000.Z520MJ3HS00YJEKRHRX0U9KGZTATPX"
CLIENT_SECRET="731702ae155269b29c1997664def3553764face6f8"
REFRESH_TOKEN=$(grep -o '"refresh_token":"[^"]*"' zoho_tokens.json | cut -d'"' -f4)

refresh_response=$(curl -s -X POST "https://accounts.zoho.com/oauth/v2/token" \
    -H "Content-Type: application/x-www-form-urlencoded" \
    -d "grant_type=refresh_token" \
    -d "client_id=$CLIENT_ID" \
    -d "client_secret=$CLIENT_SECRET" \
    -d "refresh_token=$REFRESH_TOKEN")

echo "Refresh response:"
echo "$refresh_response"

# Check if we got a new token
if echo "$refresh_response" | grep -q "access_token"; then
    NEW_ACCESS_TOKEN=$(echo "$refresh_response" | grep -o '"access_token":"[^"]*"' | cut -d'"' -f4)
    echo ""
    echo "✅ Got new access token: ${NEW_ACCESS_TOKEN:0:20}..."
    
    # Test with new token
    echo "🧪 Testing with refreshed token..."
    curl -s "https://www.zohoapis.com/workdrive/api/v1/files" \
        -H "Authorization: Zoho-oauthtoken $NEW_ACCESS_TOKEN" \
        -H "Accept: application/json" | python3 -m json.tool 2>/dev/null || echo "Raw response received"
else
    echo "❌ Token refresh failed"
fi

echo ""
echo "📝 Checking Zoho WorkDrive API documentation requirements..."
echo "Possible issues:"
echo "- Wrong data center (US/EU/IN/AU)"
echo "- Insufficient scopes in OAuth app"  
echo "- API version mismatch"
echo "- Account-specific WorkDrive setup needed"