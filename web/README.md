# NTA Web Dashboard

基于 React + TypeScript + Ant Design + Vite 构建的网络流量分析系统前端界面。

## 生产环境部署

### Docker 部署（推荐）

项目已配置 Docker 容器化部署，前端通过 Nginx 提供服务并反向代理后端 API。

```bash
# 在项目根目录执行
cd /root/NTA
docker-compose up -d
```

访问地址: `http://YOUR_SERVER_IP/`

### 手动构建部署

```bash
# 1. 安装依赖
npm install

# 2. 生产构建
npm run build

# 3. 部署到 Nginx
cp -r dist/* /var/www/nta/
```

## 功能特性

- 🎯 实时安全态势大屏
- 🚨 安全告警管理与处置
- 💻 网络资产可视化
- 🔍 威胁情报查询
- 🛡️ 高级威胁检测（DGA/DNS隧道/C2/WebShell）
- 📦 PCAP流量回溯与下载
- 📊 安全报表生成与导出
- ⚙️ 探针状态监控
- 📧 通知配置管理

## 技术栈

- **框架**: React 18 + TypeScript
- **构建工具**: Vite
- **UI 组件**: Ant Design 5
- **图表**: ECharts
- **HTTP 客户端**: Axios
- **路由**: React Router v6

## 项目结构

```
web/
├── public/              # 静态资源
├── src/
│   ├── components/      # 通用组件
│   │   └── Layout.tsx   # 主布局
│   ├── pages/           # 页面组件
│   │   ├── Login.tsx    # 登录页
│   │   ├── Dashboard.tsx # 态势大屏
│   │   ├── Alerts.tsx   # 告警管理
│   │   ├── Assets.tsx   # 资产管理
│   │   ├── ThreatIntel.tsx # 威胁情报
│   │   ├── AdvancedDetection.tsx # 高级检测
│   │   ├── PcapAnalysis.tsx # PCAP回溯
│   │   ├── Reports.tsx  # 报表中心
│   │   ├── Probes.tsx   # 探针管理
│   │   └── Settings.tsx # 系统设置
│   ├── services/        # API 服务
│   │   └── api.ts       # API 封装
│   ├── App.tsx          # 应用入口
│   ├── main.tsx         # 主入口
│   └── index.css        # 全局样式
├── nginx.conf           # Nginx 配置
├── Dockerfile           # Docker 构建文件
├── index.html
├── package.json
├── tsconfig.json
└── vite.config.ts
```

## Nginx 反向代理

生产环境下，前端静态文件由 Nginx 提供，API 请求通过反向代理转发到后端。

**关键配置** (`nginx.conf`):
```nginx
location / {
    try_files $uri $uri/ /index.html;
}

location /api {
    proxy_pass http://nta-server:8080;
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
}
```

## 构建优化

生产构建已优化:
- ✅ 代码压缩 (Terser)
- ✅ Tree Shaking
- ✅ 代码分割 (React/Antd/ECharts 独立chunk)
- ✅ 移除 console 和 debugger
- ✅ Gzip 压缩（Nginx层）

## API 调用示例

```typescript
import { alertAPI } from '@/services/api'

// 获取告警列表
const alerts = await alertAPI.list({ page: 1, page_size: 50 })

// 更新告警状态
await alertAPI.update(123, { status: 'resolved' })
```

## 浏览器支持

- Chrome >= 90
- Firefox >= 88
- Safari >= 14
- Edge >= 90

## License

MIT
