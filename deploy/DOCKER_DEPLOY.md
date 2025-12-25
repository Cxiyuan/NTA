# NTA Docker 部署指南

## 部署前准备

### 1. 系统要求
- Docker 20.10+
- Docker Compose 2.0+
- 至少 4GB 内存
- 至少 20GB 磁盘空间

### 2. 获取镜像

#### 方式一：从 GitHub Actions 下载（推荐）
1. 在 GitHub Actions 中下载构建产物
2. 导入镜像：
```bash
docker load -i nta-server-v1.0.0.tar
docker load -i nta-web-v1.0.0.tar
```

#### 方式二：本地构建
```bash
docker build -t nta-server:v1.0.0 -f Dockerfile .
docker build -t nta-web:v1.0.0 -f web/Dockerfile web/
```

## 快速部署

### 1. 准备配置文件
```bash
# 如果 config/nta.yaml 不存在，会自动从 config/nta.yaml.example 复制
# 或手动复制：
cp config/nta.yaml.example config/nta.yaml
```

### 2. 检查配置
确保 `config/nta.yaml` 中使用容器名称而不是 localhost：
```yaml
redis:
  addr: nta-redis:6379  # 使用容器名

database:
  type: postgres
  dsn: host=nta-postgres user=nta password=nta_password dbname=nta port=5432 sslmode=disable  # 使用容器名
```

### 3. 运行部署脚本
```bash
cd /path/to/NTA
./deploy/docker-deploy.sh
```

## 手动部署

### 1. 启动服务
```bash
docker-compose up -d
```

### 2. 查看状态
```bash
docker-compose ps
```

### 3. 查看日志
```bash
# 查看所有服务日志
docker-compose logs -f

# 查看特定服务日志
docker-compose logs -f nta-server
docker logs nta-server --tail 100
```

## 服务访问

- **API Server**: http://localhost:8080
- **Web UI**: http://localhost:80
- **Grafana**: http://localhost:3000 (admin/admin)
- **Prometheus**: http://localhost:9090

## 常见问题排查

### nta-server 容器持续重启

#### 问题现象
```bash
docker ps -a | grep nta-server
# 显示 Restarting (1) xxx seconds ago
```

#### 排查步骤

1. **查看容器日志**
```bash
docker logs nta-server --tail 100
```

2. **检查配置文件**
确认配置文件存在且正确：
```bash
ls -la config/nta.yaml
cat config/nta.yaml | grep -E "(addr|dsn)"
```

3. **检查数据库连接**
确保使用容器名而不是 localhost：
```yaml
# ❌ 错误
database:
  dsn: host=localhost user=nta ...

# ✓ 正确
database:
  dsn: host=nta-postgres user=nta ...
```

4. **检查网络连接**
```bash
# 查看容器网络
docker network inspect nta_nta-network

# 测试容器间连接
docker exec nta-server ping -c 2 nta-postgres
docker exec nta-server ping -c 2 nta-redis
```

5. **检查依赖服务健康状态**
```bash
docker ps --filter "name=nta-postgres"
docker ps --filter "name=nta-redis"

# 查看健康检查
docker inspect nta-postgres | grep -A 10 Health
```

### 数据库连接失败

**错误信息**：
```
failed to connect to `host=localhost user=nta database=nta`: dial error
```

**解决方案**：
修改 `config/nta.yaml`：
```yaml
database:
  dsn: host=nta-postgres user=nta password=nta_password dbname=nta port=5432 sslmode=disable
```

### Redis 连接失败

**解决方案**：
修改 `config/nta.yaml`：
```yaml
redis:
  addr: nta-redis:6379
```

### 配置文件未挂载

**问题**：容器内找不到配置文件

**解决方案**：
1. 确保配置文件存在：`ls config/nta.yaml`
2. 检查 docker-compose.yml 卷挂载：
```yaml
volumes:
  - ./config:/app/config:ro
```
3. 重启容器：`docker-compose restart nta-server`

## 服务管理

### 启动服务
```bash
docker-compose up -d
```

### 停止服务
```bash
docker-compose down
```

### 重启单个服务
```bash
docker-compose restart nta-server
```

### 查看服务状态
```bash
docker-compose ps
```

### 进入容器
```bash
docker exec -it nta-server sh
```

### 清理所有数据（谨慎操作）
```bash
docker-compose down -v  # 删除容器和数据卷
```

## 数据持久化

数据卷列表：
- `postgres-data`: PostgreSQL 数据
- `redis-data`: Redis 数据
- `nta-data`: NTA 应用数据
- `nta-logs`: 日志文件
- `nta-reports`: 报告文件
- `nta-pcap`: PCAP 文件
- `prometheus-data`: Prometheus 数据
- `grafana-data`: Grafana 数据

查看数据卷：
```bash
docker volume ls | grep nta
```

备份数据卷：
```bash
docker run --rm -v postgres-data:/data -v $(pwd):/backup alpine tar czf /backup/postgres-backup.tar.gz -C /data .
```

## 更新部署

### 更新镜像
```bash
# 1. 导入新镜像
docker load -i nta-server-v1.0.1.tar

# 2. 更新 docker-compose.yml 中的镜像版本
# 3. 重启服务
docker-compose up -d
```

### 更新配置
```bash
# 1. 修改 config/nta.yaml
# 2. 重启服务
docker-compose restart nta-server
```

## 日志管理

### 查看实时日志
```bash
docker-compose logs -f nta-server
```

### 查看最近 N 行日志
```bash
docker logs nta-server --tail 100
```

### 导出日志
```bash
docker logs nta-server > nta-server.log
```

## 监控

### Prometheus
访问 http://localhost:9090 查看指标

### Grafana
1. 访问 http://localhost:3000
2. 默认账号：admin/admin
3. 导入预配置的仪表板

## 安全建议

1. **修改默认密码**
   - 修改 PostgreSQL 密码
   - 修改 Grafana 管理员密码
   - 修改 JWT Secret

2. **使用防火墙**
   - 限制端口访问
   - 仅允许必要的外部访问

3. **定期备份**
   - 备份数据库
   - 备份配置文件
   - 备份重要数据卷

4. **更新镜像**
   - 定期更新基础镜像
   - 关注安全公告

## 联系支持

如有问题，请查看：
- GitHub Issues
- 项目文档
- 邮件：contact@qoder.com
