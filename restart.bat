@echo off
chcp 65001 >nul
echo ========================================
echo Publisher Tools - 服务重启脚本
echo ========================================
echo.

echo [步骤 1/2] 停止所有服务...
if exist "stop.bat" (
    call stop.bat
) else (
    echo [错误] 找不到 stop.bat
    pause
    exit /b 1
)

echo.
echo [步骤 2/2] 等待服务完全停止...
timeout /t 3 /nobreak >nul

echo.
echo [信息] 正在重新启动服务...
if exist "start.bat" (
    call start.bat
) else (
    echo [错误] 找不到 start.bat
    pause
    exit /b 1
)
