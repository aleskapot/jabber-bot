@echo off
setlocal enabledelayedexpansion

:: Docker deployment script for Jabber Bot (Windows)
set "PROJECT_NAME=jabber-bot"
set "DOCKER_COMPOSE_FILE=docker-compose.yml"

:help
echo Jabber Bot Docker Deployment Script
echo ==================================
echo.
echo Usage: %0 [COMMAND] [OPTIONS]
echo.
echo COMMANDS:
echo     dev         Start development environment
echo     prod        Start production environment
echo     build       Build Docker images
echo     stop        Stop running containers
echo     restart     Restart containers
echo     logs        Show container logs
echo     status      Show container status
echo     clean       Clean up containers and volumes
echo.
echo OPTIONS:
echo     --help      Show this help message
echo.
goto :eof

:check_requirements
:: Check if Docker is available
docker --version >nul 2>&1
if %ERRORLEVEL% NEQ 0 (
    echo ❌ Docker is required but not installed
    exit /b 1
)

:: Check if Docker Compose is available
docker-compose --version >nul 2>&1
if %ERRORLEVEL% NEQ 0 (
    echo ❌ Docker Compose is required but not installed
    exit /b 1
)
goto :eof

:check_env_file
if not exist ".env" (
    if exist ".env.example" (
        echo ⚠️  .env file not found, copying from .env.example
        copy .env.example .env >nul
        echo ⚠️  Please edit .env file with your configuration
        exit /b 1
    ) else (
        echo ❌ .env file not found and no .env.example available
        exit /b 1
    )
)

:: Load environment variables (simple check)
for /f "tokens=1,2 delims==" %%a in (.env) do (
    if "%%a"=="JABBER_BOT_XMPP_JID" set "XMPP_JID=%%b"
    if "%%a"=="JABBER_BOT_XMPP_PASSWORD" set "XMPP_PASSWORD=%%b"
    if "%%a"=="JABBER_BOT_XMPP_SERVER" set "XMPP_SERVER=%%b"
    if "%%a"=="JABBER_BOT_WEBHOOK_URL" set "WEBHOOK_URL=%%b"
)

:: Check required environment variables
if "%XMPP_JID%"=="" (
    echo ❌ JABBER_BOT_XMPP_JID is required
    exit /b 1
)
if "%XMPP_PASSWORD%"=="" (
    echo ❌ JABBER_BOT_XMPP_PASSWORD is required
    exit /b 1
)
if "%XMPP_SERVER%"=="" (
    echo ❌ JABBER_BOT_XMPP_SERVER is required
    exit /b 1
)
if "%WEBHOOK_URL%"=="" (
    echo ❌ JABBER_BOT_WEBHOOK_URL is required
    exit /b 1
)
goto :eof

:setup_environment
set "ENV_TYPE=%1"
if "%ENV_TYPE%"=="dev" (
    set "DOCKER_COMPOSE_FILE=docker-compose.dev.yml"
    set "PROJECT_NAME=jabber-bot-dev"
    echo 🔧 Setting up development environment
)
if "%ENV_TYPE%"=="prod" (
    set "DOCKER_COMPOSE_FILE=docker-compose.prod.yml"
    set "PROJECT_NAME=jabber-bot-prod"
    echo 🏭 Setting up production environment
)
goto :eof

:build_images
echo 🏗️  Building Docker images...
docker-compose -f %DOCKER_COMPOSE_FILE% build
if %ERRORLEVEL% NEQ 0 (
    echo ❌ Build failed
    exit /b 1
)
echo ✅ Docker images built successfully
goto :eof

:start_services
echo 🚀 Starting services...
docker-compose -f %DOCKER_COMPOSE_FILE% up -d
if %ERRORLEVEL% NEQ 0 (
    echo ❌ Failed to start services
    exit /b 1
)
echo ✅ Services started successfully
goto :show_status

:show_logs
if "%2"=="" (
    docker-compose -f %DOCKER_COMPOSE_FILE% logs -f
) else (
    docker-compose -f %DOCKER_COMPOSE_FILE% logs -f %2
)
goto :eof

:show_status
echo 📊 Container status:
docker-compose -f %DOCKER_COMPOSE_FILE% ps
goto :eof

:stop_services
echo 🛑 Stopping services...
docker-compose -f %DOCKER_COMPOSE_FILE% down
echo ✅ Services stopped
goto :eof

:restart_services
echo 🔄 Restarting services...
docker-compose -f %DOCKER_COMPOSE_FILE% restart
echo ✅ Services restarted
goto :show_status

:clean_all
echo ⚠️  This will remove all containers, networks, and volumes
set /p confirm="Are you sure? [y/N] "
if /i "!confirm!"=="y" (
    echo 🧹 Cleaning up...
    docker-compose -f %DOCKER_COMPOSE_FILE% down -v --remove-orphans
    docker system prune -f
    echo ✅ Cleanup completed
) else (
    echo ❌ Cleanup cancelled
)
goto :eof

:main
set "COMMAND=%1"
set "SERVICE=%2"

:: Check requirements
call :check_requirements

:: Execute command
if "%COMMAND%"=="dev" (
    call :setup_environment "dev"
    call :check_env_file
    call :build_images
    call :start_services
) else if "%COMMAND%"=="prod" (
    call :setup_environment "prod"
    call :check_env_file
    call :build_images
    call :start_services
) else if "%COMMAND%"=="build" (
    call :build_images
) else if "%COMMAND%"=="stop" (
    call :setup_environment "dev"
    call :stop_services
) else if "%COMMAND%"=="restart" (
    call :setup_environment "dev"
    call :restart_services
) else if "%COMMAND%"=="logs" (
    call :setup_environment "dev"
    call :show_logs "%COMMAND%" "%SERVICE%"
) else if "%COMMAND%"=="status" (
    call :setup_environment "dev"
    call :show_status
) else if "%COMMAND%"=="clean" (
    call :setup_environment "dev"
    call :clean_all
) else if "%COMMAND%"=="--help" (
    call :help
) else (
    echo ❌ Unknown command: %COMMAND%
    call :help
    exit /b 1
)

goto :eof

:: Run main function with all arguments
call :main %*