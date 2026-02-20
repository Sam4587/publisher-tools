@echo off
chcp 65001 >nul
title Publisher Tools - Stop Service
color 0C

echo.
echo ========================================
echo    Publisher Tools - Stopping Services...
echo ========================================
echo.

REM [1] Stop Go backend (port 8080)
echo [1/3] Stopping Go backend (port 8080)...
call :kill_port 8080
echo     [OK] Go backend stopped

REM [2] Stop Node backend (port 3001)
echo [2/3] Stopping Node backend (port 3001)...
call :kill_port 3001
echo     [OK] Node backend stopped

REM [3] Stop frontend (port 5173)
echo [3/3] Stopping frontend (port 5173)...
call :kill_port 5173
echo     [OK] Frontend stopped

echo.
echo ========================================
echo    All Services Stopped!
echo ========================================
echo.
echo  Window will close in 2 seconds...
timeout /t 2 /nobreak >nul
exit

REM ========================================
REM Function: kill_port - Stop process by port
REM ========================================
:kill_port
for /f "tokens=5" %%a in ('netstat -ano ^| findstr ":%1" ^| findstr "LISTENING" 2^>nul') do (
    taskkill /PID %%a /F >nul 2>&1
)
exit /b 0
