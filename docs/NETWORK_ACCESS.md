# NTA 系统局域网访问配置指南

## 问题说明

用户从局域网其他电脑访问 NTA 系统时，需要确保前端能正确代理 API 请求到后端服务器。

## 解决方案

### 方案一：使用 Docker Compose 部署（推荐）

**特点**：前后端通过 Nginx 反向代理，外部只需访问 80 端口

#### 部署步骤

1. **启动服务**
```bash
cd /root/NTA
docker-compose up -d
```

2. **查看服务状态**
```bash
docker-compose ps
```

3. **局域网访问**
```
前端访问: http://192.168.1.50/
API自动代理: http://192.168.1.50/api → http://nta-server:8080/api
```

#### 工作原理

```
用户电脑 (192.168.1.100)
    ↓
http://192.168.1.50/          ← Nginx (nta-web容器)
    ├─ /          → 返回前端静态文件
    └─ /api/*     → 代理到 nta-server:8080
                        ↓
                   NTA后端服务
```

#### 优点
- ✅ **零配置**：用户无需关心后端地址
- ✅ **跨域无忧**：同源策略，无CORS问题
- ✅ **统一入口**：一个端口访问所有服务
- ✅ **生产级别**：标准的Web应用部署方式

---

### 方案二：开发环境直接访问

**特点**：适合开发调试，前后端分离运行

#### 后端配置

1. **修改后端监听地址**（如需要）
```yaml
# /root/NTA/config/nta.yaml
server:
  host: 0.0.0.0    # 监听所有网络接口
  port: 8080
```

2. **启动后端**
```bash
cd /root/NTA
go run cmd/nta-server/main.go
```

#### 前端配置

1. **修改 `.env.development`**
```bash
# /root/NTA/web/.env.development
VITE_API_URL=http://192.168.1.50:8080
```

2. **启动前端**
```bash
cd /root/NTA/web
npm run dev
```

3. **局域网访问**
```
前端: http://192.168.1.50:3000
API代理: http://192.168.1.50:3000/api → http://192.168.1.50:8080/api
```

#### 优点
- ✅ 热重载开发
- ✅ 调试方便
- ❌ 需要开放两个端口（3000 + 8080）

---

### 方案三：纯前端静态部署 + 后端独立部署

**特点**：前端部署到任意 Web 服务器

#### 1. 构建前端
```bash
cd /root/NTA/web
npm run build
# 产物位于 dist/ 目录
```

#### 2. 部署到 Nginx

**nginx.conf 示例**
```nginx
server {
    listen 80;
    server_name nta.company.com;  # 或使用IP
    
    root /var/www/nta;
    index index.html;
    
    # 前端路由
    location / {
        try_files $uri $uri/ /index.html;
    }
    
    # API反向代理
    location /api {
        proxy_pass http://192.168.1.50:8080;  # 后端服务器地址
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_http_version 1.1;
        proxy_connect_timeout 300s;
        proxy_send_timeout 300s;
        proxy_read_timeout 300s;
    }
}
```

#### 3. 复制文件
```bash
cp -r dist/* /var/www/nta/
nginx -s reload
```

---

## 常见问题排查

### 1. 前端可以访问，但 API 请求失败

**症状**：页面加载正常，但数据无法加载，控制台显示 `ERR_CONNECTION_REFUSED`

**排查步骤**：

```bash
# 1. 检查后端是否启动
docker-compose ps nta-server
# 或
curl http://192.168.1.50:8080/health

# 2. 检查 Nginx 代理配置
docker exec -it nta-web cat /etc/nginx/nginx.conf

# 3. 查看 Nginx 错误日志
docker logs nta-web

# 4. 测试容器间网络连通性
docker exec -it nta-web ping nta-server
```

**解决方案**：
- 确保 `docker-compose.yml` 中 `nta-web` 和 `nta-server` 在同一网络
- 检查防火墙规则（`firewall-cmd --list-all`）

---

### 2. 跨域错误（CORS）

**症状**：浏览器控制台显示 `Access-Control-Allow-Origin` 错误

**原因**：直接访问后端 API 而不通过 Nginx 代理

**解决方案**：
- 方案A：**始终通过 Nginx 访问**（推荐）
  - 访问 `http://192.168.1.50/api/v1/alerts`
  - **不要** 直接访问 `http://192.168.1.50:8080/api/v1/alerts`

- 方案B：后端添加 CORS 头
  ```go
  // internal/api/server.go
  router.Use(cors.New(cors.Config{
      AllowOrigins: []string{"http://192.168.1.50"},
      AllowMethods: []string{"GET", "POST", "PUT", "DELETE"},
      AllowHeaders: []string{"Authorization", "Content-Type"},
  }))
  ```

---

### 3. WebSocket 连接失败

**症状**：实时告警推送不工作

**解决方案**：
确保 Nginx 配置支持 WebSocket 升级（已在 `nginx.conf` 中配置）：
```nginx
proxy_set_header Upgrade $http_upgrade;
proxy_set_header Connection "upgrade";
```

---

### 4. 大文件上传/下载失败

**症状**：下载 PCAP 文件时超时或中断

**解决方案**：
调整 Nginx 超时和缓冲配置（已在 `nginx.conf` 中配置）：
```nginx
client_max_body_size 100M;
proxy_connect_timeout 300s;
proxy_read_timeout 300s;
```

---

## 网络拓扑示意图

### Docker Compose 部署架构

```
┌──────────────────────────────────────────────┐
│           局域网 (192.168.1.0/24)             │
│                                              │
│  ┌─────────────┐      ┌─────────────┐       │
│  │ 用户电脑A    │      │ 用户电脑B    │       │
│  │ .100        │      │ .101        │       │
│  └──────┬──────┘      └──────┬──────┘       │
│         │                    │              │
│         └────────┬───────────┘              │
│                  │                          │
│         http://192.168.1.50/               │
│                  ↓                          │
│  ┌───────────────────────────────────────┐  │
│  │  服务器 (192.168.1.50)                │  │
│  │  ┌─────────────────────────────────┐  │  │
│  │  │ Docker Network: bridge          │  │  │
│  │  │                                 │  │  │
│  │  │  ┌──────────────┐              │  │  │
│  │  │  │ nta-web      │ :80          │  │  │
│  │  │  │ (Nginx)      │◄────┐        │  │  │
│  │  │  └───────┬──────┘     │        │  │  │
│  │  │          │ proxy_pass │        │  │  │
│  │  │          ↓            │        │  │  │
│  │  │  ┌──────────────┐     │        │  │  │
│  │  │  │ nta-server   │:8080│        │  │  │
│  │  │  │ (Go Backend) │─────┘        │  │  │
│  │  │  └───────┬──────┘              │  │  │
│  │  │          │                     │  │  │
│  │  │  ┌───────┴──────┐              │  │  │
│  │  │  │ PostgreSQL   │ :5432        │  │  │
│  │  │  └──────────────┘              │  │  │
│  │  │  ┌──────────────┐              │  │  │
│  │  │  │ Redis        │ :6379        │  │  │
│  │  │  └──────────────┘              │  │  │
│  │  └─────────────────────────────────┘  │  │
│  └───────────────────────────────────────┘  │
└──────────────────────────────────────────────┘
```

---

## 端口映射说明

| 服务          | 容器内端口 | 宿主机端口 | 外部访问          |
|--------------|-----------|-----------|------------------|
| nta-web      | 80        | 80        | http://IP/       |
| nta-server   | 8080      | 8080      | (内部，不直接访问)  |
| PostgreSQL   | 5432      | 5432      | (内部，可选开放)   |
| Redis        | 6379      | 6379      | (内部，可选开放)   |
| Prometheus   | 9090      | 9090      | http://IP:9090   |
| Grafana      | 3000      | 3000      | http://IP:3000   |

**生产环境建议**：
- ✅ 开放：80 (前端)
- ⚠️ 可选开放：9090 (Prometheus)、3000 (Grafana) - 仅限运维访问
- ❌ 禁止开放：8080、5432、6379 - 仅容器内部通信

---

## 防火墙配置

### CentOS/RHEL (firewalld)

```bash
# 开放HTTP端口
firewall-cmd --permanent --add-service=http
firewall-cmd --permanent --add-port=80/tcp

# 可选：开放监控端口
firewall-cmd --permanent --add-port=3000/tcp  # Grafana
firewall-cmd --permanent --add-port=9090/tcp  # Prometheus

# 重载配置
firewall-cmd --reload
```

### Ubuntu (ufw)

```bash
ufw allow 80/tcp
ufw allow 3000/tcp
ufw allow 9090/tcp
ufw reload
```

---

## 测试验证

### 1. 健康检查
```bash
# 前端
curl http://192.168.1.50/health

# 后端
curl http://192.168.1.50/api/v1/health
```

### 2. 完整功能测试
```bash
# 从局域网其他电脑测试
curl -X POST http://192.168.1.50/api/v1/alerts \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{"page":1}'
```

### 3. 浏览器测试
1. 打开 `http://192.168.1.50/`
2. 按 F12 打开开发者工具
3. 切换到 Network 标签
4. 登录系统，查看 API 请求
5. 确认请求地址为 `/api/v1/*`（相对路径）

---

## 总结

✅ **当前配置已完全支持局域网访问**：

1. **Docker Compose 部署**（已配置完成）
   - Nginx 自动反向代理
   - 容器内部网络互通
   - 外部只需访问 80 端口

2. **开发模式**（已配置完成）
   - `vite.config.ts` 监听 `0.0.0.0`
   - 支持环境变量配置后端地址
   - 开发服务器可局域网访问

3. **生产部署**（已提供完整配置）
   - `nginx.conf` 标准反向代理配置
   - 支持大文件传输（100MB）
   - WebSocket 长连接支持
   - Gzip 压缩优化

**用户现在可以直接使用：**
```bash
# 启动系统
docker-compose up -d

# 访问地址（假设服务器IP为192.168.1.50）
浏览器打开: http://192.168.1.50/
```

所有 API 请求会自动通过 Nginx 代理到后端，无需额外配置！
