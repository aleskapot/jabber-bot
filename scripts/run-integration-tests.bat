@echo off
setlocal enabledelayedexpansion

echo 🧪 Starting Integration Tests for Jabber Bot
echo ==========================================

:: Check if Go is available
where go >nul 2>nul
if %ERRORLEVEL% NEQ 0 (
    echo ❌ Go is required but not installed.
    exit /b 1
)

:: Set environment variables for integration tests
set INTEGRATION_TESTS=1
set JABBER_BOT_LOG_LEVEL=debug

:: Create test directories
if not exist "tmp\jabber-bot-tests" mkdir "tmp\jabber-bot-tests"
if not exist "test\reports" mkdir "test\reports"

echo 📋 Running Integration Tests...

:: Run integration tests with coverage
echo.
echo 🔧 Configuration Integration Tests...
go test -v -tags=integration -timeout 30s ./test/integration/config_integration_test.go -coverprofile=test/reports/config-integration-coverage.out
if %ERRORLEVEL% NEQ 0 (
    echo ❌ Configuration integration tests failed
    exit /b 1
)

echo.
echo 🔌 Webhook Integration Tests...
go test -v -tags=integration -timeout 60s ./test/integration/webhook_integration_test.go -coverprofile=test/reports/webhook-integration-coverage.out
if %ERRORLEVEL% NEQ 0 (
    echo ❌ Webhook integration tests failed
    exit /b 1
)

echo.
echo 🌐 Full Integration Tests...
go test -v -tags=integration -timeout 60s ./test/integration/integration_test.go -coverprofile=test/reports/full-integration-coverage.out
if %ERRORLEVEL% NEQ 0 (
    echo ❌ Full integration tests failed
    exit /b 1
)

:: Combine coverage reports
echo.
echo 📊 Combining coverage reports...
go tool cover -merge=test/reports/*-integration-coverage.out -o test/reports/integration-coverage.out
if %ERRORLEVEL% NEQ 0 (
    echo ⚠️  Could not merge coverage reports
)

:: Generate HTML coverage report
echo.
echo 📈 Generating HTML coverage report...
go tool cover -html=test/reports/integration-coverage.out -o test/reports/integration-coverage.html
if %ERRORLEVEL% NEQ 0 (
    echo ⚠️  Could not generate HTML coverage report
)

:: Display coverage summary
echo.
echo 📈 Integration Test Coverage Summary:
go tool cover -func=test/reports/integration-coverage.out
if %ERRORLEVEL% NEQ 0 (
    echo ⚠️  Could not display coverage summary
)

:: Run performance tests if requested
if "%1"=="--performance" (
    echo.
    echo ⚡ Running Performance Tests...
    go test -v -tags=integration -bench=. ./test/integration/... -benchmem
)

:: Cleanup test artifacts
echo.
echo 🧹 Cleaning up test artifacts...
if exist "tmp\jabber-bot-tests" rmdir /s /q "tmp\jabber-bot-tests"

echo.
echo ✅ Integration Tests Completed Successfully!
echo 📊 Coverage Report: test\reports\integration-coverage.html
echo 📋 Detailed Reports: test\reports\
echo.
echo 🎉 All integration tests passed!
pause