#!/bin/bash

# Simple Zoho OAuth Test Script
# Tests the OAuth flow with your actual credentials

CLIENT_ID="1000.LZP2I9FSBFLX42MQ3WTFT0EABRSYQW"
CLIENT_SECRET="2a146b72f5bfac3447df81f96db634cfa643370f3d"
REDIRECT_URI="http://localhost:8080/callback"
SCOPES="WorkDrive.files.ALL,WorkDrive.workspace.READ"

echo "🔐 Zoho OAuth 2.0 Authentication Test"
echo "===================================="
echo ""
echo "Your Zoho Credentials:"
echo "Client ID: $CLIENT_ID"
echo "Client Secret: ${CLIENT_SECRET:0:10}..." 
echo ""

# Step 1: Generate authorization URL
AUTH_URL="https://accounts.zoho.com/oauth/v2/auth"
FULL_AUTH_URL="${AUTH_URL}?response_type=code&client_id=${CLIENT_ID}&scope=${SCOPES}&redirect_uri=${REDIRECT_URI}&access_type=offline"

echo "📋 OAuth Flow Steps:"
echo "1. Visit this authorization URL in your browser:"
echo ""
echo "${FULL_AUTH_URL}"
echo ""
echo "2. Log in with your Zoho account"
echo "3. Authorize ZohoSync permissions"
echo "4. Copy the authorization code from the callback URL"
echo ""
echo "💡 The callback URL will look like:"
echo "   http://localhost:8080/callback?code=AUTHORIZATION_CODE_HERE"
echo ""
echo "🚀 Once you have the authorization code, we can exchange it for access tokens!"
echo ""
echo "⚠️  Note: This is a manual test since we can't open browsers in this environment."
echo "   In the full ZohoSync application, this would be automated."