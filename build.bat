@echo off
setlocal enabledelayedexpansion

REM GLM Usage Monitor Windows 构建脚本
REM 支持 Windows 平台构建

REM 颜色定义
set "RED=[91m"
set "GREEN=[92m"
set "YELLOW=[93m"
set "BLUE=[94m"
set "NC=[0m"

REM 默认参数
set PLATFORM=current
set ARCH=current
set CLEAN_BUILD=false
set DEV_MODE=false
set OUTPUT_DIR=build\bin

REM 显示帮助信息
:show_help
echo GLM Usage Monitor 构建脚本
echo.
echo 用法: %~nx0 [选项]
echo.
echo 选项:
echo   -p, --platform PLATFORM    目标平台 (windows^|all) (默认: current)
echo   -a, --arch ARCH             目标架构 (amd64^|arm64^|all) (默认: current)
echo   -c, --clean                 构建前清理
echo   -d, --dev                   开发模式构建
echo   -o, --output DIR            输出目录 (默认: .\build\bin)
echo   -h, --help                  显示此帮助信息
echo.
echo 示例:
echo   %~nx0                       构建当前平台
echo   %~nx0 -p all -a all         构建所有平台和架构
echo   %~nx0 -c                    清理并构建当前平台
goto :eof

REM 日志函数
:log_info
echo %BLUE%[INFO]%NC% %~1
goto :eof

:log_success
echo %GREEN%[SUCCESS]%NC% %~1
goto :eof

:log_warning
echo %YELLOW%[WARNING]%NC% %~1
goto :eof

:log_error
echo %RED%[ERROR]%NC% %~1
goto :eof

REM 解析命令行参数
:parse_args
if "%~1"=="" goto :main_loop
if "%~1"=="-p" set PLATFORM=%~2& shift & shift & goto :parse_args
if "%~1"=="--platform" set PLATFORM=%~2& shift & shift & goto :parse_args
if "%~1"=="-a" set ARCH=%~2& shift & shift & goto :parse_args
if "%~1"=="--arch" set ARCH=%~2& shift & shift & goto :parse_args
if "%~1"=="-c" set CLEAN_BUILD=true& shift & goto :parse_args
if "%~1"=="--clean" set CLEAN_BUILD=true& shift & goto :parse_args
if "%~1"=="-d" set DEV_MODE=true& shift & goto :parse_args
if "%~1"=="--dev" set DEV_MODE=true& shift & goto :parse_args
if "%~1"=="-o" set OUTPUT_DIR=%~2& shift & shift & goto :parse_args
if "%~1"=="--output" set OUTPUT_DIR=%~2& shift & shift & goto :parse_args
if "%~1"=="-h" goto :show_help
if "%~1"=="--help" goto :show_help
call :log_error "未知参数: %~1"
goto :show_help

REM 检测当前架构
:detect_current_arch
set CURRENT_ARCH=amd64
if /i "%PROCESSOR_ARCHITECTURE%"=="ARM64" set CURRENT_ARCH=arm64
if /i "%PROCESSOR_ARCHITEW6432%"=="ARM64" set CURRENT_ARCH=arm64
goto :eof

REM 执行构建
:build_platform
set platform=%~1
set arch=%~2
set clean_flag=%~3

call :log_info "开始构建平台: %platform%/%arch%"

set build_cmd=wails build

if not "%platform%"=="current" (
    set build_cmd=!build_cmd! -platform %platform%
)

if not "%arch%"=="current" (
    set build_cmd=!build_cmd! -arch %arch%
)

if "%clean_flag%"=="true" (
    set build_cmd=!build_cmd! -clean
)

if "%DEV_MODE%"=="true" (
    set build_cmd=!build_cmd! -debug
)

call :log_info "执行命令: !build_cmd!"

!build_cmd!
if !errorlevel! equ 0 (
    call :log_success "构建完成: %platform%/%arch%"

    REM 检查构建产物
    if exist "build\bin\glm-usage-monitor.exe" (
        for %%F in ("build\bin\glm-usage-monitor.exe") do (
            call :log_info "构建产物: %%F (%%~zF bytes)"
        )
    )
) else (
    call :log_error "构建失败: %platform%/%arch%"
    exit /b 1
)
goto :eof

:main_loop
call :log_info "开始 GLM Usage Monitor 构建流程"

REM 检查是否在正确的目录
if not exist "wails.json" (
    call :log_error "请在项目根目录（包含 wails.json 的目录）中运行此脚本"
    exit /b 1
)

REM 检查依赖
where wails >nul 2>nul
if !errorlevel! neq 0 (
    call :log_error "未找到 Wails CLI，请先安装: https://wails.io/docs/gettingstarted/installation"
    exit /b 1
)

where go >nul 2>nul
if !errorlevel! neq 0 (
    call :log_error "未找到 Go，请先安装: https://go.dev/doc/install"
    exit /b 1
)

REM 检测当前架构
call :detect_current_arch
call :log_info "当前平台: windows/%CURRENT_ARCH%"

REM 解析平台和架构
set platforms=
set archs=

if "%PLATFORM%"=="current" (
    set platforms=windows
) else if "%PLATFORM%"=="all" (
    set platforms=windows
) else (
    set platforms=%PLATFORM%
)

if "%ARCH%"=="current" (
    set archs=%CURRENT_ARCH%
) else if "%ARCH%"=="all" (
    set archs=amd64 arm64
) else (
    set archs=%ARCH%
)

REM 记录开始时间
for /f "tokens=2 delims==" %%a in ('wmic OS Get localdatetime /value') do set "dt=%%a"
set "YY=%dt:~2,2%" & set "YYYY=%dt:~0,4%" & set "MM=%dt:~4,2%" & set "DD=%dt:~6,2%"
set "HH=%dt:~8,2%" & set "Min=%dt:~10,2%" & set "Sec=%dt:~12,2%"
set "timestamp=%YYYY%-%MM%-%DD% %HH%:%Min%:%Sec%"

call :log_info "开始时间: %timestamp%"

REM 开始构建
for %%p in (%platforms%) do (
    for %%a in (%archs%) do (
        call :build_platform %%p %%a %CLEAN_BUILD%
        if !errorlevel! neq 0 (
            exit /b 1
        )
        call :log_success "✓ %%p/%%a 构建成功"
    )
)

REM 记录结束时间
for /f "tokens=2 delims==" %%a in ('wmic OS Get localdatetime /value') do set "dt=%%a"
set "YY=%dt:~2,2%" & set "YYYY=%dt:~0,4%" & set "MM=%dt:~4,2%" & set "DD=%dt:~6,2%"
set "HH=%dt:~8,2%" & set "Min=%dt:~10,2%" & set "Sec=%dt:~12,2%"
set "timestamp_end=%YYYY%-%MM%-%DD% %HH%:%Min%:%Sec%"

call :log_success "所有构建完成！"
call :log_info "开始时间: %timestamp%"
call :log_info "结束时间: %timestamp_end%"

REM 显示构建结果
call :log_info "构建产物位置:"
if exist "build\bin" (
    dir "build\bin" /B
)

goto :eof

REM 解析命令行参数并执行
:parse_args
if "%~1"=="" goto :main_loop
if "%~1"=="-p" set PLATFORM=%~2& shift & shift & goto :parse_args
if "%~1"=="--platform" set PLATFORM=%~2& shift & shift & goto :parse_args
if "%~1"=="-a" set ARCH=%~2& shift & shift & goto :parse_args
if "%~1"=="--arch" set ARCH=%~2& shift & shift & goto :parse_args
if "%~1"=="-c" set CLEAN_BUILD=true& shift & goto :parse_args
if "%~1"=="--clean" set CLEAN_BUILD=true& shift & goto :parse_args
if "%~1"=="-d" set DEV_MODE=true& shift & goto :parse_args
if "%~1"=="--dev" set DEV_MODE=true& shift & goto :parse_args
if "%~1"=="-o" set OUTPUT_DIR=%~2& shift & shift & goto :parse_args
if "%~1"=="--output" set OUTPUT_DIR=%~2& shift & shift & goto :parse_args
if "%~1"=="-h" goto :show_help
if "%~1"=="--help" goto :show_help
call :log_error "未知参数: %~1"
goto :show_help

REM 主入口点
call :parse_args %*