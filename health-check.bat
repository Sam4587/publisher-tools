@echo off
chcp 65001 >nul
title Publisher Tools - Health Check
color 0B

echo.
echo ========================================
echo   Publisher Tools - Health Check
echo ========================================
echo.

REM Check Hotspot Server (port 8080)
echo [1/3] Checking Hotspot Server (port 8080)...
curl -s http://localhost:8080/api/health >nul 2>&1
if %errorLevel% neq 0 (
    echo     [ERROR] Hotspot Server is not responding
    set hotspot_status=DOWN
) else (
    echo     [OK] Hotspot Server is running
    set hotspot_status=UP
)

REM Check Node Backend (port 3001)
echo [2/3] Checking Node Backend (port 3001)...
curl -s http://localhost:3001/api/health >nul 2>&1
if %errorLevel% neq 0 (
    echo     [ERROR] Node Backend is not responding
    set backend_status=DOWN
) else (
    echo     [OK] Node Backend is running
    set backend_status=UP
)

REM Check Frontend (port 5173)
echo [3/3] Checking Frontend (port 5173)...
curl -s http://localhost:5173 >nul 2>&1
if %errorLevel% neq 0 (
    echo     [ERROR] Frontend is not responding
    set frontend_status=DOWN
) else (
    echo     [OK] Frontend is running
    set frontend_status=UP
)

echo.
echo ========================================
echo   Service Status Summary
echo ========================================
echo.
echo  Hotspot Server (8080):  %hotspot_status%
echo  Node Backend (3001):    %backend_status%
echo  Frontend (5173):        %frontend_status%
echo.

if "%hotspot_status%"=="UP" if "%backend_status%"=="UP" if "%frontend_status%"=="UP" (
    echo [SUCCESS] All services are running normally
    echo.
    echo  Access URLs:
    echo    Frontend:       http://localhost:5173
    echo    Hotspot API:    http://localhost:8080/api/hot-topics
    echo    Backend API:    http://localhost:3001/api/health
) else (
    echo [WARNING] Some services are not running
    echo.
    echo  Troubleshooting:
    echo    1. Run 'start-stable.bat' to start all services
    echo    2. Check logs in the 'logs' directory
    echo    3. Verify no port conflicts exist
)

echo.
pause
