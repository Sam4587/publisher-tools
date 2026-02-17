@echo off
chcp 65001 >nul
title Publisher Tools - 一键启动
color 0A

echo.
echo ╔══════════════════════════════════════════════════╗
echo ║        Publisher Tools - 一键启动全部服务        ║
echo ╚══════════════════════════════════════════════════╝
echo.

REM 创建必要目录
if not exist "logs" mkdir logs
if not exist "pids" mkdir pids

REM 检查并停止已运行的服务
echo [检查] 端口占用情况...
echo.

netstat -ano | findstr ":8080" | findstr "LISTENING" >nul
if %errorLevel% == 0 (
    echo [警告] 端口 8080 已被占用，正在停止...
    for /f "tokens=5" %%a in ('netstat -ano ^| findstr ":8080" ^| findstr "LISTENING"') do taskkill /PID %%a /F >nul 2>&1
)

netstat -ano | findstr ":3001" | findstr "LISTENING" >nul
if %errorLevel% == 0 (
    echo [警告] 端口 3001 已被占用，正在停止...
    for /f "tokens=5" %%a in ('netstat -ano ^| findstr ":3001" ^| findstr "LISTENING"') do taskkill /PID %%a /F >nul 2>&1
)

netstat -ano | findstr ":5173" | findstr "LISTENING" >nul
if %errorLevel% == 0 (
    echo [警告] 端口 5173 已被占用，正在停止...
    for /f "tokens=5" %%a in ('netstat -ano ^| findstr ":5173" ^| findstr "LISTENING"') do taskkill /PID %%a /F >nul 2>&1
)

timeout /t 1 /nobreak >nul

echo.
echo ========================================
echo [启动] 正在启动服务...
echo ========================================
echo.

REM 1. 启动 Go 后端（端口 8080）
echo [1/3] 启动 Go 后端（端口 8080）...
if exist "bin\publisher-server.exe" (
    start "Publisher-Go-Backend" /min cmd /c "bin\publisher-server.exe -port 8080 > logs\go-backend.log 2>&1"
    timeout /t 2 /nobreak >nul
    echo       [成功] Go 后端已启动
) else (
    echo       [错误] bin\publisher-server.exe 不存在
    echo       [提示] 请先运行: cd publisher-core && go build -o ..\bin\publisher-server.exe cmd\server\main.go
)

REM 2. 启动 Node 后端（端口 3001）
echo [2/3] 启动 Node 后端（端口 3001）...
if exist "server\simple-server.js" (
    cd server
    start "Publisher-Node-Backend" /min cmd /c "node simple-server.js > ..\logs\node-backend.log 2>&1"
    cd ..
    timeout /t 2 /nobreak >nul
    echo       [成功] Node 后端已启动
) else (
    echo       [错误] server\simple-server.js 不存在
)

REM 3. 启动前端（端口 5173）
echo [3/3] 启动前端（端口 5173）...
if exist "publisher-web\node_modules" (
    cd publisher-web
    start "Publisher-Frontend" cmd /c "npm run dev > ..\logs\frontend.log 2>&1"
    cd ..
    timeout /t 3 /nobreak >nul
    echo       [成功] 前端已启动
) else (
    echo       [错误] 前端依赖未安装
    echo       [提示] 请先运行: cd publisher-web && npm install
)

echo.
echo ========================================
echo [完成] 服务启动完成！
echo ========================================
echo.
echo 服务地址:
echo   前端界面:    http://localhost:5173
echo   Go 后端:     http://localhost:8080/health
echo   Node 后端:   http://localhost:3001/api/health
echo.
echo 日志文件:
echo   logs\go-backend.log
echo   logs\node-backend.log
echo   logs\frontend.log
echo.
echo 管理命令:
echo   停止服务: 运行 stop.bat 或 双击 stop.bat
echo   查看状态: 运行 status.bat
echo.

REM 创建 PID 标记文件
echo %date% %time% > pids\services.pid

echo [提示] 5 秒后自动打开浏览器...
timeout /t 5 /nobreak >nul
start http://localhost:5173

echo.
echo [提示] 请保持此窗口打开，服务正在后台运行
echo [提示] 需要停止服务时，请运行 stop.bat
echo.
pause
