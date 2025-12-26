# NTA 流处理架构部署文档

## 📚 目录

- [系统架构](#系统架构)
- [环境要求](#环境要求)
- [快速开始](#快速开始)
- [组件说明](#组件说明)
- [配置说明](#配置说明)
- [运维指南](#运维指南)
- [故障排查](#故障排查)

---

## 系统架构

### 整体架构图

```
┌─────────────────────────────────────────────────────────────┐
│                      NTA 流处理架构 v2.0                      │
└─────────────────────────────────────────────────────────────┘

流量采集层:
  ┌──────────┐
  │   Zeek   │──► Kafka (zeek-conn/dns/http/ssl/notice)
  └──────────┘

消息队列层:
  ┌──────────────┐    ┌──────────────┐
  │  Zookeeper   │───►│    Kafka     │
  │   (协调)      │    │  (8分区)      │
  └──────────────┘    └──────┬───────┘
                             │
          ┌──────────────────┼──────────────────┐
          ↓                  ↓                  ↓
流处理层: Flink          Flink              Flink
     (C2检测)        (DGA检测)         (数据渗出)
          ↓                  ↓                  ↓
          └──────────────────┴──────────────────┘
                             │
消费层:                      ↓
  ┌──────────────────────────────┐
  │   Kafka Consumer (Go)        │
  │  - 实时威胁检测               │
  │  - 横向移动分析               │
  │  - 数据入库                   │
  └──────────┬───────────────────┘
             │
存储层:      ↓
  ┌──────────────┐    ┌──────────┐
  │ PostgreSQL   │    │  Redis   │
  └──────────────┘    └──────────┘

展示层:
  ┌──────────────┐    ┌──────────┐
  │   NTA Web    │    │ Grafana  │
  └──────────────┘    └──────────┘
```

---

## 环境要求

### 硬件要求

| 环境 | CPU | 内存 | 磁盘 | 网络 |
|------|-----|------|------|------|
| **测试环境** | 4核 | 8GB | 100GB | 1Gbps |
| **生产环境** | 8核+ | 16GB+ | 500GB+ | 10Gbps |
| **高性能** | 16核+ | 32GB+ | 1TB+ | 10Gbps+ |

### 软件要求

- **操作系统**: CentOS 7+, Ubuntu 20.04+, Anolis OS 8
- **内核**: Linux 3.10+
- **Docker**: 20.10+
- **Docker Compose**: 2.0+

---

## 快速开始

### 1. 下载离线安装包

```bash
# 解压安装包
unzip nta-deploy-v2.0.0.zip
cd nta-deploy-v2.0.0
```

### 2. 执行安装

```bash
# 赋予执行权限
chmod +x install.sh

# 运行安装脚本
sudo ./install.sh
```

### 3. 验证部署

```bash
# 检查所有容器状态
docker-compose ps

# 应该看到以下容器运行中:
# - nta-postgres
# - nta-redis
# - nta-zookeeper
# - nta-kafka
# - nta-flink-jobmanager
# - nta-flink-taskmanager
# - nta-kafka-consumer
# - nta-server
# - nta-web
# - nta-zeek
# - nta-prometheus
# - nta-grafana
```

### 4. 访问系统

- **Web界面**: http://YOUR_SERVER_IP
- **Flink Web UI**: http://YOUR_SERVER_IP:8081
- **Grafana**: http://YOUR_SERVER_IP:3000

---

## 组件说明

### Zookeeper

**作用**: Kafka集群协调器

**配置**:
- 端口: 2181
- 数据目录: `/bitnami/zookeeper`

### Kafka

**作用**: 高吞吐消息队列

**Topic列表**:
- `zeek-conn`: 网络连接日志 (8分区)
- `zeek-dns`: DNS查询日志 (8分区)
- `zeek-http`: HTTP流量日志 (8分区)
- `zeek-ssl`: SSL/TLS日志 (8分区)
- `zeek-notice`: Zeek告警日志 (8分区)

**配置**:
- 端口: 9092 (内部), 9093 (外部)
- 保留时间: 7天
- 默认副本: 1

**监控命令**:
```bash
# 列出所有Topic
docker exec nta-kafka kafka-topics.sh \
  --list --bootstrap-server localhost:9092

# 查看Topic详情
docker exec nta-kafka kafka-topics.sh \
  --describe --topic zeek-conn \
  --bootstrap-server localhost:9092

# 查看消费组
docker exec nta-kafka kafka-consumer-groups.sh \
  --list --bootstrap-server localhost:9092

# 查看消费积压
docker exec nta-kafka kafka-consumer-groups.sh \
  --describe --group nta-consumer-group \
  --bootstrap-server localhost:9092
```

### Flink

**作用**: 实时流处理引擎

**已部署作业**:
1. **C2 Beacon检测**: 10分钟滑动窗口检测规律性信标通信
2. **DGA域名检测**: 实时检测算法生成的恶意域名
3. **数据渗出检测**: 5分钟窗口检测异常上传流量

**配置**:
- JobManager端口: 8081
- TaskManager Slots: 4
- 检查点目录: `/opt/flink/checkpoints`

**管理命令**:
```bash
# 查看运行中的作业
curl http://localhost:8081/jobs

# 查看作业详情
curl http://localhost:8081/jobs/<JOB_ID>

# 取消作业
curl -X PATCH http://localhost:8081/jobs/<JOB_ID>

# 重新部署作业
bash flink-jobs/deploy-jobs.sh
```

### Kafka Consumer (Go)

**作用**: 消费Kafka消息并执行威胁检测

**检测功能**:
- C2通信检测
- WebShell检测
- 数据渗出检测
- 横向移动检测

**日志查看**:
```bash
docker logs -f nta-kafka-consumer
```

---

## 配置说明

### Kafka配置优化

编辑 `docker-compose.yml`:

```yaml
kafka:
  environment:
    # 增加分区数提升并发
    - KAFKA_CFG_NUM_PARTITIONS=16
    
    # 延长保留时间
    - KAFKA_CFG_LOG_RETENTION_HOURS=336  # 14天
    
    # 增大保留大小
    - KAFKA_CFG_LOG_RETENTION_BYTES=21474836480  # 20GB
```

### Flink资源配置

编辑 `docker-compose.yml`:

```yaml
flink-taskmanager:
  environment:
    # 增加Task Slots
    - TASK_MANAGER_NUMBER_OF_TASK_SLOTS=8
    
    # 增加内存
    deploy:
      resources:
        limits:
          memory: 4G
```

### Zeek Kafka输出配置

编辑 `zeek-scripts/kafka-output.zeek`:

```zeek
# 修改Topic前缀
const topic_prefix = "nta-prod" &redef;

# 禁用某些日志
@if (! enable_ssl_logging)
    Log::disable_stream(SSL::LOG);
@endif
```

---

## 运维指南

### 日常监控

**1. 检查Kafka积压**
```bash
# 查看所有消费组积压
for group in $(docker exec nta-kafka kafka-consumer-groups.sh \
  --list --bootstrap-server localhost:9092); do
  echo "==> $group"
  docker exec nta-kafka kafka-consumer-groups.sh \
    --describe --group $group \
    --bootstrap-server localhost:9092 | grep -E "LAG|CONSUMER-ID"
done
```

**2. 监控Flink作业**
```bash
# 访问Flink Web UI
open http://localhost:8081

# 或使用API
curl http://localhost:8081/jobs/overview
```

**3. 查看系统资源**
```bash
docker stats nta-kafka nta-flink-jobmanager nta-flink-taskmanager
```

### 扩容指南

**水平扩展Kafka Consumer**:
```bash
# 修改docker-compose.yml
kafka-consumer:
  deploy:
    replicas: 3  # 增加副本数
```

**增加Flink TaskManager**:
```bash
docker-compose up -d --scale flink-taskmanager=3
```

### 备份策略

**1. Kafka数据备份**
```bash
# 导出Topic数据
docker exec nta-kafka kafka-console-consumer.sh \
  --bootstrap-server localhost:9092 \
  --topic zeek-conn \
  --from-beginning \
  --max-messages 100000 > backup-conn.json
```

**2. Flink状态备份**
```bash
# Flink自动Checkpoint，保存在 /opt/flink/checkpoints
tar -czf flink-checkpoints-$(date +%Y%m%d).tar.gz \
  -C /var/lib/docker/volumes/nta_flink-checkpoints flink/checkpoints
```

---

## 故障排查

### Kafka无法启动

**症状**: `docker logs nta-kafka` 显示连接Zookeeper失败

**解决**:
```bash
# 1. 检查Zookeeper是否运行
docker ps | grep zookeeper

# 2. 重启Zookeeper
docker-compose restart zookeeper

# 3. 等待30秒后重启Kafka
sleep 30
docker-compose restart kafka
```

### Flink作业失败

**症状**: Flink Web UI显示作业状态为 FAILED

**解决**:
```bash
# 1. 查看JobManager日志
docker logs nta-flink-jobmanager

# 2. 查看TaskManager日志
docker logs nta-flink-taskmanager

# 3. 重新提交作业
bash flink-jobs/deploy-jobs.sh
```

### 消费积压过大

**症状**: Kafka Consumer Lag > 10000

**解决**:
```bash
# 1. 检查Consumer是否在运行
docker ps | grep kafka-consumer

# 2. 查看Consumer日志
docker logs -f nta-kafka-consumer --tail 100

# 3. 增加Consumer副本
docker-compose up -d --scale kafka-consumer=3

# 4. 临时增加Consumer处理速度(重启跳过旧消息)
docker-compose restart kafka-consumer
```

### Zeek未发送数据到Kafka

**症状**: Kafka Topic中无消息

**解决**:
```bash
# 1. 检查Zeek是否运行
docker exec nta-zeek zeekctl status

# 2. 检查Kafka插件是否加载
docker exec nta-zeek zeek -e 'print Kafka::kafka_conf;'

# 3. 查看Zeek日志
docker logs nta-zeek

# 4. 重启Zeek
docker-compose restart nta-zeek
```

---

## 性能调优

### 针对高流量场景 (>1Gbps)

**1. Kafka优化**
```yaml
kafka:
  environment:
    # 增加网络线程
    - KAFKA_CFG_NUM_NETWORK_THREADS=8
    # 增加IO线程
    - KAFKA_CFG_NUM_IO_THREADS=8
    # 增大批量大小
    - KAFKA_CFG_SOCKET_SEND_BUFFER_BYTES=1048576
    - KAFKA_CFG_SOCKET_RECEIVE_BUFFER_BYTES=1048576
```

**2. Zeek优化**
```bash
# 使用多个Zeek进程
docker exec nta-zeek zeekctl deploy --workers=4
```

**3. Flink优化**
```yaml
flink-taskmanager:
  environment:
    # 增加并行度
    - FLINK_PROPERTIES=parallelism.default: 8
  deploy:
    replicas: 2  # 多TaskManager
```

---

## 附录

### 端口列表

| 组件 | 端口 | 说明 |
|------|------|------|
| Zookeeper | 2181 | 协调服务 |
| Kafka | 9092 | 内部通信 |
| Kafka | 9093 | 外部访问 |
| Flink JobManager | 8081 | Web UI |
| NTA Server | 8080 | API服务 |
| NTA Web | 80 | Web界面 |
| Prometheus | 9090 | 监控指标 |
| Grafana | 3000 | 可视化 |

### 常见问题

**Q: 离线包有多大?**
A: 约3.5GB (包含所有镜像)

**Q: 支持集群部署吗?**
A: 当前版本为单机版，集群版需要修改Kafka/Flink为分布式配置

**Q: 数据保留多久?**
A: Kafka默认7天，PostgreSQL根据配置清理

**Q: 如何升级?**
A: 下载新版离线包，备份数据后重新部署

---

📧 技术支持: support@nta.com
📚 在线文档: https://docs.nta.com
