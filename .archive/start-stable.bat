@echo off
chcp 65001 >nul
title Publisher Tools - Stable Service Manager
color 0A

echo.
echo ========================================
echo   Publisher Tools - Stable Service Manager
echo ========================================
echo.

REM Create necessary directories
if not exist "logs" mkdir logs
if not exist "pids" mkdir pids

REM [1] Stop old processes
echo [1/4] Cleaning old processes...
call :kill_port 8080
call :kill_port 3001
call :kill_port 5173
timeout /t 3 /nobreak >nul
echo     [OK] Old processes cleaned (ports 8080, 3001, 5173)

REM [2] Start service manager (run in background, completely hidden)
echo [2/4] Starting service manager...
echo.
echo Service Manager Features:
echo   - Auto monitoring service status
echo   - Auto restart on service crash
echo   - Health check mechanism
echo   - Complete logging
echo.

if exist "service-manager.js" (
    wscript //nologo "%~dp0run_hidden.js" "%~dp0" "node service-manager.js" "%~dp0logs\service-manager.log"
    echo     [OK] Service manager started in background
) else (
    echo     [ERROR] service-manager.js not found
    pause
    exit /b 1
)

REM [3] Wait for services to start
echo [3/4] Waiting for services to start...
timeout /t 5 /nobreak >nul

REM [4] Verify services are ready
echo [4/4] Verifying service status...
echo.

REM Check hotspot server port
netstat -ano | findstr ":8080" | findstr "LISTENING" >nul
if %errorLevel% neq 0 (
    echo     [WARNING] Hotspot server not ready, but continuing...
) else (
    echo     [OK] Hotspot server ready (port 8080)
)

REM Check backend port
netstat -ano | findstr ":3001" | findstr "LISTENING" >nul
if %errorLevel% neq 0 (
    echo     [WARNING] Backend not ready, but continuing...
) else (
    echo     [OK] Backend service ready (port 3001)
)

REM Check frontend port
netstat -ano | findstr ":5173" | findstr "LISTENING" >nul
if %errorLevel% neq 0 (
    echo     [WARNING] Frontend not ready, but continuing...
) else (
    echo     [OK] Frontend service ready (port 5173)
)

echo.
echo ========================================
echo   Service Startup Complete!
echo ========================================
echo.
echo  Hotspot Server: http://localhost:8080
echo  Backend:        http://localhost:3001
echo  Frontend:       http://localhost:5173
echo.
echo  Service manager is running in background
echo  - Auto monitoring service status
echo  - Auto restart on service crash
echo.
echo  View logs: logs/ directory
echo  Stop services: run stop.bat
echo.

REM Open browser
echo Opening browser...
start "" "http://localhost:5173"

echo.
echo  Window will close in 3 seconds...
timeout /t 3 /nobreak >nul
exit

REM ========================================
REM Function: kill_port - Stop process by port
REM ========================================
:kill_port
for /f "tokens=5" %%a in ('netstat -ano ^| findstr ":%1" ^| findstr "LISTENING" 2^>nul') do (
    taskkill /PID %%a /F >nul 2>&1
)
exit /b 0
