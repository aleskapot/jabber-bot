@echo off
setlocal enabledelayedexpansion

echo 🚀 Building Jabber Bot
echo =====================

:: Check if Go is available
where go >nul 2>nul
if %ERRORLEVEL% NEQ 0 (
    echo ❌ Go is required but not installed.
    exit /b 1
)

:: Create bin directory
if not exist "bin" mkdir "bin"

:: Get version from environment or use default
if "%JABBER_BOT_VERSION%"=="" (
    set JABBER_BOT_VERSION=1.0.0
)

:: Get build time
for /f "tokens=2 delims==" %%a in ('wmic OS Get localdatetime /value') do set "dt=%%a"
set "build_time=!dt:~0,4!-!dt:~4,2!-!dt:~6,2!T!dt:~8,2!:!dt:~10,2!:!dt:~12,2!"

echo 🔨 Building version %JABBER_BOT_VERSION%...
echo 📅 Build time: !build_time!

:: Build flags
set build_flags=-ldflags "-X main.Version=%JABBER_BOT_VERSION% -X main.BuildTime=!build_time! -X main.GitCommit=$(git rev-parse --short HEAD 2>nul || echo unknown)"

echo.
echo 🏗️  Building for Windows (amd64)...
go build %build_flags% -o bin/jabber-bot.exe ./cmd/server
if %ERRORLEVEL% NEQ 0 (
    echo ❌ Build failed for Windows
    exit /b 1
)

echo ✅ Windows build successful: bin\jabber-bot.exe

:: Build for Linux if requested
if "%1"=="--all" (
    echo.
    echo 🐧 Building for Linux (amd64)...
    set GOOS=linux
    set GOARCH=amd64
    go build %build_flags% -o bin/jabber-bot-linux-amd64 ./cmd/server
    if %ERRORLEVEL% NEQ 0 (
        echo ❌ Build failed for Linux
        exit /b 1
    )
    echo ✅ Linux build successful: bin\jabber-bot-linux-amd64
    
    echo.
    echo 🐧 Building for Linux (arm64)...
    set GOOS=linux
    set GOARCH=arm64
    go build %build_flags% -o bin/jabber-bot-linux-arm64 ./cmd/server
    if %ERRORLEVEL% NEQ 0 (
        echo ❌ Build failed for Linux ARM64
        exit /b 1
    )
    echo ✅ Linux ARM64 build successful: bin\jabber-bot-linux-arm64
    
    echo.
    echo 🍎 Building for macOS (amd64)...
    set GOOS=darwin
    set GOARCH=amd64
    go build %build_flags% -o bin/jabber-bot-darwin-amd64 ./cmd/server
    if %ERRORLEVEL% NEQ 0 (
        echo ❌ Build failed for macOS
        exit /b 1
    )
    echo ✅ macOS build successful: bin\jabber-bot-darwin-amd64
    
    echo.
    echo 🍎 Building for macOS (arm64)...
    set GOOS=darwin
    set GOARCH=arm64
    go build %build_flags% -o bin/jabber-bot-darwin-arm64 ./cmd/server
    if %ERRORLEVEL% NEQ 0 (
        echo ❌ Build failed for macOS ARM64
        exit /b 1
    )
    echo ✅ macOS ARM64 build successful: bin\jabber-bot-darwin-arm64
    
    :: Reset environment
    set GOOS=windows
    set GOARCH=amd64
)

echo.
echo 📦 Build completed!
echo 📁 Binaries available in: bin\

:: List built files
echo.
echo 📋 Built files:
dir bin\jabber-bot*.exe 2>nul
dir bin\jabber-bot* 2>nul | findstr /v "Directory"

echo.
echo 🎉 Build successful!
pause