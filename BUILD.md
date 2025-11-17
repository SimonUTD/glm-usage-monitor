# 构建说明

## 概述

GLM Usage Monitor 提供了跨平台构建脚本，支持生成 macOS 和 Windows 的桌面应用程序。

## 前置要求

### 必需环境
- **Go 1.23+**: [Go 官方安装指南](https://go.dev/doc/install)
- **Node.js 24.7.0+**: [Node.js 官方网站](https://nodejs.org/)
- **Wails v2.10.2+**: [Wails 安装指南](https://wails.io/docs/gettingstarted/installation)

### 系统要求
- **macOS**: 10.15+ (Catalina 或更高版本)
- **Windows**: Windows 10 或更高版本

## 构建方式

### 1. 使用构建脚本（推荐）

#### macOS/Linux
```bash
# 在项目根目录执行
./build.sh --help                    # 查看帮助
./build.sh                           # 构建当前平台
./build.sh -c                        # 清理并构建当前平台
./build.sh -p all -a all             # 构建所有平台和架构
```

#### Windows
```batch
REM 在项目根目录执行
build.bat --help                     # 查看帮助
build.bat                            # 构建当前平台
build.bat -c                         # 清理并构建当前平台
build.bat -p all -a all              # 构建所有平台和架构
```

### 2. 使用 Wails CLI

```bash
cd src                               # 进入 src 目录
wails build                          # 构建当前平台
wails build -platform windows -amd64 # 构建 Windows amd64
wails build -platform darwin -arm64 # 构建 macOS ARM64
wails build -clean                   # 清理并构建
```

## 构建参数说明

| 参数 | 简写 | 说明 | 可选值 | 默认值 |
|------|------|------|--------|--------|
| `--platform` | `-p` | 目标平台 | `darwin`, `windows`, `all` | `current` |
| `--arch` | `-a` | 目标架构 | `amd64`, `arm64`, `all` | `current` |
| `--clean` | `-c` | 构建前清理 | - | `false` |
| `--dev` | `-d` | 开发模式构建 | - | `false` |
| `--output` | `-o` | 输出目录 | 路径 | `./build/bin` |
| `--help` | `-h` | 显示帮助信息 | - | - |

## 构建产物

### macOS
- **文件位置**: `src/build/bin/glm-usage-monitor.app`
- **文件类型**: macOS 应用程序包
- **典型大小**: ~12-15 MB

### Windows
- **文件位置**: `src/build/bin/glm-usage-monitor.exe`
- **文件类型**: Windows 可执行文件
- **典型大小**: ~12-15 MB

## 构建示例

### 场景1：开发测试
```bash
# 快速构建当前平台进行测试
./build.sh

# 或者清理后重新构建
./build.sh -c
```

### 场景2：发布构建
```bash
# 构建所有目标平台
./build.sh -p all -a all
```

### 场景3：跨平台构建
```bash
# 在 macOS 上构建 Windows 版本
./build.sh -p windows -a amd64

# 在 Windows 上构建 macOS 版本需要特殊设置，建议使用 CI/CD
```

## 常见问题

### Q: 构建失败提示 "not found: sql.DB"
**A**: 这是 Wails 的已知提示，不影响构建结果，可以安全忽略。

### Q: 构建时提示 "This darwin build contains the use of private APIs"
**A**: 这是 macOS 开发版本的正常提示，仅用于测试，不会影响应用运行。

### Q: Windows 构建失败
**A**: 确保已安装：
1. [TDM-GCC](https://jmeubank.github.io/tdm-gcc/) 或 MinGW-w64
2. [Git for Windows](https://git-scm.com/download/win)

### Q: 构建产物太大
**A**: 可以通过以下方式优化：
```bash
# 启用压缩
wails build -compress

# 或者使用 -ldflags 减少二进制大小
wails build -ldflags "-s -w"
```

## 开发模式

### 启动开发服务器
```bash
cd src
wails dev
```

开发模式特点：
- 前端热重载
- 实时错误显示
- 调试工具集成
- 浏览器访问: http://localhost:34115

### 调试技巧
1. **前端调试**: 使用浏览器开发者工具
2. **后端调试**: 查看终端输出的 Wails 日志
3. **API 测试**: 使用 http://localhost:34115 进行前后端联调

## 部署说明

### macOS
1. 将 `.app` 文件复制到 `Applications` 目录
2. 首次运行可能需要允许"来自身份不明开发者"的应用
3. 在"系统偏好设置 > 安全性与隐私"中允许运行

### Windows
1. 直接运行 `.exe` 文件
2. 首次运行可能需要通过 Windows Defender 检查
3. 如果出现警告，选择"仍要运行"即可

## 性能优化

### 构建时优化
```bash
# 生产构建，启用优化
wails build -production -compress

# 减小文件大小
wails build -ldflags "-s -w" -compress
```

### 运行时优化
- 应用启动后会创建 `~/.glm-usage-monitor/` 目录
- 数据库文件会自动创建和初始化
- 大量数据同步时建议使用分页功能

## 故障排除

### 环境检查
```bash
# 检查 Go 版本
go version

# 检查 Node.js 版本
node --version

# 检查 Wails 版本
wails version
```

### 清理重建
```bash
# 完全清理
./build.sh -c

# 或手动清理
rm -rf src/build/bin
cd src && wails build -clean
```

### 依赖问题
```bash
# 更新 Go 依赖
cd src && go mod tidy

# 重新安装前端依赖
cd src/frontend && npm install
```