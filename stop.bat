@echo off
chcp 65001 >nul
echo ========================================
echo Publisher Tools - 服务停止脚本
echo ========================================
echo.

echo [信息] 正在停止所有服务...
echo.

REM 停止 Go 后端
echo [1/3] 停止 Go 后端...
for /f "tokens=2" %%a in ('tasklist /fi "windowtitle eq Publisher-Go-Backend*" /fo list ^| findstr "PID:"') do (
    taskkill /PID %%a /F >nul 2>&1
    echo [成功] Go 后端已停止
)
if not errorlevel 1 if not errorlevel 0 echo [信息] Go 后端未运行

REM 停止 Node 后端
echo [2/3] 停止 Node 后端...
for /f "tokens=2" %%a in ('tasklist /fi "windowtitle eq Publisher-Node-Backend*" /fo list ^| findstr "PID:"') do (
    taskkill /PID %%a /F >nul 2>&1
    echo [成功] Node 后端已停止
)
if not errorlevel 1 if not errorlevel 0 echo [信息] Node 后端未运行

REM 停止前端
echo [3/3] 停止前端...
for /f "tokens=2" %%a in ('tasklist /fi "windowtitle eq Publisher-Frontend*" /fo list ^| findstr "PID:"') do (
    taskkill /PID %%a /F >nul 2>&1
    echo [成功] 前端已停止
)
if not errorlevel 1 if not errorlevel 0 echo [信息] 前端未运行

REM 根据端口查找并停止进程
echo.
echo [检查] 根据端口查找进程...

REM 检查端口 8080
for /f "tokens=5" %%a in ('netstat -ano ^| findstr ":8080" ^| findstr "LISTENING"') do (
    echo [信息] 发现端口 8080 被进程 %%a 占用
    taskkill /PID %%a /F >nul 2>&1
    echo [成功] 已停止进程 %%a
)

REM 检查端口 3001
for /f "tokens=5" %%a in ('netstat -ano ^| findstr ":3001" ^| findstr "LISTENING"') do (
    echo [信息] 发现端口 3001 被进程 %%a 占用
    taskkill /PID %%a /F >nul 2>&1
    echo [成功] 已停止进程 %%a
)

REM 检查端口 5173
for /f "tokens=5" %%a in ('netstat -ano ^| findstr ":5173" ^| findstr "LISTENING"') do (
    echo [信息] 发现端口 5173 被进程 %%a 占用
    taskkill /PID %%a /F >nul 2>&1
    echo [成功] 已停止进程 %%a
)

REM 删除 PID 文件
if exist "pids\services.pid" (
    del /f /q "pids\services.pid"
    echo [信息] PID 文件已删除
)

echo.
echo ========================================
echo [完成] 所有服务已停止
========================================
echo.
pause
