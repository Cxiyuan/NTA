# Cap Agent Web UI + Backend

现代化Web管理后台 - Vue 3 + Element Plus + Flask

## 🎨 前端技术栈

- **框架**: Vue 3 (Composition API)
- **UI库**: Element Plus
- **状态管理**: Pinia
- **路由**: Vue Router 4
- **图表**: ECharts 5
- **构建工具**: Vite
- **WebSocket**: Socket.IO Client

## 🔧 后端技术栈

- **框架**: Flask
- **实时通信**: Flask-SocketIO
- **CORS**: Flask-CORS
- **配置**: PyYAML

## 📁 项目结构

```
cap_agent/
├── web-ui/                 # 前端项目
│   ├── src/
│   │   ├── views/         # 页面组件
│   │   │   ├── Dashboard.vue      # 实时监控仪表盘
│   │   │   ├── Alerts.vue         # 告警管理
│   │   │   ├── AttackChain.vue    # 攻击链分析
│   │   │   ├── Topology.vue       # 网络拓扑
│   │   │   ├── ThreatIntel.vue    # 威胁情报
│   │   │   ├── Config.vue         # 系统配置
│   │   │   └── Reports.vue        # 报告中心
│   │   ├── components/    # 公共组件
│   │   ├── router/        # 路由配置
│   │   ├── stores/        # Pinia状态管理
│   │   ├── api/           # API接口
│   │   └── assets/        # 静态资源
│   ├── package.json
│   └── vite.config.js
├── backend/               # 后端API服务
│   ├── app.py            # Flask主程序
│   └── requirements.txt
└── README_WEB.md         # 本文档
```

## 🚀 快速开始

### 前端开发

```bash
cd web-ui

# 安装依赖
npm install

# 开发模式
npm run dev

# 构建生产版本
npm run build

# 预览生产版本
npm run preview
```

访问：http://localhost:3000

### 后端服务

```bash
cd backend

# 安装依赖
pip3 install -r requirements.txt

# 启动服务
python3 app.py
```

API服务：http://localhost:5000

## 🌟 核心功能

### 1. 实时监控仪表盘

- **统计卡片**: 严重/高危告警、APT活动、流量处理
- **告警趋势图**: 24小时/7天趋势分析
- **攻击类型分布**: 饼图展示
- **实时告警流**: WebSocket实时推送
- **数据可视化**: ECharts动态图表

### 2. 告警管理

- **告警列表**: 分页、筛选、搜索
- **严重级别**: CRITICAL/HIGH/MEDIUM/LOW
- **置信度展示**: 进度条可视化
- **详情查看**: 完整告警信息
- **告警处置**: 确认/阻断操作
- **批量操作**: 多选处理

### 3. 攻击链分析

- **APT活动列表**: 攻击者IP、风险评分
- **时间线可视化**: 攻击阶段展示
- **Kill Chain跟踪**: 侦察→利用→横向→C2
- **受害主机统计**: 影响范围分析

### 4. 网络拓扑图

- **图形视图**: ECharts力导向图
- **列表视图**: 连接详情表格
- **异常扇出检测**: 异常节点识别
- **多跳链路分析**: 横向移动路径
- **实时刷新**: 动态更新拓扑

### 5. 威胁情报管理

- **IOC库**: IP/域名/哈希/JA3指纹
- **分类管理**: C2/恶意软件/钓鱼/僵尸网络
- **置信度**: 可视化展示
- **情报源更新**: 一键更新
- **手动添加**: 支持自定义IOC

### 6. 系统配置

#### 检测配置
- 横向扫描阈值
- 认证异常配置
- PTH检测窗口

#### 白名单管理
- 监控系统
- 运维网段
- 自动化服务器

#### 决策引擎
- 告警阈值配置
- 业务规则调整

#### ML模型
- 模型训练
- 参数调优
- 模型导出

#### 系统设置
- Zeek配置
- 性能优化
- Worker管理

### 7. 报告中心

- **报告生成**: 日报/周报/月报/自定义
- **内容配置**: 执行摘要/告警详情/统计分析/网络拓扑/APT活动
- **输出格式**: HTML/PDF/JSON
- **在线查看**: 浏览器预览
- **下载导出**: 本地保存

## 🎨 UI设计特点

### 现代化设计

- **响应式布局**: 支持桌面/平板/手机
- **暗黑模式**: 主题切换
- **动画效果**: 平滑过渡
- **卡片设计**: 模块化布局
- **渐变色**: 视觉吸引力

### 交互体验

- **实时更新**: WebSocket推送
- **数据可视化**: ECharts图表
- **加载状态**: Loading提示
- **错误处理**: 友好提示
- **快捷操作**: 批量处理

### 色彩体系

- **主色**: #409EFF (蓝色)
- **成功**: #67C23A (绿色)
- **警告**: #E6A23C (橙色)
- **危险**: #F56C6C (红色)
- **信息**: #909399 (灰色)

## 📡 API接口

### 统计接口

```
GET  /api/stats              # 获取统计数据
GET  /api/stats/trend        # 获取趋势数据
```

### 告警接口

```
GET  /api/alerts             # 获取告警列表
GET  /api/alerts/:id         # 获取告警详情
POST /api/alerts/:id/handle  # 处置告警
```

### 配置接口

```
GET  /api/config             # 获取配置
PUT  /api/config             # 更新配置
GET  /api/config/whitelist   # 获取白名单
PUT  /api/config/whitelist   # 更新白名单
```

### 威胁情报接口

```
GET    /api/threat-intel/iocs       # 获取IOC列表
POST   /api/threat-intel/iocs       # 添加IOC
DELETE /api/threat-intel/iocs/:id   # 删除IOC
POST   /api/threat-intel/update     # 更新情报源
```

### 拓扑接口

```
GET /api/topology/graph      # 获取拓扑图
GET /api/topology/anomalies  # 获取异常分析
```

### 报告接口

```
GET  /api/reports                  # 获取报告列表
POST /api/reports/generate         # 生成报告
GET  /api/reports/:id/download     # 下载报告
```

## 🔌 WebSocket事件

### 客户端监听

```javascript
socket.on('connect', () => {
  console.log('Connected')
})

socket.on('new_alert', (alert) => {
  console.log('New alert:', alert)
})

socket.on('disconnect', () => {
  console.log('Disconnected')
})
```

### 服务端推送

```python
socketio.emit('new_alert', {
  'id': 1001,
  'severity': 'CRITICAL',
  'type': 'PTH攻击',
  'source': '192.168.1.100',
  'target': '10.0.1.50',
  'description': 'Pass-the-Hash攻击检测'
})
```

## 🔐 安全特性

- **CORS配置**: 跨域安全
- **输入验证**: 防止注入
- **错误处理**: 统一异常处理
- **日志记录**: 操作审计

## 📱 响应式设计

- **桌面端**: ≥1200px 完整布局
- **平板端**: 768px-1199px 自适应
- **手机端**: <768px 移动优化

## 🎯 未来计划

- [ ] 用户权限管理
- [ ] SSO单点登录
- [ ] 多语言支持
- [ ] 移动App
- [ ] 更多图表类型
- [ ] AI辅助分析
- [ ] 自动化响应

## 📄 许可证

本项目仅供安全研究和防御使用

---

**技术支持**: Cap Agent Team  
**更新时间**: 2025-12-22
