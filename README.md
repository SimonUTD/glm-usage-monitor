# 智谱AI GLM Coding Plan 账单统计系统

一个专门用于智谱AI GLM Coding Plan套餐的账单管理和统计分析系统，帮助用户实时监控API使用量、Token消耗和费用支出。

## 系统特色

- **专注智谱AI**: 深度适配智谱AI GLM Coding Plan套餐
- **实时监控**: 近5小时/1天/1周/1月的API调用和Token消耗统计
- **智能同步**: 支持全量和增量同步，自动避免重复数据
- **数据本地化**: 使用SQLite3本地数据库，确保数据隐私安全
- **一键启动**: 提供便捷的启动脚本和Docker部署方案
- **丰富图表**: 多种图表类型展示，数据可视化直观

## 核心功能
- **数据同步模块**: 全量/增量同步账单数据，支持异步处理和进度监控
- **统计分析模块**: 多维度统计API使用量、Token消耗和费用支出
- **自动监控模块**: 定时同步数据，实时预警套餐使用情况
- **用户界面模块**: 响应式设计，支持多屏幕尺寸

## 使用指南

### 1. 初始配置
1. 首次使用会自动跳转到设置页面
2. 输入您的智谱AI API Token进行配置

### 2. 数据同步
1. 进入"数据同步"页面
2. 选择要同步的账单月份
3. 点击"开始同步"，系统会自动：
   - 调用智谱AI API获取账单数据
   - 智能处理timeWindow时间窗口
   - 从billingNo解析交易时间戳
   - 避免重复数据入库

### 3. 查看统计
系统提供多维度数据统计：
- **API使用统计**: 近5小时/1天/1周/1月的调用次数和增长率
- **Token消耗统计**: 输入/输出Token使用量和进度条显示
- **费用支出统计**: 累计花费金额和环比分析
- **产品分布统计**: 不同API产品的使用情况

### 4. 自动监控
- 配置自动同步频率（建议10秒）
- 系统会定期更新数据并实时显示统计结果
- 接近套餐限制时自动预警

## API接口

### 核心接口
```bash
# 同步账单数据
POST /api/bills/sync
Body: {"billingMonth": "2025-11"}

# 获取账单列表（分页）
GET /api/bills?page=1&pageSize=20&startDate=2025-11-01&endDate=2025-11-30

# 获取统计信息
GET /api/bills/stats?period=5h

# 获取同步状态
GET /api/bills/sync-status

# 获取API使用进度
GET /api/bills/api-usage-progress
```

### 配置接口
```bash
# 保存API Token
POST /api/tokens/save
Body: {"token": "your-api-token"}

# 配置自动同步
POST /api/auto-sync/config
Body: {"enabled": true, "interval": 10}
```

## 常见问题

### Q: 数据同步失败怎么办？
A: 检查API Token是否正确，网络连接是否正常，查看日志文件了解详细错误信息。

