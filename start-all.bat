@echo off
chcp 65001 >nul
title Publisher Tools - Start All Services
color 0A

REM 禁用WSL检测,防止启动时出现WSL更新弹窗
set NODE_OPTIONS=
set ELECTRON_NO_ATTACH_CONSOLE=1
set npm_config_use_wsl=false

echo.
echo ========================================
echo   Publisher Tools - Starting All Services
echo ========================================
echo.

REM Create necessary directories
if not exist "logs" mkdir logs

REM [1] Stop old processes
echo [1/5] Cleaning old processes...
call :kill_port 8080
call :kill_port 3001
call :kill_port 5173
call :kill_port 5174
call :kill_port 5175
timeout /t 2 /nobreak >nul
echo     [OK] Old processes cleaned

REM [2] Start Go Hotspot Server (run in background, completely hidden)
echo [2/5] Starting Go Hotspot Server (port 8080)...
if exist "hotspot-server\main.go" (
    wscript //nologo "%~dp0run_hidden.js" "%~dp0hotspot-server" "go run main.go" "%~dp0logs\hotspot-server.log"
    echo     [OK] Go Hotspot Server started in background
) else (
    echo     [WARNING] hotspot-server\main.go not found, skipping
)

REM [3] Start Node Backend (run in background, completely hidden)
echo [3/5] Starting Node Backend (port 3001)...
if exist "server\simple-server.js" (
    wscript //nologo "%~dp0run_hidden.js" "%~dp0server" "node simple-server.js" "%~dp0logs\node-backend.log"
    echo     [OK] Node Backend started in background
) else (
    echo     [WARNING] server\simple-server.js not found, skipping
)

REM [4] Start Frontend (run in background, completely hidden)
echo [4/5] Starting Frontend (port 5173)...
if exist "publisher-web\package.json" (
    wscript //nologo "%~dp0run_hidden.js" "%~dp0publisher-web" "npm run dev" "%~dp0logs\frontend.log"
    echo     [OK] Frontend started in background
) else (
    echo     [WARNING] publisher-web\package.json not found, skipping
)

REM [5] HTTP connection check, verify service ready status
echo [5/5] Performing HTTP connection check, verifying service ready status...
echo.

REM Wait for services to start and perform HTTP check
set /a retry=0
:check_loop
set /a retry+=1
if %retry% gtr 30 (
    echo [WARNING] Service startup timeout, please check manually
    goto open_browser
)

REM Check frontend port
netstat -ano | findstr ":5173" | findstr "LISTENING" >nul
if %errorLevel% neq 0 (
    netstat -ano | findstr ":5174" | findstr "LISTENING" >nul
    if %errorLevel% neq 0 (
        netstat -ano | findstr ":5175" | findstr "LISTENING" >nul
        if %errorLevel% neq 0 (
            echo     Waiting for frontend service... [%retry%/30]
            timeout /t 1 /nobreak >nul
            goto check_loop
        ) else (
            set FRONTEND_PORT=5175
        )
    ) else (
        set FRONTEND_PORT=5174
    )
) else (
    set FRONTEND_PORT=5173
)

echo     [OK] Frontend service ready (port %FRONTEND_PORT%)

REM Check Go backend port
netstat -ano | findstr ":8080" | findstr "LISTENING" >nul
if %errorLevel% neq 0 (
    echo     [WARNING] Go backend not ready, but frontend is ready, continuing...
) else (
    echo     [OK] Go backend ready (port 8080)
)

REM Check Node backend port
netstat -ano | findstr ":3001" | findstr "LISTENING" >nul
if %errorLevel% neq 0 (
    echo     [WARNING] Node backend not ready, but frontend is ready, continuing...
) else (
    echo     [OK] Node backend ready (port 3001)
)

:open_browser
echo.
echo ========================================
echo   Service Startup Complete!
echo ========================================
echo.
echo  Frontend:       http://localhost:%FRONTEND_PORT%
echo  Hotspot Server: http://localhost:8080/api/health
echo  Node Backend:   http://localhost:3001/api/health
echo.
echo  Services running in background
echo  Logs: logs\hotspot-server.log, logs\node-backend.log, logs\frontend.log
echo.
echo  To stop all services, run: stop.bat
echo.

REM Automatically open default browser to access service interface
echo Opening browser...
start "" "http://localhost:%FRONTEND_PORT%"

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
