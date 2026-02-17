@echo off
chcp 65001 >nul
title Publisher Tools - 服务运行中
echo ========================================
echo Publisher Tools - 服务启动脚本
echo ========================================
echo.

REM 检查是否已经运行
if exist "pids\services.pid" (
    echo [警告] 服务可能已经在运行
    echo.
    set /p choice="是否要重启服务？(Y/N): "
    if /i "%choice%"=="Y" (
        echo [信息] 正在停止旧服务...
        call stop.bat
        timeout /t 2 /nobreak >nul
    ) else (
        echo [取消] 操作已取消
        pause
        exit /b 0
    )
)

REM 检查必要文件
echo [检查] 验证文件...
if not exist "bin\publisher-server.exe" (
    echo [错误] Go 后端未编译
    echo [信息] 正在编译...
    call build.bat
    if %errorLevel% neq 0 (
        echo [错误] 编译失败，无法启动服务
        pause
        exit /b 1
    )
)

if not exist "publisher-web\node_modules" (
    echo [错误] 前端依赖未安装
    echo [信息] 请先运行 install.bat
    pause
    exit /b 1
)

if not exist "server\node_modules" (
    echo [错误] Node 后端依赖未安装
    echo [信息] 请先运行 install.bat
    pause
    exit /b 1
)

REM 创建日志目录
if not exist "logs" mkdir logs

REM 检查端口是否被占用
echo.
echo [检查] 端口占用情况...
netstat -ano | findstr ":8080" >nul
if %errorLevel% == 0 (
    echo [警告] 端口 8080 已被占用
    for /f "tokens=5" %%a in ('netstat -ano ^| findstr ":8080"') do set PID=%%a
    echo [信息] 占用进程 PID: %PID%
    set /p kill="是否结束该进程？(Y/N): "
    if /i "!kill!"=="Y" (
        taskkill /PID %PID% /F >nul 2>&1
        echo [成功] 进程已结束
    )
)

netstat -ano | findstr ":3001" >nul
if %errorLevel% == 0 (
    echo [警告] 端口 3001 已被占用
    for /f "tokens=5" %%a in ('netstat -ano ^| findstr ":3001"') do set PID=%%a
    echo [信息] 占用进程 PID: %PID%
    set /p kill="是否结束该进程？(Y/N): "
    if /i "!kill!"=="Y" (
        taskkill /PID %PID% /F >nul 2>&1
        echo [成功] 进程已结束
    )
)

netstat -ano | findstr ":5173" >nul
if %errorLevel% == 0 (
    echo [警告] 端口 5173 已被占用
    for /f "tokens=5" %%a in ('netstat -ano ^| findstr ":5173"') do set PID=%%a
    echo [信息] 占用进程 PID: %PID%
    set /p kill="是否结束该进程？(Y/N): "
    if /i "!kill!"=="Y" (
        taskkill /PID %PID% /F >nul 2>&1
        echo [成功] 进程已结束
    )
)

echo.
echo ========================================
echo [启动] 正在启动服务...
echo ========================================
echo.

REM 启动 Go 后端（端口 8080）
echo [1/3] 启动 Go 后端服务（端口 8080）...
start "Publisher-Go-Backend" /min cmd /c "bin\publisher-server.exe -port 8080 > logs\go-backend.log 2>&1"
timeout /t 2 /nobreak >nul
echo [成功] Go 后端已启动

REM 启动 Node 后端（端口 3001）
echo [2/3] 启动 Node 后端服务（端口 3001）...
cd server
start "Publisher-Node-Backend" /min cmd /c "node _deprecated\index.js > ..\logs\node-backend.log 2>&1"
cd ..
timeout /t 2 /nobreak >nul
echo [成功] Node 后端已启动

REM 启动前端（端口 5173）
echo [3/3] 启动前端服务（端口 5173）...
cd publisher-web
start "Publisher-Frontend" /min cmd /c "npm run dev > ..\logs\frontend.log 2>&1"
cd ..
timeout /t 3 /nobreak >nul
echo [成功] 前端已启动

echo.
echo ========================================
echo [完成] 所有服务启动成功！
echo ========================================
echo.
echo 服务地址:
echo   前端界面:    http://localhost:5173
echo   Go 后端:     http://localhost:8080
echo   Node 后端:   http://localhost:3001
echo.
echo 日志文件:
if exist "logs\go-backend.log" echo   Go 后端:     logs\go-backend.log
if exist "logs\node-backend.log" echo   Node 后端:   logs\node-backend.log
if exist "logs\frontend.log" echo   前端:        logs\frontend.log
echo.
echo 管理命令:
if exist "stop.bat" echo   停止服务:    stop.bat
if exist "restart.bat" echo   重启服务:    restart.bat
echo.

REM 创建 PID 标记文件
echo %date% %time% > pids\services.pid

REM 自动打开浏览器
echo [提示] 5 秒后自动打开浏览器...
timeout /t 5 /nobreak >nul
start http://localhost:5173

echo.
echo [提示] 请保持此窗口打开，服务正在后台运行
echo [提示] 关闭此窗口不会停止服务
if exist "stop.bat" echo [提示] 需要停止服务时，请运行 stop.bat
echo.
pause
