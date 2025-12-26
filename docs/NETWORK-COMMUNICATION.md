# 服务间通信配置说明

## 网络架构

### Host 网络模式服务
- **nta-server**: `network_mode: host` (需要访问物理网卡抓包)
- **nta-zeek**: `network_mode: host` (需要直接访问网络接口)

### Bridge 网络模式服务
- 所有微服务: `networks: nta-network`
- 数据库/缓存: PostgreSQL, Redis
- 流处理平台: Kafka, Flink
- 基础设施: Traefik, Consul, Jaeger

## 通信方式

### 1. nta-server (host模式) 访问其他服务

nta-server 使用 `localhost` 访问已映射端口的服务：

```yaml
# 数据库连接
DB_HOST=localhost
DB_PORT=5432

# 微服务访问
http://localhost:8081  # auth-service
http://localhost:8082  # asset-service
http://localhost:8083  # detection-service
http://localhost:8084  # alert-service
http://localhost:8085  # report-service
http://localhost:8086  # notification-service
http://localhost:8087  # probe-service
http://localhost:8088  # intel-service
```

### 2. 微服务之间通信

微服务之间使用 Docker 服务名通信：

```yaml
# auth-service 访问 postgres
DB_HOST=postgres
DB_PORT=5432

# asset-service 调用 intel-service
http://intel-service:8088/api/v1/threat-intel/check
```

### 3. Traefik 路由

Traefik 通过服务名路由到微服务：

```yaml
# traefik-config.yml
auth-service:
  loadBalancer:
    servers:
      - url: "http://auth-service:8081"
```

### 4. Zeek 访问 Kafka

Zeek 使用 host 网络模式，访问 Kafka 外部端口：

```yaml
KAFKA_BROKERS=localhost:9093  # Kafka EXTERNAL listener
```

## 端口映射表

| 服务 | 容器端口 | 宿主机端口 | 用途 |
|------|----------|------------|------|
| postgres | 5432 | 5432 | nta-server 访问 |
| redis | 6379 | 6379 | nta-server 访问 |
| auth-service | 8081 | 8081 | nta-server 访问 |
| asset-service | 8082 | 8082 | nta-server 访问 |
| detection-service | 8083 | 8083 | nta-server 访问 |
| alert-service | 8084 | 8084 | nta-server 访问 |
| report-service | 8085 | 8085 | nta-server 访问 |
| notification-service | 8086 | 8086 | nta-server 访问 |
| probe-service | 8087 | 8087 | nta-server 访问 |
| intel-service | 8088 | 8088 | nta-server 访问 |
| traefik | 80 | 80 | Web 访问 |
| traefik-dashboard | 8080 | 8888 | 管理界面 |
| consul | 8500 | 8500 | 服务发现 |
| jaeger | 16686 | 16686 | 链路追踪 |
| kafka-internal | 9092 | 9092 | 容器内访问 |
| kafka-external | 9093 | 9093 | 宿主机访问 |
| flink | 8081 | 8081 | Flink Dashboard |

## 数据库初始化

PostgreSQL 容器启动时自动创建微服务数据库：

```bash
/docker-entrypoint-initdb.d/init-databases.sh
```

创建的数据库：
- auth_db
- asset_db
- alert_db
- report_db
- notify_db
- probe_db
- intel_db

## 配置验证

### 检查服务连通性

```bash
# 从 nta-server 容器测试
docker exec nta-server wget -O- http://localhost:8081/health  # auth-service
docker exec nta-server wget -O- http://localhost:8082/health  # asset-service

# 从微服务测试数据库
docker exec nta-auth-service nc -zv postgres 5432

# 测试 Traefik 路由
curl http://localhost/api/v1/auth/users
```

### 查看网络配置

```bash
# 查看 bridge 网络
docker network inspect nta_nta-network

# 查看服务 IP
docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' nta-auth-service
```

## 故障排查

### 问题: nta-server 无法连接微服务

**检查端口映射**:
```bash
docker-compose ps
netstat -tlnp | grep 808[1-8]
```

### 问题: 微服务无法连接数据库

**检查数据库是否创建**:
```bash
docker exec nta-postgres psql -U nta -c "\l"
```

### 问题: Traefik 路由 404

**检查 Traefik 配置**:
```bash
docker exec nta-traefik cat /etc/traefik/dynamic/traefik-config.yml
```

**查看 Traefik 日志**:
```bash
docker logs nta-traefik
```
