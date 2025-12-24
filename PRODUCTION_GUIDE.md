# NTA 生产环境部署指南

## 快速部署

### 一键启动

```bash
cd /root/NTA
docker-compose up -d
```

### 访问系统

```
前端地址: http://YOUR_SERVER_IP/
例如: http://192.168.1.50/
```

---

## 架构说明

```
┌─────────────────────────────────────────────┐
│           用户浏览器 (局域网任意PC)            │
└────────────────┬────────────────────────────┘
                 │
                 │ http://192.168.1.50/
                 ↓
┌─────────────────────────────────────────────┐
│          Docker Host (192.168.1.50)         │
│                                             │
│  ┌───────────────────────────────────────┐ │
│  │  nta-web (Nginx) - 端口 80           │ │
│  │  ┌─────────────────────────────────┐ │ │
│  │  │  静态文件: /                    │ │ │
│  │  │  API代理: /api → nta-server:8080│ │ │
│  │  └─────────────────────────────────┘ │ │
│  └───────────────┬───────────────────────┘ │
│                  │                         │
│  ┌───────────────▼───────────────────────┐ │
│  │  nta-server (Go) - 端口 8080        │ │
│  └───────────────┬───────────────────────┘ │
│                  │                         │
│      ┌───────────┼───────────┐            │
│      │           │           │            │
│  ┌───▼────┐ ┌───▼────┐ ┌───▼────┐       │
│  │Postgres│ │ Redis  │ │  PCAP  │       │
│  │ :5432  │ │ :6379  │ │ Volume │       │
│  └────────┘ └────────┘ └────────┘       │
└─────────────────────────────────────────────┘
```

---

## 服务说明

| 容器名 | 端口映射 | 用途 | 外部访问 |
|--------|---------|------|---------|
| nta-web | 80:80 | 前端+反向代理 | ✅ http://IP/ |
| nta-server | 8080:8080 | 后端API | ❌ 仅内部 |
| nta-postgres | 5432:5432 | 数据库 | ❌ 仅内部 |
| nta-redis | 6379:6379 | 缓存 | ❌ 仅内部 |
| nta-prometheus | 9090:9090 | 监控指标 | ⚠️ 可选开放 |
| nta-grafana | 3000:3000 | 监控面板 | ⚠️ 可选开放 |

---

## 配置优化

### 1. 数据库配置

**首次启动后，创建配置文件**:

```bash
mkdir -p /root/NTA/config
cat > /root/NTA/config/nta.yaml <<EOF
server:
  host: 0.0.0.0
  port: 8080
  mode: release

database:
  type: postgres
  dsn: "host=nta-postgres user=nta password=nta_password dbname=nta port=5432 sslmode=disable"

redis:
  addr: nta-redis:6379
  password: ""
  db: 0

security:
  jwt_secret: "CHANGE_THIS_SECRET_KEY_MIN_32_CHARACTERS_LONG"
  enable_tls: false
  rate_limit_requests: 100
  rate_limit_window: 60

backup:
  enabled: true
  backup_dir: /app/backups
  interval_hours: 24
  retention_days: 7
EOF
```

### 2. 修改默认密码

```bash
# PostgreSQL 密码
docker-compose down
vi docker-compose.yml  # 修改 POSTGRES_PASSWORD

# JWT 密钥
vi /root/NTA/config/nta.yaml  # 修改 jwt_secret
```

### 3. 防火墙配置

```bash
# CentOS/RHEL
firewall-cmd --permanent --add-port=80/tcp
firewall-cmd --reload

# Ubuntu
ufw allow 80/tcp
ufw reload

# 验证
firewall-cmd --list-ports  # CentOS
# 或
ufw status  # Ubuntu
```

---

## 运维命令

### 查看服务状态

```bash
docker-compose ps
```

### 查看日志

```bash
# 所有服务
docker-compose logs -f

# 特定服务
docker-compose logs -f nta-server
docker-compose logs -f nta-web

# 最近100行
docker-compose logs --tail=100 nta-server
```

### 重启服务

```bash
# 重启所有服务
docker-compose restart

# 重启单个服务
docker-compose restart nta-server
docker-compose restart nta-web
```

### 更新部署

```bash
# 拉取最新代码
cd /root/NTA
git pull

# 重新构建并启动
docker-compose up -d --build
```

### 数据备份

```bash
# 备份 PostgreSQL 数据库
docker exec -t nta-postgres pg_dump -U nta nta > nta_backup_$(date +%Y%m%d).sql

# 备份 PCAP 文件
tar -czf pcap_backup_$(date +%Y%m%d).tar.gz -C /var/lib/docker/volumes/nta_nta-pcap/_data .

# 备份配置文件
tar -czf config_backup_$(date +%Y%m%d).tar.gz /root/NTA/config
```

### 数据恢复

```bash
# 恢复数据库
cat nta_backup_20250101.sql | docker exec -i nta-postgres psql -U nta -d nta

# 恢复 PCAP 文件
tar -xzf pcap_backup_20250101.tar.gz -C /var/lib/docker/volumes/nta_nta-pcap/_data
```

---

## 性能优化

### 1. Nginx 调优

已在 `nginx.conf` 中配置:
- ✅ Gzip 压缩
- ✅ 静态文件缓存
- ✅ 长连接支持
- ✅ 100MB 文件上传限制
- ✅ 300秒代理超时

### 2. 资源限制

编辑 `docker-compose.yml` 添加资源限制:

```yaml
services:
  nta-server:
    deploy:
      resources:
        limits:
          cpus: '2'
          memory: 2G
        reservations:
          cpus: '1'
          memory: 512M
```

### 3. PostgreSQL 调优

```bash
docker exec -it nta-postgres bash
# 进入容器后编辑 /var/lib/postgresql/data/postgresql.conf
vi /var/lib/postgresql/data/postgresql.conf

# 推荐配置
shared_buffers = 256MB
effective_cache_size = 1GB
maintenance_work_mem = 64MB
checkpoint_completion_target = 0.9
wal_buffers = 16MB
max_connections = 100
```

---

## 监控告警

### Prometheus 指标

访问: `http://192.168.1.50:9090`

**关键指标**:
- `nta_http_requests_total` - HTTP请求总数
- `nta_alerts_total` - 告警总数
- `nta_active_probes` - 活跃探针数
- `nta_packets_processed_total` - 处理数据包数

### Grafana 看板

访问: `http://192.168.1.50:3000`
- 用户名: `admin`
- 密码: `admin` (首次登录后修改)

---

## 常见问题

### 1. 容器无法启动

```bash
# 查看错误日志
docker-compose logs nta-server

# 检查端口占用
netstat -tulnp | grep -E '80|8080|5432|6379'

# 清理重启
docker-compose down
docker-compose up -d
```

### 2. 前端可以访问但无数据

```bash
# 检查后端健康状态
curl http://192.168.1.50/api/v1/health

# 检查数据库连接
docker-compose logs nta-server | grep -i "database\|postgres"

# 检查 Nginx 代理
docker exec -it nta-web cat /etc/nginx/nginx.conf
```

### 3. 数据库连接失败

```bash
# 检查 PostgreSQL 状态
docker-compose ps nta-postgres

# 手动测试连接
docker exec -it nta-postgres psql -U nta -d nta

# 查看 PostgreSQL 日志
docker-compose logs nta-postgres
```

### 4. 磁盘空间不足

```bash
# 清理 PCAP 旧文件（保留30天）
find /var/lib/docker/volumes/nta_nta-pcap/_data -name "*.pcap" -mtime +30 -delete

# 清理 Docker 缓存
docker system prune -a --volumes

# 清理日志
truncate -s 0 /var/lib/docker/containers/*/*-json.log
```

---

## 安全加固

### 1. 启用 HTTPS

```bash
# 安装 Certbot
yum install certbot -y  # CentOS
apt install certbot -y  # Ubuntu

# 获取证书
certbot certonly --standalone -d nta.yourdomain.com

# 修改 Nginx 配置
vi /root/NTA/web/nginx.conf
# 添加 SSL 配置 (见下方示例)
```

**Nginx HTTPS 配置示例**:
```nginx
server {
    listen 443 ssl http2;
    server_name nta.yourdomain.com;

    ssl_certificate /etc/letsencrypt/live/nta.yourdomain.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/nta.yourdomain.com/privkey.pem;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers HIGH:!aNULL:!MD5;

    # 其他配置同 HTTP
}

server {
    listen 80;
    server_name nta.yourdomain.com;
    return 301 https://$server_name$request_uri;
}
```

### 2. IP 访问控制

编辑 `nginx.conf` 添加 IP 白名单:

```nginx
location / {
    allow 192.168.1.0/24;  # 允许局域网
    deny all;              # 拒绝其他
}
```

### 3. 限流防护

已在 `nginx.conf` 中配置:
```nginx
limit_req_zone $binary_remote_addr zone=api_limit:10m rate=10r/s;
limit_req zone=api_limit burst=20 nodelay;
```

---

## 系统要求

### 最低配置
- CPU: 2核
- 内存: 4GB
- 磁盘: 50GB
- 操作系统: CentOS 7+ / Ubuntu 20.04+

### 推荐配置
- CPU: 4核
- 内存: 8GB
- 磁盘: 200GB SSD
- 网络: 千兆网卡

---

## 健康检查

```bash
# 前端健康检查
curl http://192.168.1.50/health

# 后端健康检查
curl http://192.168.1.50/api/v1/health

# 数据库健康检查
docker exec nta-postgres pg_isready -U nta

# Redis 健康检查
docker exec nta-redis redis-cli ping
```

---

## 升级指南

```bash
# 1. 备份数据
docker exec -t nta-postgres pg_dump -U nta nta > backup.sql

# 2. 停止服务
docker-compose down

# 3. 拉取新代码
git pull

# 4. 重新构建
docker-compose build

# 5. 启动服务
docker-compose up -d

# 6. 检查状态
docker-compose ps
docker-compose logs -f
```

---

## 技术支持

- 项目文档: `/root/NTA/docs/`
- 架构说明: `/root/NTA/docs/ARCHITECTURE.md`
- API 文档: `/root/NTA/docs/API.md`
- 网络配置: `/root/NTA/docs/NETWORK_ACCESS.md`
