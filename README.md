# NTA 网络流量分析系统

## 项目简介

NTA (Network Traffic Analysis) 是一个基于 Zeek 的网络流量分析系统，提供实时流量监控、威胁检测、资产发现等功能。

## 系统架构

- **后端服务**: Go 语言开发的 API 服务器
- **前端界面**: Vue.js Web 管理界面
- **流量分析**: Zeek 网络监控引擎
- **数据存储**: PostgreSQL 数据库
- **消息队列**: Apache Kafka
- **缓存服务**: Redis

## 系统要求

### 硬件要求
- CPU: 4核及以上
- 内存: 8GB及以上
- 磁盘: 50GB可用空间

### 软件要求
- 操作系统: Ubuntu 20.04/22.04/24.04, CentOS 7/8, Rocky Linux 8/9
- 需要 root 权限

## 部署方式

### 离线安装（推荐）

1. 下载离线部署包
```bash
# 从 GitHub Releases 下载 nta-offline-deploy-*.tar.gz
wget https://github.com/Cxiyuan/NTA/releases/download/v2.0.0/nta-offline-deploy.tar.gz
```

2. 解压部署包
```bash
tar -xzf nta-offline-deploy-*.tar.gz
cd nta-offline-deploy
```

3. 运行安装脚本
```bash
sudo bash install.sh
```

4. 等待安装完成（首次安装约需 20-30 分钟）

### 在线安装

```bash
git clone https://github.com/Cxiyuan/NTA.git
cd NTA
sudo bash install.sh
```

## 访问系统

安装完成后，可以通过以下地址访问：

- **Web界面**: http://服务器IP:8090
- **API服务**: http://服务器IP:8080

默认账户：
- 用户名: `admin`
- 密码: `admin123`

⚠️ **首次登录后请立即修改默认密码！**

## 服务管理

### 查看服务状态
```bash
systemctl status nta-server
systemctl status nta-postgres
systemctl status nta-redis
systemctl status nta-kafka
systemctl status nta-zeek
```

### 启动/停止服务
```bash
# 启动所有服务
systemctl start nta-server nta-postgres nta-redis nta-kafka nta-zeek

# 停止所有服务
systemctl stop nta-server nta-postgres nta-redis nta-kafka nta-zeek

# 重启服务
systemctl restart nta-server
```

### 查看日志
```bash
# 查看服务日志
journalctl -u nta-server -f

# 查看应用日志
tail -f /var/log/nta/nta-server.log

# 查看 Zeek 日志
tail -f /var/lib/nta/zeek-logs/current/*.log
```

## 配置说明

### 主配置文件
配置文件位置: `/opt/nta/config/nta.yaml`

主要配置项：
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
  interface: eth0  # 监听网卡
```

修改配置后需要重启服务：
```bash
systemctl restart nta-server
```

### Zeek 探针配置

⚠️ **重要**: Zeek 探针需要在 Web 界面配置后启动

1. 登录 Web 界面
2. 进入 **系统管理** > **探针管理** > **内置探针**
3. 选择要监听的网卡接口
4. 配置 BPF 过滤规则（可选）
5. 点击启动探针

## 卸载系统

```bash
cd /opt/nta
sudo bash uninstall.sh
```

卸载时会询问是否删除数据文件，请根据需要选择。

## 目录结构

```
/opt/nta/              # 程序安装目录
├── bin/               # 可执行文件
├── config/            # 配置文件
└── web/               # Web前端文件

/var/lib/nta/          # 数据目录
├── postgres/          # PostgreSQL 数据
├── redis/             # Redis 数据
├── kafka/             # Kafka 数据
├── zeek-logs/         # Zeek 日志
└── pcap/              # PCAP 文件

/var/log/nta/          # 日志目录
├── nta/               # 应用日志
├── postgres/          # 数据库日志
├── redis/             # Redis 日志
├── kafka/             # Kafka 日志
└── zeek/              # Zeek 日志

/opt/postgres/         # PostgreSQL 程序
/opt/redis/            # Redis 程序
/opt/kafka/            # Kafka 程序
/opt/zeek/             # Zeek 程序
```

## 功能特性

- ✅ 实时流量监控
- ✅ 网络资产发现
- ✅ 威胁情报集成
- ✅ 异常行为检测
- ✅ 横向移动检测
- ✅ APT 攻击检测
- ✅ 报告生成
- ✅ 告警通知
- ✅ 用户权限管理
- ✅ 审计日志

## 常见问题

### 服务启动失败
```bash
# 查看详细错误日志
journalctl -u nta-server -n 100

# 检查端口占用
netstat -tlnp | grep 8080
```

### 数据库连接失败
```bash
# 检查 PostgreSQL 服务状态
systemctl status nta-postgres

# 测试数据库连接
/opt/postgres/bin/psql -U nta -d nta -h localhost
```

### Zeek 探针无法启动
```bash
# 检查 Zeek 配置
/opt/zeek/bin/zeekctl check

# 查看 Zeek 日志
tail -f /var/log/nta/zeek/zeek.log
```

## 技术支持

- 项目主页: https://github.com/Cxiyuan/NTA
- 问题反馈: https://github.com/Cxiyuan/NTA/issues

## 许可证

本项目采用 MIT 许可证
