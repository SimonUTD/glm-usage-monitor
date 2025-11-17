#!/bin/bash

# GLM Usage Monitor 跨平台构建脚本
# 支持 macOS 和 Windows 平台构建

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 日志函数
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 显示帮助信息
show_help() {
    echo "GLM Usage Monitor 构建脚本"
    echo ""
    echo "用法: $0 [选项]"
    echo ""
    echo "选项:"
    echo "  -p, --platform PLATFORM    目标平台 (darwin|windows|all) (默认: current)"
    echo "  -a, --arch ARCH             目标架构 (amd64|arm64|all) (默认: current)"
    echo "  -c, --clean                 构建前清理"
    echo "  -d, --dev                   开发模式构建"
    echo "  -o, --output DIR            输出目录 (默认: ./build/bin)"
    echo "  -h, --help                  显示此帮助信息"
    echo ""
    echo "示例:"
    echo "  $0                          # 构建当前平台"
    echo "  $0 -p all -a all            # 构建所有平台和架构"
    echo "  $0 -p windows -a amd64      # 构建 Windows amd64"
    echo "  $0 -c                       # 清理并构建当前平台"
}

# 默认参数
PLATFORM="current"
ARCH="current"
CLEAN_BUILD=false
DEV_MODE=false
OUTPUT_DIR="./build/bin"

# 解析命令行参数
while [[ $# -gt 0 ]]; do
    case $1 in
        -p|--platform)
            PLATFORM="$2"
            shift 2
            ;;
        -a|--arch)
            ARCH="$2"
            shift 2
            ;;
        -c|--clean)
            CLEAN_BUILD=true
            shift
            ;;
        -d|--dev)
            DEV_MODE=true
            shift
            ;;
        -o|--output)
            OUTPUT_DIR="$2"
            shift 2
            ;;
        -h|--help)
            show_help
            exit 0
            ;;
        *)
            log_error "未知参数: $1"
            show_help
            exit 1
            ;;
    esac
done

# 获取当前平台和架构
detect_current_platform() {
    case "$(uname -s)" in
        Darwin*)
            CURRENT_PLATFORM="darwin"
            CURRENT_ARCH="$(uname -m)"
            # 将 ARM64 转换为标准格式
            if [[ "$CURRENT_ARCH" == "arm64" ]]; then
                CURRENT_ARCH="arm64"
            fi
            ;;
        Linux*)
            CURRENT_PLATFORM="linux"
            CURRENT_ARCH="$(uname -m)"
            ;;
        CYGWIN*|MINGW*|MSYS*)
            CURRENT_PLATFORM="windows"
            CURRENT_ARCH="$(uname -m)"
            ;;
        *)
            log_error "不支持的操作系统: $(uname -s)"
            exit 1
            ;;
    esac
}

# 执行构建
build_platform() {
    local platform=$1
    local arch=$2
    local clean_flag=$3

    log_info "开始构建平台: $platform/$arch"

    # 构建 Wails 应用
    local build_cmd="wails build"
    local platform_target="$platform/$arch"

    if [[ "$platform" == "current" && "$arch" == "current" ]]; then
        platform_target=""
    fi

    if [[ -n "$platform_target" ]]; then
        build_cmd="$build_cmd -platform $platform_target"
    fi

    if [[ "$clean_flag" == true ]]; then
        build_cmd="$build_cmd -clean"
    fi

    if [[ "$DEV_MODE" == true ]]; then
        build_cmd="$build_cmd -debug"
    fi

    log_info "执行命令: $build_cmd"

    if eval "$build_cmd"; then
        log_success "构建完成: $platform/$arch"

        # 获取构建产物信息
        local output_file
        case $platform in
            darwin)
                if [[ -f "build/bin/glm-usage-monitor.app/Contents/MacOS/glm-usage-monitor" ]]; then
                    output_file="build/bin/glm-usage-monitor.app"
                    local size=$(du -sh "$output_file" | cut -f1)
                    log_info "构建产物: $output_file ($size)"
                fi
                ;;
            windows)
                if [[ -f "build/bin/glm-usage-monitor.exe" ]]; then
                    output_file="build/bin/glm-usage-monitor.exe"
                    local size=$(du -sh "$output_file" | cut -f1)
                    log_info "构建产物: $output_file ($size)"
                fi
                ;;
        esac
    else
        log_error "构建失败: $platform/$arch"
        return 1
    fi
}

# 主函数
main() {
    log_info "开始 GLM Usage Monitor 构建流程"

    # 检查是否在正确的目录
    if [[ ! -f "wails.json" ]]; then
        log_error "请在项目根目录（包含 wails.json 的目录）中运行此脚本"
        exit 1
    fi

    # 检查依赖
    if ! command -v wails &> /dev/null; then
        log_error "未找到 Wails CLI，请先安装: https://wails.io/docs/gettingstarted/installation"
        exit 1
    fi

    if ! command -v go &> /dev/null; then
        log_error "未找到 Go，请先安装: https://go.dev/doc/install"
        exit 1
    fi

    # 检测当前平台
    detect_current_platform
    log_info "当前平台: $CURRENT_PLATFORM/$CURRENT_ARCH"

    # 解析平台和架构
    local platforms=()
    local archs=()

    if [[ "$PLATFORM" == "current" ]]; then
        platforms+=("$CURRENT_PLATFORM")
    elif [[ "$PLATFORM" == "all" ]]; then
        platforms+=("darwin" "windows")
    else
        platforms+=("$PLATFORM")
    fi

    if [[ "$ARCH" == "current" ]]; then
        archs+=("$CURRENT_ARCH")
    elif [[ "$ARCH" == "all" ]]; then
        archs+=("amd64" "arm64")
    else
        archs+=("$ARCH")
    fi

    # 开始构建
    local build_start_time=$(date +%s)

    for platform in "${platforms[@]}"; do
        for arch in "${archs[@]}"; do
            if build_platform "$platform" "$arch" "$CLEAN_BUILD"; then
                log_success "✓ $platform/$arch 构建成功"
            else
                log_error "✗ $platform/$arch 构建失败"
                exit 1
            fi
        done
    done

    local build_end_time=$(date +%s)
    local build_duration=$((build_end_time - build_start_time))

    log_success "所有构建完成！耗时: ${build_duration}s"

    # 显示构建结果
    log_info "构建产物位置:"
    if [[ -d "build/bin" ]]; then
        ls -la build/bin/
    fi
}

# 执行主函数
main "$@"