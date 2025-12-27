# NTA 安装组件清单

## 1. 系统依赖包 (apt-get install)

### 编译工具
- `build-essential` - 基础编译工具集 (gcc, g++, make等)
- `cmake` - CMake 构建工具
- `gcc` - GNU C 编译器
- `g++` - GNU C++ 编译器

### 开发库
- `libpcap-dev` - 网络数据包捕获库 (Zeek需要)
- `libssl-dev` - OpenSSL 开发库 (PostgreSQL, Redis需要)
- `zlib1g-dev` - 压缩库 (PostgreSQL, Redis需要)
- `libreadline-dev` - 命令行编辑库 (PostgreSQL需要)
- `libncurses5-dev` - 终端控制库
- `libmaxminddb-dev` - GeoIP 数据库库 (Zeek需要)
- `libkrb5-dev` - Kerberos 开发库 (Zeek需要)

### Zeek 专用依赖
- `flex` - 词法分析器
- `bison` - 语法分析器
- `swig` - 接口生成工具

### Python
- `python3` - Python 3 解释器
- `python3-pip` - Python 包管理器

### 系统工具
- `curl` - HTTP 客户端
- `wget` - 文件下载工具
- `tar` - 归档工具
- `gzip` - 压缩工具
- `bzip2` - 压缩工具
- `net-tools` - 网络工具 (netstat等)
- `tcpdump` - 网络抓包工具
- `iproute2` - IP 路由工具
- `systemd` - 系统服务管理
- `vim` - 文本编辑器
- `nano` - 文本编辑器

## 2. 编译安装的组件

### PostgreSQL 15.5
- **源码路径**: `packages/postgresql-15.5.tar.gz`
- **安装位置**: `/opt/postgres`
- **数据目录**: `/var/lib/nta/postgres`
- **配置选项**: `./configure --prefix=/opt/postgres --with-openssl`
- **编译时间**: 约 5-10 分钟
- **磁盘占用**: ~200MB

### Redis 7.2.3
- **源码路径**: `packages/redis-7.2.3.tar.gz`
- **安装位置**: `/opt/redis`
- **数据目录**: `/var/lib/nta/redis`
- **编译命令**: `make -j$(nproc)`
- **编译时间**: 约 2-5 分钟
- **磁盘占用**: ~50MB

### Kafka 3.6.1 (含 Zookeeper)
- **二进制包**: `packages/kafka_2.13-3.6.1.tgz`
- **安装位置**: `/opt/kafka`
- **数据目录**: `/var/lib/nta/kafka`
- **编译时间**: 无需编译（Java 应用）
- **磁盘占用**: ~100MB
- **依赖**: Java Runtime (已包含在系统中)

### Zeek 6.0.3
- **源码路径**: `packages/zeek-6.0.3.tar.gz`
- **安装位置**: `/opt/zeek`
- **数据目录**: `/var/lib/nta/zeek-logs`
- **配置选项**: `./configure --prefix=/opt/zeek`
- **编译时间**: 约 15-30 分钟 ⚠️ (最耗时)
- **磁盘占用**: ~500MB

## 3. NTA 应用程序

### nta-server
- **二进制文件**: `bin/nta-server`
- **安装位置**: `/opt/nta/bin/nta-server`
- **语言**: Go 1.23
- **编译环境**: Ubuntu 24.04
- **磁盘占用**: ~30MB

### nta-kafka-consumer
- **二进制文件**: `bin/nta-kafka-consumer`
- **安装位置**: `/opt/nta/bin/nta-kafka-consumer`
- **语言**: Go 1.23
- **编译环境**: Ubuntu 24.04
- **磁盘占用**: ~20MB

### Web 前端
- **静态文件**: `web/*`
- **安装位置**: `/opt/nta/web/`
- **框架**: Vue.js
- **磁盘占用**: ~10MB

## 4. 配置文件

### 应用配置
- `config/nta.yaml` → `/opt/nta/config/nta.yaml`
- `config/threat_feed.json` → `/opt/nta/config/` (可选)
- `config/license.key` → `/opt/nta/config/` (可选)
- `config/public.pem` → `/opt/nta/config/` (可选)

### Zeek 脚本
- `zeek-scripts/*` → `/opt/zeek/share/zeek/site/`

### 部署脚本
- `scripts/init-databases.sh` → `/opt/nta/scripts/`

## 5. Systemd 服务

安装脚本会创建以下 systemd 服务：

1. `nta-postgres.service` - PostgreSQL 数据库
2. `nta-redis.service` - Redis 缓存
3. `nta-zookeeper.service` - Zookeeper 服务
4. `nta-kafka.service` - Kafka 消息队列
5. `nta-zeek.service` - Zeek 流量分析 (手动启动)
6. `nta-kafka-consumer.service` - Kafka 消费者
7. `nta-server.service` - NTA 主服务

## 6. 目录结构

```
/opt/nta/                    # NTA 应用程序
├── bin/
│   ├── nta-server
│   └── nta-kafka-consumer
├── config/
│   └── nta.yaml
└── web/

/opt/postgres/               # PostgreSQL
├── bin/
└── lib/

/opt/redis/                  # Redis
├── bin/
└── etc/

/opt/kafka/                  # Kafka + Zookeeper
├── bin/
├── config/
└── libs/

/opt/zeek/                   # Zeek
├── bin/
├── etc/
└── share/

/var/lib/nta/                # 数据目录
├── postgres/
├── redis/
├── kafka/
├── zeek-logs/
├── pcap/
└── backups/

/var/log/nta/                # 日志目录
├── nta/
├── postgres/
├── redis/
├── kafka/
└── zeek/
```

## 7. 安装时间估算

| 步骤 | 预计时间 |
|------|----------|
| 系统依赖安装 | 2-5 分钟 |
| PostgreSQL 编译 | 5-10 分钟 |
| Redis 编译 | 2-5 分钟 |
| Kafka 解压配置 | 1 分钟 |
| **Zeek 编译** | **15-30 分钟** ⚠️ |
| NTA 应用部署 | 2 分钟 |
| 服务配置启动 | 2-3 分钟 |
| **总计** | **约 30-60 分钟** |

⚠️ Zeek 编译时间取决于 CPU 性能，是最耗时的步骤

## 8. 磁盘空间需求

| 组件 | 磁盘占用 |
|------|----------|
| 系统依赖 | ~500MB |
| PostgreSQL | ~200MB |
| Redis | ~50MB |
| Kafka | ~100MB |
| Zeek | ~500MB |
| NTA 应用 | ~60MB |
| 日志和数据 (初始) | ~100MB |
| **临时编译文件** | ~1GB |
| **总计** | **约 2.5GB** |

建议至少预留 **50GB** 可用空间用于日志和数据增长。

## 9. 内存需求

### 运行时内存占用
- PostgreSQL: ~256MB
- Redis: ~50MB
- Zookeeper: ~100MB
- Kafka: ~512MB
- Zeek: ~200MB (取决于流量)
- nta-server: ~100MB
- nta-kafka-consumer: ~50MB
- **总计**: ~1.3GB

建议系统至少有 **8GB** 内存。

## 10. 网络端口

| 服务 | 端口 | 说明 |
|------|------|------|
| nta-server | 8080 | API 服务 |
| Web UI | 8090 | Web 界面 |
| PostgreSQL | 5432 | 数据库 (仅本地) |
| Redis | 6379 | 缓存 (仅本地) |
| Kafka | 9092 | 消息队列 (仅本地) |
| Zookeeper | 2181 | 协调服务 (仅本地) |

## 11. 安装前检查清单

- [ ] 系统是 Ubuntu 24.04 LTS
- [ ] 有 root 权限
- [ ] 磁盘可用空间 > 50GB
- [ ] 内存 > 8GB
- [ ] CPU > 4 核
- [ ] 端口 8080, 8090 未被占用
- [ ] 可以访问互联网（下载系统依赖）

## 12. 安装命令

```bash
# 1. 解压部署包
tar -xzf nta-offline-deploy-*.tar.gz
cd nta-offline-deploy-*

# 2. 运行安装脚本
sudo bash install.sh

# 3. 等待安装完成
# 首次安装约需 30-60 分钟
# 期间会显示各组件的安装进度
```

## 13. 验证安装

```bash
# 检查服务状态
systemctl status nta-postgres
systemctl status nta-redis
systemctl status nta-kafka
systemctl status nta-server

# 测试 API
curl http://localhost:8080/health

# 访问 Web 界面
# http://服务器IP:8090
```
