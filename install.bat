@echo off
chcp 65001 >nul
echo ========================================
echo Publisher Tools - 环境安装脚本
echo ========================================
echo.

REM 检查管理员权限
net session >nul 2>&1
if %errorLevel% == 0 (
    echo [信息] 已获取管理员权限
echo.) else (
    echo [警告] 未获取管理员权限，部分功能可能受限
echo.)
)

REM 检查 Go 环境
echo [步骤 1/6] 检查 Go 环境...
where go >nul 2>&1
if %errorLevel% == 0 (
    for /f "tokens=3" %%i in ('go version') do set GO_VERSION=%%i
    echo [成功] Go 已安装: %GO_VERSION%
echo.) else (
    echo [错误] Go 未安装或未添加到 PATH
    echo.
    echo 请按照以下步骤安装 Go:
    echo 1. 访问 https://go.dev/dl/
    echo 2. 下载 Windows 安装包（如 go1.21.6.windows-amd64.msi）
    echo 3. 运行安装程序
    echo 4. 重启命令行窗口后再次运行此脚本
    echo.
    pause
    exit /b 1
)

REM 检查 Node.js 环境
echo [步骤 2/6] 检查 Node.js 环境...
where node >nul 2>&1
if %errorLevel% == 0 (
    for /f "tokens=1" %%i in ('node --version') do set NODE_VERSION=%%i
    echo [成功] Node.js 已安装: %NODE_VERSION%
echo.) else (
    echo [错误] Node.js 未安装或未添加到 PATH
    echo.
    echo 请按照以下步骤安装 Node.js:
    echo 1. 访问 https://nodejs.org/
    echo 2. 下载 LTS 版本（推荐 18.x 或更高）
    echo 3. 运行安装程序
    echo 4. 重启命令行窗口后再次运行此脚本
    echo.
    pause
    exit /b 1
)

REM 创建必要目录
echo [步骤 3/6] 创建必要目录...
if not exist "bin" mkdir bin
if not exist "cookies" mkdir cookies
if not exist "uploads" mkdir uploads
if not exist "logs" mkdir logs
if not exist "pids" mkdir pids
if not exist "data" mkdir data
echo [成功] 目录创建完成
echo.

REM 安装前端依赖
echo [步骤 4/6] 安装前端依赖...
cd publisher-web
if not exist "node_modules" (
    echo [信息] 正在安装前端依赖，请稍候...
    call npm install --registry=https://registry.npmmirror.com
    if %errorLevel% == 0 (
        echo [成功] 前端依赖安装完成
    ) else (
        echo [错误] 前端依赖安装失败
        cd ..
        pause
        exit /b 1
    )
) else (
    echo [信息] 前端依赖已存在，跳过安装
)
cd ..
echo.

REM 安装 Node 后端依赖
echo [步骤 5/6] 安装 Node 后端依赖...
cd server
if not exist "node_modules" (
    echo [信息] 正在安装 Node 后端依赖，请稍候...
    call npm install --registry=https://registry.npmmirror.com
    if %errorLevel% == 0 (
        echo [成功] Node 后端依赖安装完成
    ) else (
        echo [错误] Node 后端依赖安装失败
        cd ..
        pause
        exit /b 1
    )
) else (
    echo [信息] Node 后端依赖已存在，跳过安装
)
cd ..
echo.

REM 编译 Go 后端
echo [步骤 6/6] 编译 Go 后端...
if not exist "bin\publisher-server.exe" (
    echo [信息] 正在编译 Go 后端，请稍候...
    cd publisher-core
    go mod download
    go build -o ..\bin\publisher-server.exe .\cmd\server
    if %errorLevel% == 0 (
        echo [成功] Go 后端编译完成
    ) else (
        echo [错误] Go 后端编译失败
        cd ..
        pause
        exit /b 1
    )
    cd ..
) else (
    echo [信息] Go 后端已存在，跳过编译
)
echo.

echo ========================================
echo [完成] 环境安装成功！
echo ========================================
echo.
echo 下一步:
echo 1. 运行 start.bat 启动服务
echo 2. 浏览器访问 http://localhost:5173
echo.
pause
