# Cap Agent - Docker 部署指南

## 📦 快速部署

### 前提条件

- Docker 20.10+
- Docker Compose 2.0+
- 主机具有网络抓包权限
- 至少 8GB RAM 和 50GB 磁盘空间

### 1. 解压部署包

```bash
cd /root
tar -xzf cap-agent-latest.tar.gz
cd cap-agent-release
```

### 2. 配置网络接口

编辑 `docker-compose.yml`，修改环境变量：

```yaml
environment:
  - ZEEK_IFACE=eth0  # 修改为实际的网络接口名称
```

查看可用网络接口：

```bash
ip addr show
```

### 3. 创建必要的目录

```bash
mkdir -p logs reports config
```

### 4. 部署方式

#### 方式 1：一键部署（推荐）

启动所有服务（Zeek + Backend + Analyzer）：

```bash
docker-compose up -d
```

#### 方式 2：分步部署

```bash
# 仅启动 Zeek 流量分析
docker-compose up -d cap-agent

# 启动 Web 后端 API
docker-compose up -d cap-agent-backend

# 启动分析引擎
docker-compose up -d cap-agent-analyzer
```

#### 方式 3：使用 docker run

```bash
# 构建镜像
docker build -t cap-agent:latest .

# 运行 Zeek（需要 host 网络模式）
docker run -d \
  --name cap-agent \
  --privileged \
  --network host \
  -e ZEEK_IFACE=eth0 \
  -v $(pwd)/logs:/opt/cap-agent/logs \
  -v $(pwd)/reports:/opt/cap-agent/reports \
  -v $(pwd)/config:/opt/cap-agent/config \
  cap-agent:latest all
```

### 5. 验证部署

```bash
# 查看容器状态
docker-compose ps

# 查看 Zeek 运行状态
docker-compose exec cap-agent zeekctl status

# 查看日志
docker-compose logs -f cap-agent
docker-compose logs -f cap-agent-backend
docker-compose logs -f cap-agent-analyzer

# 测试 Backend API
curl http://localhost:5000/health
```

## 🔧 配置说明

### 环境变量

| 变量名 | 说明 | 默认值 |
|-------|------|--------|
| ZEEK_IFACE | 监听的网络接口 | eth0 |
| ZEEK_LOG_DIR | Zeek 日志目录 | /var/spool/zeek |
| PYTHONUNBUFFERED | Python 输出缓冲 | 1 |

### 端口映射

| 端口 | 服务 | 说明 |
|-----|------|------|
| 5000 | Backend API | Web 管理界面后端 |
| 5001 | Web UI | Web 管理界面（可选） |

### 数据卷

| 容器路径 | 主机路径 | 说明 |
|---------|---------|------|
| /opt/cap-agent/logs | ./logs | 应用日志 |
| /opt/cap-agent/reports | ./reports | 检测报告 |
| /opt/cap-agent/config | ./config | 配置文件 |
| /var/spool/zeek | /var/spool/zeek | Zeek 原始日志 |

## 🎯 服务说明

### cap-agent（主服务）

- 运行 Zeek 流量分析引擎
- 需要 `--privileged` 和 `--network host` 模式
- 监听指定网络接口的流量
- 生成结构化日志到 `/var/spool/zeek`

### cap-agent-backend

- 提供 REST API 接口
- 查询和管理检测结果
- Web UI 的后端服务
- 监听端口：5000

### cap-agent-analyzer

- 实时分析 Zeek 日志
- 执行机器学习检测
- 生成威胁报告
- 触发告警通知

## 📊 使用示例

### 查看实时告警

```bash
# 进入分析器容器
docker-compose exec cap-agent-analyzer bash

# 运行实时监控
python3 /opt/cap-agent/analyzer/integrated_engine.py --realtime
```

### 生成检测报告

```bash
docker-compose exec cap-agent-analyzer python3 \
  /opt/cap-agent/analyzer/integrated_engine.py \
  -i /var/spool/zeek/current/conn.log \
  -r /opt/cap-agent/reports/report-$(date +%Y%m%d).html
```

### 查看 Zeek 日志

```bash
# 连接日志
docker-compose exec cap-agent tail -f /var/spool/zeek/current/conn.log

# 横向移动日志
docker-compose exec cap-agent tail -f /var/spool/zeek/current/lateral_movement.log

# DNS 查询日志
docker-compose exec cap-agent tail -f /var/spool/zeek/current/dns.log
```

### 修改检测配置

```bash
# 编辑配置文件
vim config/detection.yaml

# 重启分析器使配置生效
docker-compose restart cap-agent-analyzer
```

## 🔄 运维管理

### 启动服务

```bash
docker-compose start
```

### 停止服务

```bash
docker-compose stop
```

### 重启服务

```bash
docker-compose restart
```

### 删除服务（保留数据）

```bash
docker-compose down
```

### 完全清理（删除数据）

```bash
docker-compose down -v
rm -rf logs reports
```

### 更新镜像

```bash
# 拉取新版本
tar -xzf cap-agent-latest-new.tar.gz

# 重新构建
docker-compose build --no-cache

# 重启服务
docker-compose up -d
```

### 查看资源占用

```bash
docker stats cap-agent cap-agent-backend cap-agent-analyzer
```

## 🐛 故障排查

### Zeek 无法启动

**问题**：`zeekctl deploy` 失败

**解决方案**：

1. 检查网络接口是否正确：
   ```bash
   docker-compose exec cap-agent ip addr show
   ```

2. 确认容器有 privileged 权限：
   ```bash
   docker inspect cap-agent | grep Privileged
   ```

3. 查看 Zeek 错误日志：
   ```bash
   docker-compose exec cap-agent cat /opt/zeek/logs/zeekctl.log
   ```

### 容器无法访问网络

**问题**：容器内无法抓包

**解决方案**：

1. 使用 host 网络模式（已配置）
2. 检查主机防火墙规则
3. 确认 SELinux/AppArmor 没有阻止

### Backend API 无响应

**问题**：`curl http://localhost:5000/health` 超时

**解决方案**：

1. 检查容器状态：
   ```bash
   docker-compose ps cap-agent-backend
   ```

2. 查看容器日志：
   ```bash
   docker-compose logs cap-agent-backend
   ```

3. 进入容器检查：
   ```bash
   docker-compose exec cap-agent-backend netstat -tlnp | grep 5000
   ```

### 磁盘空间不足

**问题**：Zeek 日志占用大量空间

**解决方案**：

1. 配置日志轮转：
   编辑 `/opt/zeek/etc/zeekctl.cfg`：
   ```
   LogRotationInterval = 3600    # 1小时轮转
   LogExpireInterval = 86400     # 24小时过期
   ```

2. 手动清理旧日志：
   ```bash
   docker-compose exec cap-agent find /var/spool/zeek -name "*.log.gz" -mtime +7 -delete
   ```

### Python 分析器崩溃

**问题**：analyzer 容器频繁重启

**解决方案**：

1. 增加内存限制：
   在 `docker-compose.yml` 中添加：
   ```yaml
   mem_limit: 8g
   ```

2. 查看崩溃日志：
   ```bash
   docker-compose logs --tail 100 cap-agent-analyzer
   ```

## 🔐 安全建议

1. **限制 Backend API 访问**：
   ```yaml
   ports:
     - "127.0.0.1:5000:5000"  # 仅本地访问
   ```

2. **使用非 root 用户运行**：
   在 Dockerfile 中添加：
   ```dockerfile
   RUN useradd -m capagent
   USER capagent
   ```

3. **加密日志传输**：
   配置 TLS/SSL 证书

4. **定期更新镜像**：
   及时应用安全补丁

## 📈 性能优化

### 资源限制

编辑 `docker-compose.yml`：

```yaml
services:
  cap-agent:
    deploy:
      resources:
        limits:
          cpus: '4'
          memory: 8G
        reservations:
          cpus: '2'
          memory: 4G
```

### 多实例部署

针对高流量环境，使用多个 Zeek 实例：

```yaml
services:
  cap-agent-1:
    <<: *cap-agent-common
    environment:
      - ZEEK_IFACE=eth0
  
  cap-agent-2:
    <<: *cap-agent-common
    environment:
      - ZEEK_IFACE=eth1
```

## 📞 技术支持

- 查看日志：`docker-compose logs -f`
- 问题反馈：GitHub Issues
- 文档：README.md

## 🔄 版本历史

- v2.0 (2025-12-23): 
  - ✅ 添加 Docker 支持
  - ✅ 添加 docker-compose 配置
  - ✅ 完善部署文档
  - ✅ 支持多容器分离部署

---

**部署成功后，访问 http://localhost:5000 查看管理界面。**
