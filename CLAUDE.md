# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 项目概述

这是一个基于 Wails v2 框架开发的桌面应用程序，用于统计智谱AI GLM Coding Plan 的账单信息。项目从原来的 Node.js 后端架构转换为 Go 后端 + Vue3 前端的桌面应用，支持编译为 Windows 和 macOS 可执行文件。

## 技术栈

- **后端框架**: Wails v2 (Go 1.23)
- **前端框架**: Vue 3 + Vite
- **UI 组件库**: Element Plus
- **图表库**: ECharts + Chart.js + vue-echarts + vue-chartjs
- **数据库**: SQLite3
- **HTTP 客户端**: Axios
- **日期处理**: Day.js

## 开发环境要求

- Go 1.23+
- Node.js 24.7.0+
- npm 11.5.1+
- Wails v2.10.2+

## 常用开发命令

### 前端开发
```bash
cd src/frontend
npm install          # 安装依赖
npm run dev          # 启动开发服务器
npm run build        # 构建生产版本
npm run preview      # 预览构建结果
```

### Wails 应用开发
```bash
cd src
wails dev            # 启动开发模式（前端热重载 + Go 后端）
wails build          # 构建生产版本（生成桌面应用）
wails build -clean   # 清理并重新构建
```

### 模块安装
```bash
cd src/frontend
npm install [package-name]    # 安装前端依赖
go mod tidy                    # 整理 Go 依赖
go get [package-name]         # 安装 Go 依赖
```

## 项目架构

### 目录结构
```
src/
├── app.go              # 主应用结构（后端业务逻辑）
├── main.go             # 应用入口点
├── wails.json          # Wails 配置文件
├── go.mod              # Go 模块依赖
├── frontend/           # Vue3 前端代码
│   ├── src/
│   │   ├── views/      # 页面组件（Stats, Bills, Sync, Settings, Onboarding）
│   │   ├── components/ # 公共组件
│   │   ├── api/        # API 调用封装
│   │   ├── utils/      # 工具函数
│   │   └── composables/ # Vue3 组合式函数
│   ├── package.json    # 前端依赖配置
│   └── vite.config.js  # Vite 构建配置
└── build/             # 构建输出目录
```

### 应用配置
- **应用名称**: glm-usage-monitor
- **窗口尺寸**: 1024x768
- **背景色**: RGBA(27, 38, 54, 1)
- **前端开发服务器**: 自动检测端口
- **资源嵌入**: 使用 Go embed 嵌入前端构建产物

## 核心功能模块

### 1. 数据同步模块 (Sync.vue)
- 调用智谱AI API 获取账单数据
- 支持增量同步和全量同步
- 实时进度监控和状态展示
- 异步处理避免界面阻塞

### 2. 统计展示模块 (Stats.vue)
- 近5小时调用次数统计
- Token 使用量分析
- 会员等级限制提醒
- 可视化图表展示（ECharts）

### 3. 账单管理模块 (Bills.vue)
- 账单列表展示和筛选
- 详细账单信息查看
- 数据导出功能

### 4. 设置管理模块 (Settings.vue)
- API Token 配置
- 自动同步设置
- 数据库管理

### 5. 引导页面模块 (Onboarding.vue)
- 初次使用引导
- 基础配置设置

## 数据库设计

### 数据存储位置
- **macOS**: `~/.glm-usage-monitor/`
- **Windows**: `%USERPROFILE%\.glm-usage-monitor\`

### 核心表结构
- **expense_bills**: 账单明细表（包含智谱AI返回的所有字段）
- **api_tokens**: API Token 配置表
- **sync_history**: 同步历史记录表
- **auto_sync_config**: 自动同步配置表
- **membership_tier_limits**: 会员等级限制表

### 重要字段说明
- `time_window`: 从 API 返回的原始时间窗口字段
- `time_window_start/time_window_end`: 拆分后的时间窗口
- `transaction_time`: 从 billingNo 提取的13位时间戳转换的时间
- `create_time`: 记录插入数据库的时间

## API 集成

### 智谱AI账单查询接口
```bash
curl --request GET \
  --url 'https://bigmodel.cn/api/finance/expenseBill/expenseBillList?billingMonth=2025-11&pageNum=1&pageSize=20' \
  --header 'Authorization: Bearer 用户的Token'
```

### 数据处理要求
1. **timeWindow 拆分**: 将返回的 timeWindow 字段拆分为 start 和 end 两个字段
2. **时间戳提取**: 从 billingNo 字段截取 customerId 之后的13位时间戳，转换为 transaction_time
3. **字段映射**: API 返回字段与数据库表字段一一对应
4. **记录创建时间**: 每条记录都需要 create_time 字段

## 前端开发规范

### 样式约定
- **主色调**: #4D6782（莫兰迪色系）
- **背景**: 渐变色 #f5f5f3 0% 到 #e8e6e1 100%
- **毛玻璃效果**: header 使用 backdrop-filter: blur(8px)
- **圆角规范**: 统一使用 Element Plus 的圆角样式
- **禁止使用**: 严格禁止在代码、文档、UI中使用 emoji 图标

### 路由结构
- `/stats`: 统计信息页面
- `/bills`: 账单列表页面
- `/sync`: 数据同步页面
- `/settings`: 设置页面
- `/onboarding`: 引导页面（无 header）

### 状态管理
- 使用 Vue3 Composition API
- 通过 wailsjs/go 目录调用后端 Go 方法
- 前后端数据传递使用 JSON 格式

## 后端开发规范

### Go 方法绑定
所有需要从前端调用的 Go 方法都需要在 `main.go` 的 `Bind` 数组中注册：

```go
Bind: []interface{}{
    app,
},
```

### 方法命名约定
- 使用 PascalCase 命名（Go 导出方法约定）
- 方法名应体现功能，如 `SyncBills`, `GetStats`, `SaveSettings`
- 错误处理应返回详细的错误信息

### 数据库操作
- 使用 Go 的 database/sql 包操作 SQLite
- 所有 SQL 语句应使用参数化查询防止注入
- 事务处理确保数据一致性
- 适当的错误处理和日志记录

## 开发注意事项

### 性能优化
- 大量数据同步时使用分页和异步处理
- 图表数据应进行适当的聚合和缓存
- 数据库查询添加必要索引

### 用户体验
- 提供实时进度反馈
- 友好的错误提示信息
- 响应式布局适配不同屏幕尺寸

### 安全考虑
- API Token 安全存储（不提交到版本控制）
- 数据库文件权限控制
- 输入数据验证和清理

## 构建和部署

### 开发模式
```bash
wails dev  # 启动开发服务器，支持前端热重载
```

### 生产构建
```bash
wails build              # 构建当前平台的应用
wails build -platform windows/darwin -amd64  # 跨平台构建
```

### 构建产物
- **Windows**: `.exe` 可执行文件
- **macOS**: `.app` 应用包
- 前端资源已嵌入到可执行文件中，无需额外部署

## 故障排除

### 常见问题
1. **前端依赖安装失败**: 检查 Node.js 版本是否为 24.7.0+
2. **Wails 构建失败**: 确保 Go 版本为 1.23+ 并且 Wails CLI 已正确安装
3. **API 调用失败**: 检查 Token 配置和网络连接
4. **数据库错误**: 确保 `~/.glm-usage-monitor/` 目录存在且有写权限

### 调试技巧
- 使用 `wails dev` 进行开发调试
- 浏览器开发者工具调试前端（http://localhost:34115）
- 查看 Wails 日志输出定位后端问题