#!/bin/bash

# OAuth Flow Test with New Credentials
CLIENT_ID="1000.Z520MJ3HS00YJEKRHRX0U9KGZTATPX"
CLIENT_SECRET="731702ae155269b29c1997664def3553764face6f8"
REDIRECT_URI="http://localhost:8080/callback"
SCOPES="WorkDrive.files.ALL,WorkDrive.workspace.READ,WorkDrive.organization.READ"

echo "🔐 ZohoSync OAuth 2.0 Test - ZHSyncTest App"
echo "==========================================="
echo ""
echo "✅ Client ID: $CLIENT_ID"
echo "✅ Redirect URI: $REDIRECT_URI"
echo ""

# Generate authorization URL
AUTH_URL="https://accounts.zoho.com/oauth/v2/auth"
FULL_AUTH_URL="${AUTH_URL}?response_type=code&client_id=${CLIENT_ID}&scope=${SCOPES}&redirect_uri=${REDIRECT_URI}&access_type=offline"

echo "🚀 STEP 1: Visit this authorization URL:"
echo ""
echo "${FULL_AUTH_URL}"
echo ""
echo "📋 STEP 2: After authorization, you'll be redirected to:"
echo "   http://localhost:8080/callback?code=AUTHORIZATION_CODE"
echo ""
echo "💡 STEP 3: If the callback fails (network issue), look for the code in the URL"
echo "   and paste it here manually."
echo ""

# Start simple HTTP server to catch the callback
echo "🌐 Starting callback server on port 8080..."
python3 -c "
import http.server
import socketserver
from urllib.parse import parse_qs, urlparse

class OAuthHandler(http.server.BaseHTTPRequestHandler):
    def do_GET(self):
        if self.path.startswith('/callback'):
            parsed_url = urlparse(self.path)
            params = parse_qs(parsed_url.query)
            
            if 'code' in params:
                auth_code = params['code'][0]
                print(f'\\n🎉 SUCCESS! Authorization code: {auth_code}')
                
                self.send_response(200)
                self.send_header('Content-type', 'text/html')
                self.end_headers()
                self.wfile.write(b'<h1>Success! Authorization received. You can close this window.</h1>')
            else:
                self.send_response(400)
                self.end_headers()

try:
    with socketserver.TCPServer(('0.0.0.0', 8080), OAuthHandler) as httpd:
        print('⏳ Waiting for OAuth callback... (Ctrl+C to stop)')
        httpd.serve_forever()
except KeyboardInterrupt:
    print('\\n🛑 Server stopped')
"