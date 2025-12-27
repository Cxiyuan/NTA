# NTA 部署检查清单

## 系统架构确认

### ✅ 已完成的重构
1. **移除 Docker 容器化部署** - 改为原生二进制部署
2. **移除微服务架构** - 所有功能集成到单体应用
3. **移除 Flink 流处理** - 简化为 Kafka 消息队列
4. **统一数据库** - 所有数据存储在单个 PostgreSQL 数据库

### ✅ 当前架构
```
NTA System (Native Binary Deployment)
├── nta-server (Port 8080)          # 主应用服务
├── nta-kafka-consumer              # Kafka 消费者
├── PostgreSQL (Port 5432)          # 数据库
├── Redis (Port 6379)               # 缓存
├── Kafka + Zookeeper (Port 9092/2181) # 消息队列
└── Zeek                            # 流量分析引擎
```

## 前后端通信检查

### ✅ 后端 API 路由 (server.go)
- `/api/v1/auth/*` - 认证接口
- `/api/v1/assets/*` - 资产管理
- `/api/v1/alerts/*` - 告警管理
- `/api/v1/threat-intel/*` - 威胁情报
- `/api/v1/probes/*` - 探针管理
- `/api/v1/builtin-probe/*` - 内置探针
- `/api/v1/reports/*` - 报告生成
- `/api/v1/notifications/*` - 通知配置
- `/api/v1/pcap/*` - PCAP 管理
- `/api/v1/detection/*` - 检测引擎
- `/api/v1/users/*` - 用户管理
- `/api/v1/roles/*` - 角色管理
- `/api/v1/tenants/*` - 租户管理
- `/api/v1/config/*` - 系统配置
- `/api/v1/stream/*` - Kafka 状态
- `/api/v1/license` - 许可证管理

### ✅ 前端 API 调用 (api.ts)
- `authAPI` - 登录/登出/当前用户
- `alertAPI` - 告警列表/详情/更新
- `assetAPI` - 资产列表/详情
- `threatIntelAPI` - 威胁情报查询/更新
- `probeAPI` - 探针注册/心跳
- `reportAPI` - 报告列表/生成/下载
- `notificationAPI` - 通知配置
- `pcapAPI` - PCAP 查询/下载
- `builtinProbeAPI` - 内置探针管理

### ✅ API 通信配置
- 前端 baseURL: `/api/v1` (代理到 http://localhost:8080)
- 生产环境: `VITE_API_BASE_URL=http://localhost:8080/api/v1`
- 认证方式: JWT Bearer Token
- 超时时间: 10秒

## 配置文件检查

### ✅ 后端配置 (config/nta.yaml)
```yaml
server:
  host: 0.0.0.0
  port: 8080

database:
  dsn: host=localhost user=nta password=nta_password dbname=nta port=5432

redis:
  addr: localhost:6379

zeek:
  log_dir: /var/lib/nta/zeek-logs
  interface: eth0
```

### ✅ 前端配置 (web/.env.production)
```
VITE_API_BASE_URL=http://localhost:8080/api/v1
```

## 部署脚本检查

### ✅ 安装脚本 (install.sh)
- [x] 仅支持 Ubuntu 24.04
- [x] 安装系统依赖
- [x] 编译安装 PostgreSQL
- [x] 编译安装 Redis
- [x] 安装 Kafka + Zookeeper
- [x] 编译安装 Zeek
- [x] 部署 NTA 应用
- [x] 创建 systemd 服务
- [x] 初始化数据库
- [x] 配置防火墙

### ✅ GitHub Actions (build-offline-package.yml)
- [x] 编译 Go 二进制文件
- [x] 构建 Vue.js 前端
- [x] 下载依赖包 (PostgreSQL, Redis, Kafka, Zeek)
- [x] 打包离线部署包
- [x] 生成版本信息和 README

## 代码清理检查

### ✅ 已删除的文件/目录
- [x] `services/` - 微服务目录
- [x] `docker/` - Docker 配置
- [x] `docker-compose.yml` - Docker Compose 配置
- [x] `Dockerfile` - 所有 Dockerfile
- [x] `flink-jobs/` - Flink 作业
- [x] `pkg/client/microservice_client.go` - 微服务客户端
- [x] `scripts/verify-communication.sh` - 微服务验证脚本
- [x] `scripts/update-workflow.sh` - 过时的工作流脚本
- [x] `deploy/traefik-config.yml` - Traefik 配置
- [x] `deploy/DOCKER_DEPLOY.md` - Docker 部署文档
- [x] `deploy/kubernetes.yaml` - K8s 配置

### ✅ 已更新的文件
- [x] `cmd/nta-server/main.go` - 路径更新为 /opt/nta 和 /var/lib/nta
- [x] `internal/zeek/manager.go` - Docker 命令改为 systemctl
- [x] `internal/kafka/manager.go` - 移除 Flink 集成
- [x] `internal/api/stream_handlers.go` - 移除 Flink 路由
- [x] `config/nta.yaml` - 更新为本地路径
- [x] `web/.env.production` - 更新 API 地址

## 潜在问题排查

### ⚠️ 需要注意的点

1. **端口冲突**
   - 8080 (nta-server)
   - 5432 (PostgreSQL)
   - 6379 (Redis)
   - 9092 (Kafka)
   - 2181 (Zookeeper)

2. **文件权限**
   - /opt/nta - nta 用户
   - /var/lib/nta - nta 用户
   - /opt/zeek - root 用户

3. **Zeek 配置**
   - 需要在 Web 界面手动配置监听网卡
   - 不会自动启动，需用户主动启动

4. **数据库连接**
   - 确保 PostgreSQL 已启动
   - 确保 nta 数据库已创建
   - 确保 nta 用户权限正确

5. **Kafka 依赖**
   - Zookeeper 必须先于 Kafka 启动
   - Kafka 启动需要 10-15 秒

## 部署流程

1. 下载离线部署包
2. 解压到目标目录
3. 运行 `sudo bash install.sh`
4. 等待安装完成（约 30-60 分钟）
5. 访问 http://服务器IP:8090
6. 使用 admin/admin123 登录
7. 配置 Zeek 监听网卡
8. 启动内置探针

## 验证步骤

```bash
# 1. 检查所有服务状态
systemctl status nta-postgres
systemctl status nta-redis
systemctl status nta-zookeeper
systemctl status nta-kafka
systemctl status nta-kafka-consumer
systemctl status nta-server

# 2. 检查端口监听
ss -tlnp | grep -E "8080|5432|6379|9092|2181"

# 3. 测试 API
curl http://localhost:8080/health
curl http://localhost:8080/api/v1/assets

# 4. 查看日志
journalctl -u nta-server -f
tail -f /var/log/nta/nta/nta-server.log
```

## 总结

✅ **所有重构任务已完成**
✅ **前后端 API 通信已验证**
✅ **配置文件已更新**
✅ **部署脚本已完善**
✅ **代码清理已完成**
✅ **系统架构简化完成**

⚠️ **注意事项**：
- 仅支持 Ubuntu 24.04 LTS
- 首次安装需要编译 Zeek，耗时较长
- Zeek 探针需要手动配置后启动
- 默认密码登录后请及时修改
