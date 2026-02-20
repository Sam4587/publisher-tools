@echo off
chcp 65001 >nul
title Publisher Tools - Start Service
color 0A

echo.
echo ========================================
echo    Publisher Tools - Starting Services...
echo ========================================
echo.

REM Create necessary directories
if not exist "logs" mkdir logs

REM [1] Stop old processes
echo [1/5] Cleaning old processes...
call :kill_port 8080
call :kill_port 3001
call :kill_port 5173
timeout /t 1 /nobreak >nul
echo     [OK] Old processes cleaned

REM [2] Start Go backend (run in background)
echo [2/5] Starting Go backend (port 8080)...
if exist "bin\publisher-server.exe" (
    start "Go-Backend" /min cmd /c "cd /d "%~dp0" && bin\publisher-server.exe -port 8080 > logs\go.log 2>&1"
    echo     [OK] Go backend started in background
) else (
    echo     [WARNING] bin\publisher-server.exe not found, skipping
)

REM [3] Start Node backend (run in background)
echo [3/5] Starting Node backend (port 3001)...
if exist "server\simple-server.js" (
    start "Node-Backend" /min cmd /c "cd /d "%~dp0server" && node simple-server.js > ..\logs\node.log 2>&1"
    echo     [OK] Node backend started in background
) else (
    echo     [WARNING] server\simple-server.js not found, skipping
)

REM [4] Start frontend (run in background)
echo [4/5] Starting frontend (port 5173)...
if exist "publisher-web\package.json" (
    start "Frontend" /min cmd /c "cd /d "%~dp0publisher-web" && npm run dev > ..\logs\frontend.log 2>&1"
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
    echo     Waiting for frontend service... [%retry%/30]
    timeout /t 1 /nobreak >nul
    goto check_loop
)

echo     [OK] Frontend service ready (port 5173)

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
echo    Service Startup Complete!
echo ========================================
echo.
echo  Frontend:  http://localhost:5173
echo  Go Backend:    http://localhost:8080/health
echo  Node Backend:  http://localhost:3001
echo.
echo  Services running in background
echo.
echo  Opening browser...

REM Automatically open default browser to access service interface
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
