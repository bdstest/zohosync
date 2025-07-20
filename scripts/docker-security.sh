#!/bin/bash

# ZohoSync Docker Security Testing Script
# Comprehensive security testing using Docker containers

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")

cd "${PROJECT_ROOT}"

echo "🐳 ZohoSync Docker Security Testing Suite"
echo "========================================"
echo "Timestamp: ${TIMESTAMP}"
echo "Project: ${PROJECT_ROOT}"
echo ""

# Function to check if Docker is available
check_docker() {
    if ! command -v docker >/dev/null 2>&1; then
        echo "❌ Docker is not installed or not in PATH"
        echo "   Please install Docker: https://docs.docker.com/get-docker/"
        exit 1
    fi
    
    if ! docker info >/dev/null 2>&1; then
        echo "❌ Docker daemon is not running"
        echo "   Please start Docker daemon"
        exit 1
    fi
    
    echo "✅ Docker is available and running"
}

# Function to check if Docker Compose is available
check_docker_compose() {
    if ! command -v docker-compose >/dev/null 2>&1; then
        echo "❌ Docker Compose is not installed or not in PATH"
        echo "   Please install Docker Compose: https://docs.docker.com/compose/install/"
        exit 1
    fi
    
    echo "✅ Docker Compose is available"
}

# Function to build security Docker image
build_security_image() {
    echo ""
    echo "🔨 Building ZohoSync Security Docker Image..."
    echo "============================================="
    
    docker build -f Dockerfile.security \
        --target security-scan \
        --tag zohosync:security-latest \
        --tag zohosync:security-${TIMESTAMP} \
        .
    
    echo "✅ Security Docker image built successfully"
}

# Function to run comprehensive security scan
run_security_scan() {
    echo ""
    echo "🔍 Running Comprehensive Security Scan..."
    echo "========================================"
    
    # Ensure reports directory exists
    mkdir -p security/reports
    
    # Run security scan using Docker Compose
    docker-compose -f docker-compose.security.yml run --rm zohosync-security
    
    echo "✅ Security scan completed"
    echo "📋 Reports available in: security/reports/"
}

# Function to run quick security check
run_quick_security() {
    echo ""
    echo "⚡ Running Quick Security Check..."
    echo "================================"
    
    docker-compose -f docker-compose.security.yml run --rm zohosync-security-quick
    
    echo "✅ Quick security check completed"
}

# Function to test build in Docker
test_build() {
    echo ""
    echo "🔨 Testing Build in Docker Environment..."
    echo "======================================="
    
    docker-compose -f docker-compose.security.yml run --rm zohosync-build
    
    echo "✅ Build test completed"
}

# Function to start development environment
start_dev_environment() {
    echo ""
    echo "🛠️  Starting Development Environment..."
    echo "====================================="
    
    docker-compose -f docker-compose.security.yml up -d zohosync-dev
    
    echo "✅ Development environment started"
    echo "🔗 Connect with: docker-compose -f docker-compose.security.yml exec zohosync-dev bash"
}

# Function to stop all services
stop_services() {
    echo ""
    echo "🛑 Stopping All Services..."
    echo "=========================="
    
    docker-compose -f docker-compose.security.yml down
    
    echo "✅ All services stopped"
}

# Function to clean up Docker resources
cleanup() {
    echo ""
    echo "🧹 Cleaning Up Docker Resources..."
    echo "================================="
    
    # Stop and remove containers
    docker-compose -f docker-compose.security.yml down --remove-orphans
    
    # Remove unused images (optional)
    read -p "Remove unused Docker images? (y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        docker image prune -f
        echo "✅ Unused images removed"
    fi
    
    echo "✅ Cleanup completed"
}

# Function to show logs
show_logs() {
    echo ""
    echo "📋 Showing Service Logs..."
    echo "========================="
    
    docker-compose -f docker-compose.security.yml logs -f
}

# Function to display usage information
usage() {
    echo "Usage: $0 [COMMAND]"
    echo ""
    echo "Commands:"
    echo "  scan          Run comprehensive security scan"
    echo "  quick         Run quick security check"
    echo "  build         Test build in Docker environment"
    echo "  dev           Start development environment"
    echo "  stop          Stop all services"
    echo "  logs          Show service logs"
    echo "  cleanup       Clean up Docker resources"
    echo "  help          Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0 scan       # Run full security analysis"
    echo "  $0 quick      # Quick security check"
    echo "  $0 dev        # Start dev environment"
    echo ""
}

# Main execution logic
main() {
    check_docker
    check_docker_compose
    
    case "${1:-scan}" in
        "scan")
            build_security_image
            run_security_scan
            ;;
        "quick")
            run_quick_security
            ;;
        "build")
            test_build
            ;;
        "dev")
            start_dev_environment
            ;;
        "stop")
            stop_services
            ;;
        "logs")
            show_logs
            ;;
        "cleanup")
            cleanup
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