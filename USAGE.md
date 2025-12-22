# Zeek横向渗透检测探针使用文档

## 快速开始

### 1. 安装部署

```bash
cd /root/cap_agent
sudo ./deploy/setup.sh
```

安装过程会自动完成:
- 检查并安装Zeek
- 安装Python依赖
- 部署检测脚本
- 配置网络接口
- 启动探针服务

### 2. 验证运行状态

```bash
zeekctl status
```

查看日志:
```bash
tail -f /var/log/zeek/current/lateral_movement.log
tail -f /var/log/zeek/current/notice.log
```

### 3. 实时监控

```bash
python3 /root/cap_agent/analyzer/monitor.py -m
```

### 4. 历史分析

```bash
python3 /root/cap_agent/analyzer/monitor.py -a -H 24
```

## 检测能力

### 横向扫描检测
- **特征**: 单IP短时间内扫描大量内网主机
- **阈值**: 默认20台主机/5分钟
- **日志**: `LATERAL_SCAN`

### 认证暴力破解
- **协议**: SMB, RDP, SSH
- **特征**: 多次认证失败
- **阈值**: 5次失败/5分钟
- **日志**: `SMB_BRUTEFORCE`, `RDP_BRUTEFORCE`

### Pass-the-Hash攻击
- **特征**: 相同NTLM Hash在多台主机使用
- **阈值**: 3台主机/1小时
- **日志**: `PASS_THE_HASH`

### PSExec远程执行
- **特征**: 向ADMIN$/C$写入文件
- **协议**: SMB
- **日志**: `PSEXEC`

### WMI远程执行
- **特征**: 调用IWbemServices等WMI接口
- **协议**: DCE-RPC
- **日志**: `WMI_EXECUTION`

### RDP跳板
- **特征**: 单IP连接多台主机RDP
- **阈值**: 5台主机
- **日志**: `RDP_HOPPING`

## 配置说明

编辑配置文件: `/root/cap_agent/config/detection.yaml`

```yaml
detection:
  scan:
    threshold: 20              # 扫描检测阈值
    time_window: 300           # 时间窗口(秒)
  
  authentication:
    fail_threshold: 5          # 认证失败阈值
    pth_window: 3600          # PTH检测窗口(秒)
```

## 告警输出

### 日志格式

**Zeek原生日志** (`/var/log/zeek/current/lateral_movement.log`):
```
时间戳  UID  源IP  源端口  目标IP  目标端口  攻击类型  严重级别  描述  证据
```

**JSON告警** (Python分析引擎):
```json
{
  "timestamp": "2025-12-22T10:30:45",
  "type": "LATERAL_SCAN",
  "severity": "HIGH",
  "source_ip": "192.168.1.100",
  "target_count": 25,
  "description": "横向扫描检测"
}
```

## 运维命令

### Zeek控制
```bash
zeekctl start      # 启动
zeekctl stop       # 停止
zeekctl restart    # 重启
zeekctl deploy     # 重新部署配置
zeekctl diag       # 诊断
```

### 日志管理
```bash
ls /var/log/zeek/current/
zeek-cut < conn.log ts id.orig_h id.resp_h id.resp_p
```

### 分析引擎
```bash
./deploy/start_analyzer.sh
python3 analyzer/monitor.py -m
python3 analyzer/monitor.py -a -H 24
```

## 性能优化

### 调整Zeek进程数
编辑 `/usr/local/zeek/etc/node.cfg`:
```ini
[worker-1]
type=worker
host=localhost
interface=eth0
lb_method=pf_ring
lb_procs=4
```

### 内存优化
```zeek
redef table_expire_interval = 5min;
redef table_incremental_step = 1000;
```

## 故障排查

### Zeek未启动
```bash
zeekctl diag
tail -f /usr/local/zeek/logs/current/stderr.log
```

### 无告警输出
1. 检查网卡是否正确: `zeekctl config | grep interface`
2. 检查流量是否镜像: `tcpdump -i eth0 -c 10`
3. 检查内网配置: `cat /usr/local/zeek/etc/networks.cfg`

### Python分析引擎报错
```bash
pip3 install -r requirements.txt
python3 -c "import yaml; print('OK')"
```

## 集成方案

### 与SIEM集成
```bash
tail -F /var/log/zeek/current/lateral_movement.log | \
  filebeat -c filebeat.yml
```

### 与Kafka集成
```python
from kafka import KafkaProducer
producer = KafkaProducer(bootstrap_servers='localhost:9092')

for alert in detector.alerts:
    producer.send('lateral-alerts', json.dumps(alert).encode())
```

## 架构图

```
┌─────────────────┐
│  网络镜像流量    │
└────────┬────────┘
         │
    ┌────▼─────┐
    │   Zeek   │
    │  Engine  │
    └────┬─────┘
         │
    ┌────▼─────────────────┐
    │  Zeek Scripts        │
    │ - lateral-scan.zeek  │
    │ - lateral-auth.zeek  │
    │ - lateral-exec.zeek  │
    └────┬─────────────────┘
         │
    ┌────▼──────────────┐
    │  Log Files        │
    │ - conn.log        │
    │ - ntlm.log        │
    │ - smb.log         │
    │ - lateral_*.log   │
    └────┬──────────────┘
         │
    ┌────▼─────────────┐
    │ Python Analyzer  │
    │  detector.py     │
    └────┬─────────────┘
         │
    ┌────▼─────────┐
    │   Alerts     │
    │ JSON/Syslog  │
    └──────────────┘
```

## 参考资料

- [Zeek官方文档](https://docs.zeek.org/)
- [MITRE ATT&CK - Lateral Movement](https://attack.mitre.org/tactics/TA0008/)
- [横向移动检测最佳实践](https://www.sans.org/white-papers/)
