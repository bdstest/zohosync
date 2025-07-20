#!/bin/bash

# Comprehensive WorkDrive Diagnostics
ACCESS_TOKEN=$(grep -o '"access_token":"[^"]*"' zoho_tokens.json | cut -d'"' -f4)

echo "🏥 Zoho WorkDrive Account Diagnostics"
echo "===================================="
echo ""

echo "🔍 Step 1: Check if WorkDrive is enabled for your account"
echo "--------------------------------------------------------"

# Test Zoho Account API first (should work)
echo "Testing basic Zoho account access..."
curl -s "https://accounts.zoho.com/oauth/v2/users/me" \
    -H "Authorization: Zoho-oauthtoken $ACCESS_TOKEN" | python3 -m json.tool 2>/dev/null || echo "Account API test completed"

echo ""
echo "🔍 Step 2: Check WorkDrive specific endpoints"
echo "---------------------------------------------"

# Test if WorkDrive service is available
echo "Testing WorkDrive service availability..."

# Try the most basic WorkDrive endpoint
endpoints_to_test=(
    "https://www.zohoapis.com/workdrive/api/v1"
    "https://www.zohoapis.com/workdrive/api/v1/"
    "https://www.zohoapis.com/workdrive/home"
    "https://workdrive.zoho.com/api/v1"
)

for endpoint in "${endpoints_to_test[@]}"; do
    echo "Testing: $endpoint"
    response=$(curl -s -w "HTTP_CODE:%{http_code}" "$endpoint" \
        -H "Authorization: Zoho-oauthtoken $ACCESS_TOKEN" \
        -H "Accept: application/json")
    
    http_code=$(echo "$response" | grep -o "HTTP_CODE:[0-9]*" | cut -d: -f2)
    content=$(echo "$response" | sed 's/HTTP_CODE:[0-9]*$//')
    
    echo "  HTTP Code: $http_code"
    if [ ! -z "$content" ]; then
        echo "  Response: ${content:0:100}..."
    fi
    echo ""
done

echo "🔍 Step 3: Check OAuth scopes and permissions"
echo "---------------------------------------------"

# Check what our token can access
echo "Current token scopes from OAuth response:"
grep -o '"scope":"[^"]*"' token_response.json | cut -d'"' -f4

echo ""
echo "🔍 Step 4: Manual WorkDrive Check"
echo "--------------------------------"
echo "Please manually verify:"
echo "1. Login to https://workdrive.zoho.com"
echo "2. Confirm WorkDrive is enabled for your account"
echo "3. Check if you have any folders/files"
echo "4. Verify your account region (US/EU/IN/AU)"
echo ""

echo "🔍 Step 5: Alternative API Test"
echo "------------------------------"
echo "Testing with simpler endpoint structure..."

# Try Zoho's general API structure
simple_endpoints=(
    "https://www.zohoapis.com/workdrive"
    "https://www.zohoapis.com/workdrive/"
    "https://www.zohoapis.com/workdrive/home"
)

for endpoint in "${simple_endpoints[@]}"; do
    echo "Testing: $endpoint"
    curl -s -I "$endpoint" -H "Authorization: Zoho-oauthtoken $ACCESS_TOKEN" | grep -E "(HTTP|Location|Error)"
    echo ""
done

echo "💡 Diagnostic Summary:"
echo "====================="
echo "If all tests show authentication errors, possible causes:"
echo "1. WorkDrive not enabled for your Zoho account"
echo "2. Wrong data center region"
echo "3. OAuth app needs WorkDrive-specific configuration"
echo "4. Account permissions insufficient"
echo ""
echo "✅ Next step: Verify WorkDrive access at https://workdrive.zoho.com"