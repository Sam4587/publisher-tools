@echo off
chcp 65001 >nul
title Publisher Tools - 快速启动
color 0A

echo.
echo ╔══════════════════════════════════════╗
echo ║                                      ║
echo ║    Publisher Tools - 快速启动        ║
echo ║                                      ║
echo ╚══════════════════════════════════════╝
echo.

REM 检查是否首次运行
if not exist "bin\publisher-server.exe" (
    echo [提示] 首次运行，正在安装环境...
    echo.
    call install.bat
    if %errorLevel% neq 0 (
        echo.
        echo [错误] 环境安装失败
        pause
        exit /b 1
    )
)

echo [启动] 正在启动服务...
if exist "start.bat" (
    call start.bat
) else (
    echo [错误] 找不到 start.bat
    pause
    exit /b 1
)
