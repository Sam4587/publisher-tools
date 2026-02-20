@echo off
chcp 65001 >nul
title Publisher Tools - Stop Service
color 0C

echo.
echo ========================================
echo    Publisher Tools - Stopping Services...
echo ========================================
echo.

REM [1] Stop Go backend
echo [1/3] Stopping Go backend (port 8080)...
call :kill_port 8080
if %errorLevel% equ 0 (
    echo     [OK] Port 8080 cleaned
) else (
    echo     [INFO] Port 8080 not in use
)

REM [2] Stop Node backend
echo [2/3] Stopping Node backend (port 3001)...
call :kill_port 3001
if %errorLevel% equ 0 (
    echo     [OK] Port 3001 cleaned
) else (
    echo     [INFO] Port 3001 not in use
)

REM [3] Stop frontend
echo [3/3] Stopping frontend (port 5173)...
call :kill_port 5173
if %errorLevel% equ 0 (
    echo     [OK] Port 5173 cleaned
) else (
    echo     [INFO] Port 5173 not in use
)

REM Clean PID files
if exist "pids\services.pid" (
    del /f /q "pids\services.pid" >nul 2>&1
    echo.
    echo [INFO] PID files cleaned
)

echo.
echo ========================================
echo    All Services Stopped
echo ========================================
echo.
echo  Window will close in 2 seconds...
timeout /t 2 /nobreak >nul
exit

REM ========================================
REM Function: kill_port - Stop process by port
REM ========================================
:kill_port
setlocal enabledelayedexpansion
set "found=0"
for /f "tokens=5" %%a in ('netstat -ano ^| findstr ":%1" ^| findstr "LISTENING" 2^>nul') do (
    taskkill /PID %%a /F >nul 2>&1
    set "found=1"
)
endlocal & exit /b %found%
