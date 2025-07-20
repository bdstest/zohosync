#!/bin/bash

# OAuth Callback Server for LXC/Docker Environment
# Handles the OAuth redirect and captures the authorization code

PORT=8080
echo "🚀 Starting OAuth callback server on port $PORT..."
echo "📡 Server will be accessible at:"
echo "   - Inside LXC: http://localhost:$PORT/callback"
echo "   - From host: http://[LXC-IP]:$PORT/callback"
echo ""

# Get LXC IP address
LXC_IP=$(hostname -I | awk '{print $1}')
echo "📍 LXC IP Address: $LXC_IP"
echo "🔗 External access: http://$LXC_IP:$PORT/callback"
echo ""

# Start simple callback server
python3 -c "
import http.server
import socketserver
from urllib.parse import parse_qs, urlparse
import json

class OAuthHandler(http.server.BaseHTTPRequestHandler):
    def do_GET(self):
        if self.path.startswith('/callback'):
            # Parse the authorization code
            parsed_url = urlparse(self.path)
            params = parse_qs(parsed_url.query)
            
            if 'code' in params:
                auth_code = params['code'][0]
                print('\\n🎉 SUCCESS! Authorization code received:')
                print(f'   Code: {auth_code}')
                print('\\n📋 Next steps:')
                print('   1. Copy the authorization code above')
                print('   2. Stop this server (Ctrl+C)')
                print('   3. Exchange code for access tokens')
                
                # Send success response
                self.send_response(200)
                self.send_header('Content-type', 'text/html')
                self.end_headers()
                self.wfile.write(b'''
                <html><body>
                <h1>🎉 Authorization Successful!</h1>
                <p>Authorization code received. You can close this window.</p>
                <p>Return to the terminal to continue.</p>
                </body></html>
                ''')
            else:
                # Error case
                self.send_response(400)
                self.send_header('Content-type', 'text/html')
                self.end_headers()
                error = params.get('error', ['Unknown error'])[0]
                self.wfile.write(f'<h1>❌ OAuth Error: {error}</h1>'.encode())
        else:
            self.send_response(404)
            self.end_headers()
    
    def log_message(self, format, *args):
        pass  # Suppress default logging

with socketserver.TCPServer(('0.0.0.0', $PORT), OAuthHandler) as httpd:
    print(f'🌐 OAuth callback server running on http://0.0.0.0:$PORT')
    print('⏳ Waiting for OAuth callback...')
    print('   (Use Ctrl+C to stop)')
    try:
        httpd.serve_forever()
    except KeyboardInterrupt:
        print('\\n🛑 Server stopped')
"