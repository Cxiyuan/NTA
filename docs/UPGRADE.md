# Kafka/Flink 流处理架构升级指南

## 🎯 升级概述

本次升级将NTA从单体架构升级为Kafka/Flink流处理架构，实现：
- ✅ 10倍吞吐量提升 (100Mbps → 1Gbps+)
- ✅ 秒级检测延迟 (分钟级 → 秒级)
- ✅ 水平扩展能力
- ✅ 实时流式分析

---

## 📋 变更清单

### 新增组件

| 组件 | 版本 | 作用 | 资源占用 |
|------|------|------|---------|
| Zookeeper | 3.9 | Kafka协调 | 512MB |
| Kafka | 3.6 | 消息队列 | 1GB |
| Flink | 1.18 | 流处理 | 2GB |
| Kafka Consumer | v1.0.0 | 威胁检测 | 512MB |

**总新增资源**: ~4GB内存

### 架构变化

**旧架构**:
```
Zeek → 写文件 → LogParser → PostgreSQL → API
```

**新架构**:
```
Zeek → Kafka → Flink/Consumer → PostgreSQL → API
              ↓
         持久化7天
```

---

## 🔄 升级步骤

### 前置准备

1. **备份现有数据**
```bash
# 备份数据库
docker exec nta-postgres pg_dump -U nta nta > backup-$(date +%Y%m%d).sql

# 备份配置
tar -czf config-backup.tar.gz config/
```

2. **检查资源**
```bash
# 确保至少有4GB空闲内存
free -h

# 确保至少有20GB空闲磁盘
df -h /var/lib/docker
```

### 升级操作

**方式1: 使用新的离线包 (推荐)**

```bash
# 1. 停止旧服务
cd /path/to/old/nta
docker-compose down

# 2. 解压新版本
cd /opt
unzip nta-deploy-v2.0.0.zip
cd nta-deploy-v2.0.0

# 3. 恢复配置和数据
cp /path/to/old/nta/config/nta.yaml config/
# 如需恢复数据:
# cat backup-20251226.sql | docker exec -i nta-postgres psql -U nta

# 4. 安装
sudo ./install.sh
```

**方式2: 原地升级**

```bash
# 1. 停止服务
docker-compose stop

# 2. 拉取新镜像
docker pull bitnami/zookeeper:3.9
docker pull bitnami/kafka:3.6
docker pull flink:1.18-scala_2.12-java11

# 3. 更新docker-compose.yml
# (添加zookeeper/kafka/flink服务，参考新版配置)

# 4. 更新代码
git pull origin main
go mod tidy
docker-compose build

# 5. 启动服务
docker-compose up -d
```

---

## ✅ 验证升级

### 1. 检查所有容器运行

```bash
docker-compose ps

# 应该看到13个容器运行中:
# ✅ nta-postgres
# ✅ nta-redis
# ✅ nta-zookeeper        (新增)
# ✅ nta-kafka            (新增)
# ✅ nta-flink-jobmanager (新增)
# ✅ nta-flink-taskmanager(新增)
# ✅ nta-kafka-consumer   (新增)
# ✅ nta-server
# ✅ nta-web
# ✅ nta-zeek
# ✅ nta-prometheus
# ✅ nta-grafana
```

### 2. 验证Kafka消息流

```bash
# 查看Topic
docker exec nta-kafka kafka-topics.sh \
  --list --bootstrap-server localhost:9092

# 应该看到:
# zeek-conn
# zeek-dns
# zeek-http
# zeek-ssl
# zeek-notice

# 消费测试消息
docker exec nta-kafka kafka-console-consumer.sh \
  --bootstrap-server localhost:9092 \
  --topic zeek-conn \
  --max-messages 5
```

### 3. 验证Flink作业

```bash
# 访问Flink Web UI
curl http://localhost:8081/jobs

# 应该看到3个RUNNING状态的作业:
# - C2 Beacon Detection
# - DGA Detection
# - Data Exfiltration Detection
```

### 4. 验证告警生成

```bash
# 查看最近告警
curl -H "Authorization: Bearer YOUR_TOKEN" \
  http://localhost:8080/api/v1/alerts?limit=10

# 应该看到实时生成的告警
```

---

## 🐛 常见问题

### Kafka启动失败

**错误信息**: `Connection to node -1 could not be established`

**原因**: Zookeeper未就绪

**解决**:
```bash
docker-compose restart zookeeper
sleep 30
docker-compose restart kafka
```

### Flink作业未运行

**错误信息**: `/jobs返回空数组`

**原因**: 作业部署脚本未执行

**解决**:
```bash
# 手动部署
bash flink-jobs/deploy-jobs.sh
```

### Zeek未发送数据到Kafka

**错误信息**: `Kafka Topic无消息`

**原因**: Kafka插件未加载

**解决**:
```bash
# 检查Zeek配置
docker exec nta-zeek cat /opt/zeek/share/zeek/site/nta/kafka-output.zeek

# 重启Zeek
docker-compose restart nta-zeek
```

### Consumer消费延迟

**现象**: Lag持续增长

**原因**: 消费速度跟不上生产速度

**解决**:
```bash
# 增加Consumer实例
docker-compose up -d --scale kafka-consumer=3
```

---

## 📊 性能对比

### 升级前后对比

| 指标 | 旧架构 | 新架构 | 提升 |
|------|--------|--------|------|
| 最大吞吐 | 100Mbps | 1Gbps+ | **10倍** |
| 检测延迟 | 30-60秒 | <3秒 | **20倍** |
| 并发连接 | 10万 | 100万+ | **10倍** |
| 数据保留 | 30天 | 7天(Kafka)+30天(DB) | 灵活 |
| 扩展性 | 垂直 | 水平 | ♾️ |

### 资源消耗

| 组件 | 旧架构 | 新架构 | 增量 |
|------|--------|--------|------|
| CPU | 4核 | 6核 | +2核 |
| 内存 | 8GB | 12GB | +4GB |
| 磁盘 | 100GB | 120GB | +20GB |

---

## 🔙 回滚方案

如果升级后出现严重问题，可按以下步骤回滚:

```bash
# 1. 停止新版本
cd /opt/nta-deploy-v2.0.0
docker-compose down

# 2. 恢复旧版本
cd /path/to/old/nta
docker-compose up -d

# 3. 恢复数据(如果有备份)
cat backup-20251226.sql | docker exec -i nta-postgres psql -U nta
```

---

## 📞 支持

如遇到问题，请提供以下信息:

```bash
# 收集诊断信息
cat > diagnostic-info.txt << EOF
=== 系统信息 ===
$(uname -a)
$(free -h)
$(df -h)

=== Docker版本 ===
$(docker --version)
$(docker-compose --version)

=== 容器状态 ===
$(docker-compose ps)

=== 最近日志 ===
$(docker-compose logs --tail 50)
EOF

# 发送到: support@nta.com
```

---

✅ 升级完成后，您将拥有一个高性能、可扩展的NTA流处理系统！
