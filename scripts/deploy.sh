#!/bin/bash

# Docker deployment script for Jabber Bot
# Supports development, production, and monitoring setups

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
DOCKER_COMPOSE_FILE="docker-compose.yml"
PROJECT_NAME="jabber-bot"

# Functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

show_help() {
    cat << EOF
Jabber Bot Docker Deployment Script

Usage: $0 [COMMAND] [OPTIONS]

COMMANDS:
    dev         Start development environment
    prod        Start production environment
    build       Build Docker images
    stop        Stop running containers
    restart     Restart containers
    logs        Show container logs
    status      Show container status
    clean       Clean up containers and volumes
    monitoring  Start monitoring stack (Prometheus + Grafana)

OPTIONS:
    --no-build  Skip building images
    --pull      Pull latest images
    --detach    Run in background (default)
    --help      Show this help message

ENVIRONMENT VARIABLES:
    JABBER_BOT_XMPP_JID          (Required) XMPP JID
    JABBER_BOT_XMPP_PASSWORD      (Required) XMPP password
    JABBER_BOT_XMPP_SERVER        (Required) XMPP server address
    JABBER_BOT_WEBHOOK_URL        (Required) Webhook endpoint URL

EXAMPLES:
    # Development with live reload
    $0 dev

    # Production deployment
    $0 prod --pull

    # Only monitoring stack
    $0 monitoring

    # Show logs
    $0 logs jabber-bot

    # Clean everything
    $0 clean

EOF
}

check_requirements() {
    if ! command -v docker &> /dev/null; then
        log_error "Docker is required but not installed"
        exit 1
    fi

    if ! command -v docker-compose &> /dev/null; then
        log_error "Docker Compose is required but not installed"
        exit 1
    fi
}

check_env_file() {
    if [ ! -f ".env" ]; then
        if [ -f ".env.example" ]; then
            log_warning ".env file not found, copying from .env.example"
            cp .env.example .env
            log_warning "Please edit .env file with your configuration"
            exit 1
        else
            log_error ".env file not found and no .env.example available"
            exit 1
        fi
    fi

    # Check required environment variables
    source .env
    local missing_vars=()

    [ -z "$JABBER_BOT_XMPP_JID" ] && missing_vars+=("JABBER_BOT_XMPP_JID")
    [ -z "$JABBER_BOT_XMPP_PASSWORD" ] && missing_vars+=("JABBER_BOT_XMPP_PASSWORD")
    [ -z "$JABBER_BOT_XMPP_SERVER" ] && missing_vars+=("JABBER_BOT_XMPP_SERVER")
    [ -z "$JABBER_BOT_WEBHOOK_URL" ] && missing_vars+=("JABBER_BOT_WEBHOOK_URL")

    if [ ${#missing_vars[@]} -gt 0 ]; then
        log_error "Missing required environment variables:"
        for var in "${missing_vars[@]}"; do
            echo "  - $var"
        done
        exit 1
    fi
}

setup_environment() {
    local env_type=$1
    
    case $env_type in
        "dev")
            DOCKER_COMPOSE_FILE="docker-compose.dev.yml"
            PROJECT_NAME="jabber-bot-dev"
            log_info "Setting up development environment"
            ;;
        "prod")
            DOCKER_COMPOSE_FILE="docker-compose.prod.yml"
            PROJECT_NAME="jabber-bot-prod"
            log_info "Setting up production environment"
            ;;
        "monitoring")
            # Use main compose file but only monitoring services
            log_info "Setting up monitoring stack"
            ;;
    esac
}

build_images() {
    if [ "$NO_BUILD" = "true" ]; then
        log_info "Skipping image build"
        return 0
    fi

    log_info "Building Docker images..."
    
    if [ "$PULL" = "true" ]; then
        docker-compose -f $DOCKER_COMPOSE_FILE pull
    fi
    
    docker-compose -f $DOCKER_COMPOSE_FILE build
    log_success "Docker images built successfully"
}

start_services() {
    local service=${2:-""}
    
    log_info "Starting services..."
    
    if [ -n "$service" ]; then
        docker-compose -f $DOCKER_COMPOSE_FILE up -d $service
    else
        docker-compose -f $DOCKER_COMPOSE_FILE up -d
    fi
    
    log_success "Services started successfully"
    show_status
}

show_logs() {
    local service=${2:-""}
    
    if [ -n "$service" ]; then
        docker-compose -f $DOCKER_COMPOSE_FILE logs -f $service
    else
        docker-compose -f $DOCKER_COMPOSE_FILE logs -f
    fi
}

show_status() {
    log_info "Container status:"
    docker-compose -f $DOCKER_COMPOSE_FILE ps
}

stop_services() {
    log_info "Stopping services..."
    docker-compose -f $DOCKER_COMPOSE_FILE down
    log_success "Services stopped"
}

restart_services() {
    log_info "Restarting services..."
    docker-compose -f $DOCKER_COMPOSE_FILE restart
    log_success "Services restarted"
    show_status
}

clean_all() {
    log_warning "This will remove all containers, networks, and volumes"
    read -p "Are you sure? [y/N] " -n 1 -r
    echo
    
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        log_info "Cleaning up..."
        docker-compose -f $DOCKER_COMPOSE_FILE down -v --remove-orphans
        docker system prune -f
        log_success "Cleanup completed"
    else
        log_info "Cleanup cancelled"
    fi
}

# Main script execution
main() {
    # Default options
    NO_BUILD=false
    PULL=false
    
    # Parse arguments
    COMMAND=""
    SERVICE=""
    
    while [[ $# -gt 0 ]]; do
        case $1 in
            dev|prod|monitoring|build|stop|restart|logs|status|clean)
                COMMAND="$1"
                shift
                ;;
            --no-build)
                NO_BUILD=true
                shift
                ;;
            --pull)
                PULL=true
                shift
                ;;
            --detach)
                # Already default behavior
                shift
                ;;
            --help|-h)
                show_help
                exit 0
                ;;
            *)
                if [ -z "$SERVICE" ] && [ "$COMMAND" = "logs" ]; then
                    SERVICE="$1"
                else
                    log_error "Unknown option: $1"
                    show_help
                    exit 1
                fi
                shift
                ;;
        esac
    done

    # Check requirements
    check_requirements

    # Execute command
    case $COMMAND in
        dev)
            setup_environment "dev"
            check_env_file
            build_images
            start_services "dev"
            ;;
        prod)
            setup_environment "prod"
            check_env_file
            build_images
            start_services "prod"
            ;;
        monitoring)
            setup_environment "monitoring"
            build_images
            start_services "prometheus grafana redis"
            ;;
        build)
            build_images
            ;;
        stop)
            setup_environment "dev" # Default to dev compose file
            stop_services
            ;;
        restart)
            setup_environment "dev" # Default to dev compose file
            restart_services
            ;;
        logs)
            setup_environment "dev" # Default to dev compose file
            show_logs "logs" "$SERVICE"
            ;;
        status)
            setup_environment "dev" # Default to dev compose file
            show_status
            ;;
        clean)
            setup_environment "dev" # Default to dev compose file
            clean_all
            ;;
        "")
            log_error "No command specified"
            show_help
            exit 1
            ;;
        *)
            log_error "Unknown command: $COMMAND"
            show_help
            exit 1
            ;;
    esac
}

# Run main function with all arguments
main "$@"