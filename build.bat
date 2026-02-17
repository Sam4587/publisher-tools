@echo off
chcp 65001 >nul
echo ========================================
echo Publisher Tools - 编译脚本
echo ========================================
echo.

REM 检查 Go 环境
echo [检查] Go 环境...
where go >nul 2>&1
if %errorLevel% neq 0 (
    echo [错误] Go 未安装
    echo 请先运行 install.bat 安装环境
    pause
    exit /b 1
)

REM 创建输出目录
if not exist "bin" mkdir bin

REM 编译 Go 后端
echo.
echo [编译] 正在编译 Go 后端...
cd publisher-core

echo [信息] 清理旧文件...
if exist "..\bin\publisher-server.exe" del /f /q "..\bin\publisher-server.exe"

echo [信息] 下载依赖...
go mod download

echo [信息] 开始编译...
go build -ldflags="-s -w" -o ..\bin\publisher-server.exe .\cmd\server

if %errorLevel% == 0 (
    echo.
    echo ========================================
    echo [成功] 编译完成！
    echo ========================================
    echo.
    echo 输出文件: bin\publisher-server.exe
    for %%F in ("bin\publisher-server.exe") do echo 文件大小: %%~zF 字节
    echo.
) else (
    echo.
    echo ========================================
    echo [错误] 编译失败！
    echo ========================================
    echo.
    echo 可能的原因:
    echo 1. Go 版本过低（需要 1.21 或更高）
    echo 2. 依赖下载失败
    echo 3. 代码有错误
    echo.
    echo 请检查错误信息并修复问题后重试
    echo.
)

cd ..
pause
