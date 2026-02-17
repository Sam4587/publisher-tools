@echo off
chcp 65001 >nul
echo ========================================
echo Publisher Tools - 查看日志
echo ========================================
echo.

echo 可用的日志文件:
echo.
if exist "logs\go-backend.log" (
    echo [1] Go 后端日志
)
if exist "logs\node-backend.log" (
    echo [2] Node 后端日志
)
if exist "logs\frontend.log" (
    echo [3] 前端日志
)
echo.

set /p choice="请选择要查看的日志 (1/2/3) 或按 Enter 查看全部: "

if "%choice%"=="" (
    echo.
    echo ========================================
    echo Go 后端日志:
    echo ========================================
    if exist "logs\go-backend.log" (
        type "logs\go-backend.log"
    ) else (
        echo [日志文件不存在]
    )
    echo.
    echo ========================================
    echo Node 后端日志:
    echo ========================================
    if exist "logs\node-backend.log" (
        type "logs\node-backend.log"
    ) else (
        echo [日志文件不存在]
    )
    echo.
    echo ========================================
    echo 前端日志:
    echo ========================================
    if exist "logs\frontend.log" (
        type "logs\frontend.log"
    ) else (
        echo [日志文件不存在]
    )
) else if "%choice%"=="1" (
    echo.
    echo ========================================
    echo Go 后端日志:
    echo ========================================
    if exist "logs\go-backend.log" (
        type "logs\go-backend.log"
    ) else (
        echo [日志文件不存在]
    )
) else if "%choice%"=="2" (
    echo.
    echo ========================================
    echo Node 后端日志:
    echo ========================================
    if exist "logs\node-backend.log" (
        type "logs\node-backend.log"
    ) else (
        echo [日志文件不存在]
    )
) else if "%choice%"=="3" (
    echo.
    echo ========================================
    echo 前端日志:
    echo ========================================
    if exist "logs\frontend.log" (
        type "logs\frontend.log"
    ) else (
        echo [日志文件不存在]
    )
) else (
    echo [错误] 无效的选择
)

echo.
pause
