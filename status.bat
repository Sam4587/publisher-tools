@echo off
chcp 65001 >nul
echo ========================================
echo Publisher Tools - 服务状态检查
echo ========================================
echo.

echo [检查] 服务运行状态...
echo.

REM 检查端口状态
echo 端口状态:
echo ----------------------------------------

REM 检查 8080 端口
netstat -ano | findstr ":8080" | findstr "LISTENING" >nul
if %errorLevel% == 0 (
    echo [运行中] Go 后端（端口 8080）
    for /f "tokens=5" %%a in ('netstat -ano ^| findstr ":8080" ^| findstr "LISTENING"') do echo           PID: %%a
) else (
    echo [已停止] Go 后端（端口 8080）
)

REM 检查 3001 端口
netstat -ano | findstr ":3001" | findstr "LISTENING" >nul
if %errorLevel% == 0 (
    echo [运行中] Node 后端（端口 3001）
    for /f "tokens=5" %%a in ('netstat -ano ^| findstr ":3001" ^| findstr "LISTENING"') do echo           PID: %%a
) else (
    echo [已停止] Node 后端（端口 3001）
)

REM 检查 5173 端口
netstat -ano | findstr ":5173" | findstr "LISTENING" >nul
if %errorLevel% == 0 (
    echo [运行中] 前端（端口 5173）
    for /f "tokens=5" %%a in ('netstat -ano ^| findstr ":5173" ^| findstr "LISTENING"') do echo           PID: %%a
) else (
    echo [已停止] 前端（端口 5173）
)

echo.
echo 窗口进程:
echo ----------------------------------------

REM 检查窗口进程
tasklist /fi "windowtitle eq Publisher-Go-Backend*" /fo list 2>nul | findstr "PID:" >nul
if %errorLevel% == 0 (
    echo [运行中] Go 后端窗口
    for /f "tokens=2" %%a in ('tasklist /fi "windowtitle eq Publisher-Go-Backend*" /fo list ^| findstr "PID:"') do echo           PID: %%a
) else (
    echo [未运行] Go 后端窗口
)

tasklist /fi "windowtitle eq Publisher-Node-Backend*" /fo list 2>nul | findstr "PID:" >nul
if %errorLevel% == 0 (
    echo [运行中] Node 后端窗口
    for /f "tokens=2" %%a in ('tasklist /fi "windowtitle eq Publisher-Node-Backend*" /fo list ^| findstr "PID:"') do echo           PID: %%a
) else (
    echo [未运行] Node 后端窗口
)

tasklist /fi "windowtitle eq Publisher-Frontend*" /fo list 2>nul | findstr "PID:" >nul
if %errorLevel% == 0 (
    echo [运行中] 前端窗口
    for /f "tokens=2" %%a in ('tasklist /fi "windowtitle eq Publisher-Frontend*" /fo list ^| findstr "PID:"') do echo           PID: %%a
) else (
    echo [未运行] 前端窗口
)

echo.
echo 日志文件:
echo ----------------------------------------
if exist "logs\go-backend.log" (
    echo [存在] Go 后端日志: logs\go-backend.log
    for %%F in ("logs\go-backend.log") do echo        大小: %%~zF 字节
) else (
    echo [不存在] Go 后端日志
)

if exist "logs\node-backend.log" (
    echo [存在] Node 后端日志: logs\node-backend.log
    for %%F in ("logs\node-backend.log") do echo        大小: %%~zF 字节
) else (
    echo [不存在] Node 后端日志
)

if exist "logs\frontend.log" (
    echo [存在] 前端日志: logs\frontend.log
    for %%F in ("logs\frontend.log") do echo        大小: %%~zF 字节
) else (
    echo [不存在] 前端日志
)

echo.
echo ========================================
echo 访问地址:
echo   前端: http://localhost:5173
echo   Go API: http://localhost:8080/health
echo   Node API: http://localhost:3001/api/health
echo ========================================
echo.
pause
