#!/bin/bash

echo "=================================="
echo "  SMS Notification System Startup"
echo "=================================="
echo ""

# Build frontend if dist doesn't exist
if [ ! -d "web/dist" ]; then
    echo "Building frontend..."
    cd web && npm run build && cd ..
fi

# Build Go server
echo "Building server..."
go build -o sms-server .

# Start the server
echo ""
echo "Starting SMS server on http://localhost:8080"
echo "Login: admin / admin123"
echo ""
./sms-server &
SERVER_PID=$!
sleep 2

# Start Cloudflare Tunnel
echo ""
echo "Starting Cloudflare Tunnel..."
echo "Your public URL will appear below:"
echo ""
cloudflared tunnel --url http://localhost:8080

# Cleanup on exit
kill $SERVER_PID 2>/dev/null
