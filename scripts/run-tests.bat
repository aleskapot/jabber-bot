@echo off
setlocal enabledelayedexpansion

echo 🧪 Running All Tests for Jabber Bot
echo ===================================

:: Check if Go is available
where go >nul 2>nul
if %ERRORLEVEL% NEQ 0 (
    echo ❌ Go is required but not installed.
    exit /b 1
)

:: Create test directories
if not exist "test\reports" mkdir "test\reports"

echo 📋 Running Unit Tests...

:: Run all unit tests
echo.
echo 🔍 Running Go unit tests...
go test -v -timeout 30s ./... -coverprofile=test/reports/unit-coverage.out
if %ERRORLEVEL% NEQ 0 (
    echo ❌ Unit tests failed
    exit /b 1
)

echo.
echo ✅ Unit tests completed successfully!

:: Ask user if they want to run integration tests
echo.
set /p run_integration="🔌 Run integration tests? (y/N): "
if /i "!run_integration!"=="y" (
    echo.
    echo 🌐 Running Integration Tests...
    
    :: Set environment variables for integration tests
    set INTEGRATION_TESTS=1
    set JABBER_BOT_LOG_LEVEL=debug
    
    :: Create temp directory for integration tests
    if not exist "tmp" mkdir "tmp"
    
    echo.
    echo 🧪 Integration Tests:
    go test -v -tags=integration -timeout 60s ./test/integration/... -coverprofile=test/reports/integration-coverage.out
    if %ERRORLEVEL% NEQ 0 (
        echo ❌ Integration tests failed
        exit /b 1
    )
    
    echo.
    echo ✅ Integration tests completed successfully!
    
    :: Combine coverage reports
    echo.
    echo 📊 Combining coverage reports...
    go tool cover -merge=test/reports/unit-coverage.out test/reports/integration-coverage.out -o test/reports/combined-coverage.out
    
    :: Generate HTML coverage report
    echo.
    echo 📈 Generating combined HTML coverage report...
    go tool cover -html=test/reports/combined-coverage.out -o test/reports/combined-coverage.html
    
    :: Display coverage summary
    echo.
    echo 📈 Combined Test Coverage Summary:
    go tool cover -func=test/reports/combined-coverage.out
    
    echo.
    echo 📊 Combined Coverage Report: test\reports\combined-coverage.html
    
    :: Cleanup temp directory
    if exist "tmp" rmdir /s /q "tmp"
) else (
    echo.
    echo 📈 Unit Test Coverage Summary:
    go tool cover -func=test/reports/unit-coverage.out
    
    :: Generate HTML coverage report for unit tests only
    echo.
    echo 📈 Generating unit test HTML coverage report...
    go tool cover -html=test/reports/unit-coverage.out -o test/reports/unit-coverage.html
    
    echo.
    echo 📊 Unit Coverage Report: test\reports\unit-coverage.html
)

echo.
echo 🎉 Test suite completed successfully!
echo 📋 All reports available in: test\reports\
pause