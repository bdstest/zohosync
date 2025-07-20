#!/bin/bash

# ZohoSync GUI Docker Setup Script
# Configures X11 forwarding and runs GUI applications in Docker

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

cd "${PROJECT_ROOT}"

echo "🖥️  ZohoSync GUI Docker Setup"
echo "============================"
echo ""

# Function to detect display server
detect_display_server() {
    if [ -n "$WAYLAND_DISPLAY" ]; then
        echo "🔍 Detected Wayland display server"
        export DISPLAY_SERVER="wayland"
    elif [ -n "$DISPLAY" ]; then
        echo "🔍 Detected X11 display server"
        export DISPLAY_SERVER="x11"
    else
        echo "⚠️  No display server detected, attempting X11 setup"
        export DISPLAY=":0"
        export DISPLAY_SERVER="x11"
    fi
}

# Function to setup X11 permissions
setup_x11_permissions() {
    echo "🔧 Setting up X11 permissions..."
    
    # Allow connections to X server
    if command -v xhost >/dev/null 2>&1; then
        xhost +local:docker 2>/dev/null || true
        xhost +local:root 2>/dev/null || true
        echo "✅ X11 permissions configured"
    else
        echo "⚠️  xhost not found, GUI may not work properly"
        echo "   Install: sudo apt-get install x11-xserver-utils"
    fi
}

# Function to setup environment variables
setup_environment() {
    echo "🔧 Setting up environment variables..."
    
    # Set user ID and group ID for Docker
    export USER_ID=$(id -u)
    export GROUP_ID=$(id -g)
    export UID=${USER_ID}
    export GID=${GROUP_ID}
    
    # Set display environment
    export DISPLAY=${DISPLAY:-:0}
    export XDG_RUNTIME_DIR=${XDG_RUNTIME_DIR:-/tmp}
    
    echo "   USER_ID: $USER_ID"
    echo "   GROUP_ID: $GROUP_ID"
    echo "   DISPLAY: $DISPLAY"
    echo "   XDG_RUNTIME_DIR: $XDG_RUNTIME_DIR"
    echo "✅ Environment configured"
}

# Function to test X11 connection
test_x11_connection() {
    echo "🧪 Testing X11 connection..."
    
    if command -v xdpyinfo >/dev/null 2>&1; then
        if xdpyinfo >/dev/null 2>&1; then
            echo "✅ X11 connection working"
            return 0
        else
            echo "❌ X11 connection failed"
            return 1
        fi
    else
        echo "⚠️  xdpyinfo not found, cannot test X11 connection"
        echo "   Install: sudo apt-get install x11-utils"
        return 0
    fi
}

# Function to build Docker images
build_docker_images() {
    echo ""
    echo "🔨 Building Docker images with GUI support..."
    echo "============================================"
    
    docker compose -f docker compose.gui.yml build
    
    echo "✅ Docker images built successfully"
}

# Function to start GUI development environment
start_gui_dev() {
    echo ""
    echo "🚀 Starting GUI development environment..."
    echo "========================================"
    
    docker compose -f docker compose.gui.yml up -d zohosync-gui
    
    echo "✅ GUI development environment started"
    echo ""
    echo "🔗 Connect with:"
    echo "   docker compose -f docker compose.gui.yml exec zohosync-gui bash"
    echo ""
    echo "📱 To test GUI application:"
    echo "   docker compose -f docker compose.gui.yml exec zohosync-gui make build-gui"
    echo "   docker compose -f docker compose.gui.yml exec zohosync-gui ./zohosync"
}

# Function to test GUI application
test_gui_application() {
    echo ""
    echo "🧪 Testing GUI application in Docker..."
    echo "======================================"
    
    echo "Building GUI application..."
    docker compose -f docker compose.gui.yml run --rm zohosync-gui-test
    
    echo ""
    echo "🎯 Attempting to run GUI application..."
    docker compose -f docker compose.gui.yml run --rm zohosync-gui bash -c "
        echo 'Building ZohoSync GUI...'
        make build-gui
        echo 'GUI application built successfully!'
        echo 'To run GUI: ./zohosync'
        echo 'Note: GUI requires display server connection'
    "
    
    echo "✅ GUI application test completed"
}

# Function to run security scan
run_security_scan() {
    echo ""
    echo "🔍 Running comprehensive security scan..."
    echo "======================================="
    
    docker compose -f docker compose.security.yml run --rm zohosync-security
    
    echo "✅ Security scan completed"
    echo "📋 Reports available in: security/reports/"
}

# Function to show usage
usage() {
    echo "Usage: $0 [COMMAND]"
    echo ""
    echo "Commands:"
    echo "  setup         Setup GUI environment and permissions"
    echo "  start         Start GUI development environment"
    echo "  test          Test GUI application build"
    echo "  security      Run security scan"
    echo "  interactive   Start interactive GUI environment"
    echo "  stop          Stop all services"
    echo "  help          Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0 setup       # Setup GUI environment"
    echo "  $0 start       # Start GUI dev environment"
    echo "  $0 test        # Test GUI application"
    echo "  $0 interactive # Start interactive session"
    echo ""
}

# Function to start interactive session
start_interactive() {
    echo ""
    echo "🎮 Starting interactive GUI session..."
    echo "===================================="
    
    docker compose -f docker compose.gui.yml run --rm zohosync-gui
    
    echo "✅ Interactive session ended"
}

# Function to stop all services
stop_services() {
    echo ""
    echo "🛑 Stopping all GUI services..."
    echo "=============================="
    
    docker compose -f docker compose.gui.yml down
    docker compose -f docker compose.security.yml down
    
    echo "✅ All services stopped"
}

# Main execution logic
main() {
    detect_display_server
    
    case "${1:-setup}" in
        "setup")
            setup_x11_permissions
            setup_environment
            test_x11_connection
            build_docker_images
            echo ""
            echo "✅ GUI Docker environment setup complete!"
            echo "🚀 Next: Run '$0 start' to launch development environment"
            ;;
        "start")
            setup_environment
            start_gui_dev
            ;;
        "test")
            setup_environment
            test_gui_application
            ;;
        "security")
            run_security_scan
            ;;
        "interactive")
            setup_environment
            start_interactive
            ;;
        "stop")
            stop_services
            ;;
        "help"|"-h"|"--help")
            usage
            ;;
        *)
            echo "❌ Unknown command: $1"
            echo ""
            usage
            exit 1
            ;;
    esac
}

# Execute main function with all arguments
main "$@"