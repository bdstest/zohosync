#!/bin/bash

# Exchange Authorization Code for Access Tokens
CLIENT_ID="1000.Z520MJ3HS00YJEKRHRX0U9KGZTATPX"
CLIENT_SECRET="731702ae155269b29c1997664def3553764face6f8"
REDIRECT_URI="http://localhost:8080/callback"
AUTH_CODE="1000.0c92bb469087fbcfe14331e2f85819cb.fc35b99e2d7bd00223323446440f6e1e"

echo "🔄 Exchanging Authorization Code for Access Tokens"
echo "================================================"
echo ""
echo "📋 Request Details:"
echo "   Client ID: $CLIENT_ID"
echo "   Auth Code: ${AUTH_CODE:0:20}..."
echo "   Redirect URI: $REDIRECT_URI"
echo ""

# Exchange code for tokens
echo "🚀 Making token exchange request..."

curl -X POST "https://accounts.zoho.com/oauth/v2/token" \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=authorization_code" \
  -d "client_id=$CLIENT_ID" \
  -d "client_secret=$CLIENT_SECRET" \
  -d "redirect_uri=$REDIRECT_URI" \
  -d "code=$AUTH_CODE" \
  -v 2>&1 | tee token_response.json

echo ""
echo "📄 Response saved to: token_response.json"
echo ""

# Check if we got tokens
if grep -q "access_token" token_response.json; then
    echo "🎉 SUCCESS! Tokens received"
    echo ""
    echo "📋 Extracting tokens..."
    
    # Extract access token (simple grep method)
    ACCESS_TOKEN=$(grep -o '"access_token":"[^"]*"' token_response.json | cut -d'"' -f4)
    REFRESH_TOKEN=$(grep -o '"refresh_token":"[^"]*"' token_response.json | cut -d'"' -f4)
    
    echo "✅ Access Token: ${ACCESS_TOKEN:0:20}..."
    echo "✅ Refresh Token: ${REFRESH_TOKEN:0:20}..."
    
    # Save tokens for API testing
    cat > zoho_tokens.json << EOF
{
  "access_token": "$ACCESS_TOKEN",
  "refresh_token": "$REFRESH_TOKEN",
  "client_id": "$CLIENT_ID",
  "client_secret": "$CLIENT_SECRET"
}
EOF
    
    echo ""
    echo "💾 Tokens saved to: zoho_tokens.json"
    echo "🚀 Ready to test Zoho WorkDrive API!"
    
else
    echo "❌ Token exchange failed"
    echo "📋 Check the response above for error details"
fi