@echo off
chcp 65001 >nul
title Publisher Tools - 服务管理
color 0A

:main
cls
echo.
echo ========================================
echo    Publisher Tools - 服务管理中心
echo ========================================
echo.
echo  1. 启动所有服务 (含夸克浏览器)
echo  2. 停止所有服务
echo  3. 查看服务状态
echo  4. 一键安装环境并启动
echo  5. 编译项目
echo  0. 退出
echo.
echo ========================================
set /p choice=请选择 (0-5):

if "%choice%"=="1" goto start
if "%choice%"=="2" goto stop
if "%choice%"=="3" goto status
if "%choice%"=="4" goto install
if "%choice%"=="5" goto build
if "%choice%"=="0" exit
goto main

:start
echo.
echo [启动] 正在启动服务...
echo.

echo [1] 停止旧进程...
call :kill_port 8080
call :kill_port 3001
call :kill_port 5173
timeout /t 1 /nobreak >nul

echo [2] 启动 Go 后端...
if exist "bin\publisher-server.exe" (
    if not exist "logs" mkdir logs
    start "Go-Backend" /min cmd /c "bin\publisher-server.exe -port 8080 > logs\go.log 2>&1"
    echo     [OK] Go 后端 (8080)
) else (
    echo     [错误] bin\publisher-server.exe 不存在
)

echo [3] 启动 Node 后端...
if exist "server\simple-server.js" (
    start "Node-Backend" /min cmd /c "cd /d "%~dp0server" && node simple-server.js > ..\logs\node.log 2>&1"
    echo     [OK] Node 后端 (3001)
) else (
    echo     [错误] server\simple-server.js 不存在
)

echo [4] 启动前端...
if exist "publisher-web\node_modules" (
    start "Frontend" /min cmd /c "cd /d "%~dp0publisher-web" && npm run dev > ..\logs\frontend.log 2>&1"
    echo     [OK] 前端 (5173)
) else (
    echo     [错误] publisher-web\node_modules 不存在
)

echo [5] 打开夸克浏览器...
timeout /t 2 /nobreak >nul
start "" "quark" "http://localhost:5173"

echo.
echo ========================================
echo 完成！服务已启动 (可关闭此窗口)
echo ========================================
echo.
echo 前端: http://localhost:5173
echo Go后端: http://localhost:8080/health
echo.
echo 注意: 服务在后台运行，可直接关闭此窗口
pause >nul
exit

:stop
echo.
echo [停止] 正在停止服务...
call :kill_port 8080
call :kill_port 3001
call :kill_port 5173
echo.
echo 完成！服务已停止
pause >nul
goto main

:status
echo.
echo [状态] 服务运行情况
echo.
call :check_port 8080 "Go 后端"
call :check_port 3001 "Node 后端"
call :check_port 5173 "前端"
echo.
pause >nul
goto main

:install
echo.
echo [安装] 正在安装环境...
echo.

if not exist "bin" mkdir bin
if not exist "logs" mkdir logs
if not exist "cookies" mkdir cookies
if not exist "uploads" mkdir uploads
echo [目录] 创建完成

echo.
echo [npm] 安装前端依赖...
cd publisher-web
if not exist "node_modules" call npm install --registry=https://registry.npmmirror.com
cd ..

echo [npm] 安装后端依赖...
cd server
if not exist "node_modules" call npm install --registry=https://registry.npmmirror.com
cd ..

echo.
echo [Go] 编译后端...
cd publisher-core
go mod download
go build -ldflags="-s -w" -o ..\bin\publisher-server.exe .\cmd\server
cd ..

echo.
echo [完成] 环境就绪
pause >nul
goto main

:build
echo.
echo [编译] 正在编译...
cd publisher-core
go mod tidy
go build -ldflags="-s -w" -o ..\bin\publisher-server.exe .\cmd\server
cd ..
if exist "bin\publisher-server.exe" (
    echo [完成] bin\publisher-server.exe
) else (
    echo [错误] 编译失败
)
pause >nul
goto main

:kill_port
for /f "tokens=5" %%a in ('netstat -ano ^| findstr ":%1" ^| findstr "LISTENING"') do taskkill /PID %%a /F >nul 2>&1
exit /b 0

:check_port
netstat -ano | findstr ":%1" ^| findstr "LISTENING" >nul
if %errorLevel%==0 (
    echo [%1] %2 - 运行中
) else (
    echo [%1] %2 - 已停止
)
exit /b 0
