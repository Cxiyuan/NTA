# NTA 网络流量分析探针 - 部署指南

## 系统要求

### 硬件要求
- **CPU**: 4核心及以上 (推荐8核)
- **内存**: 4GB及以上 (推荐8GB，Go版本内存需求大幅降低)
- **磁盘**: 100GB及以上可用空间
- **架构**: x86_64 (Linux 64位)

### 支持的操作系统
- **Ubuntu 24.04 LTS** (仅支持此版本)

## 快速安装

### 1. 下载项目

```bash
git clone https://github.com/Cxiyuan/NTA.git
cd NTA
```

### 2. 运行安装脚本

```bash
sudo bash deploy/install.sh
```

安装脚本会自动完成：
- ✅ 检测操作系统和硬件环境
- ✅ 安装系统依赖包
- ✅ 编译安装 Zeek 6.0.3
- ✅ 安装 Python 依赖
- ✅ 创建服务用户
- ✅ 配置 Zeek 监听网络接口
- ✅ 创建 systemd 服务
- ✅ 启动所有服务

### 3. 验证安装

```bash
# 检查服务状态
systemctl status nta-zeek
systemctl status nta-backend
systemctl status nta-analyzer
systemctl status nta-probe-manager

# 检查 Zeek 运行状态
zeekctl status

# 访问 API
curl http://localhost:5000/health
```

## 服务管理

### 启动服务
```bash
systemctl start nta-zeek nta-backend nta-analyzer nta-probe-manager
```

### 停止服务
```bash
systemctl stop nta-zeek nta-backend nta-analyzer nta-probe-manager
```

### 重启服务
```bash
systemctl restart nta-zeek nta-backend nta-analyzer nta-probe-manager
```

### 查看服务状态
```bash
systemctl status nta-zeek nta-backend nta-analyzer nta-probe-manager
```

### 开机自启
```bash
systemctl enable nta-zeek nta-backend nta-analyzer nta-probe-manager
```

## 日志查看

### 系统日志
```bash
# Backend 日志
journalctl -u nta-backend -f

# Analyzer 日志
journalctl -u nta-analyzer -f

# Probe Manager 日志
journalctl -u nta-probe-manager -f

# Zeek 日志
journalctl -u nta-zeek -f
```

### 应用日志
```bash
# 探针日志
tail -f /opt/nta-probe/logs/*.log

# Zeek 原始日志
tail -f /var/spool/zeek/*.log
```

## 架构说明

### 服务组件

```
┌─────────────────┐
│  nta-zeek       │  Zeek 流量采集
│  (root)         │  监听网络接口
└────────┬────────┘
         │
┌────────▼────────┐
│  nta-analyzer   │  流量分析引擎
│  (nta:5001)     │  威胁检测
└────────┬────────┘
         │
┌────────▼────────┐  ┌─────────────────┐
│  nta-backend    │  │ nta-probe-mgr   │
│  (nta:5000)     │  │ (nta:6000)      │
│  Web API        │  │ 探针管理        │
└─────────────────┘  └─────────────────┘
         │                    │
         └────────┬───────────┘
                  │
           ┌──────▼──────┐
           │   Redis     │
           │   :6379     │
           └─────────────┘
```

### 目录结构

```
/opt/nta-probe/
├── analyzer/              # 分析引擎
├── backend/               # Web API
├── asset_discovery/       # 资产发现
├── threat_intel_service/  # 威胁情报
├── probe_manager/         # 探针管理
├── encryption_analyzer/   # 加密流量分析
├── audit_service/         # 审计服务
├── license_service/       # 授权管理
├── report_service/        # 报表生成
├── apt_detector/          # APT检测
├── zeek-scripts/          # Zeek 脚本
├── config/                # 配置文件
├── logs/                  # 日志目录
├── reports/               # 报告目录
└── templates/             # 报告模板
```

## 配置说明

### 网络接口配置

编辑 Zeek 配置：
```bash
vim /opt/zeek/etc/node.cfg
```

修改监听接口：
```ini
[zeek]
type=standalone
host=localhost
interface=eth0  # 修改为实际网卡
```

重启服务：
```bash
systemctl restart nta-zeek
```

### 探针配置

配置文件位置：`/opt/nta-probe/config/`

- `asset_discovery.json` - 资产发现配置
- `threat_intel.json` - 威胁情报源配置
- `probe_manager.json` - 探针管理配置
- `license.json` - License 配置
- `apt_iocs.json` - APT IOC 库

### Redis 配置

默认连接本地 Redis: `redis://localhost:6379`

修改配置：
```bash
vim /opt/nta-probe/config/probe_manager.json
```

## 功能特性

### ✅ 核心检测能力
- 横向移动检测 (SMB/RDP/SSH)
- 异常登录检测
- 数据窃取检测
- C2 通信检测
- 机器学习异常检测

### ✅ 商业增强功能
- 🔍 **资产发现** - 自动识别网络资产和服务指纹
- 🛡️ **威胁情报** - 对接 ThreatFox 等情报源
- 🌐 **多探针协同** - 支持分布式探针管理
- 🔐 **加密流量分析** - TLS 元数据分析、JA3 指纹
- 📝 **审计日志** - 操作审计和合规记录
- 🔑 **License 授权** - 功能授权和时间控制
- 📊 **报表生成** - PDF/Excel 报告导出
- 🎯 **APT 检测** - Kill Chain 分析、IOC 狩猎

## 访问地址

- **Backend API**: http://服务器IP:5000
- **Probe Manager**: http://服务器IP:6000
- **Redis**: localhost:6379

## 卸载

```bash
# 停止服务
systemctl stop nta-zeek nta-backend nta-analyzer nta-probe-manager

# 禁用服务
systemctl disable nta-zeek nta-backend nta-analyzer nta-probe-manager

# 删除服务文件
rm -f /etc/systemd/system/nta-*.service
systemctl daemon-reload

# 删除安装目录
rm -rf /opt/nta-probe

# 删除 Zeek (可选)
rm -rf /opt/zeek

# 删除服务用户
userdel -r nta
```

## 故障排查

### 服务无法启动

1. 检查日志：
```bash
journalctl -u nta-backend -n 50
```

2. 检查端口占用：
```bash
netstat -tlnp | grep -E '5000|6000'
```

3. 检查权限：
```bash
ls -la /opt/nta-probe
```

### Zeek 无法采集流量

1. 检查网卡权限：
```bash
ip link show
```

2. 确认网卡名称正确：
```bash
zeekctl config | grep interface
```

3. 手动重启 Zeek：
```bash
zeekctl restart
```

## 技术支持

- GitHub Issues: https://github.com/Cxiyuan/NTA/issues
- 邮箱: contact@example.com

## License

本项目采用商业授权模式，使用前请联系获取 License。