#!/usr/bin/env python3

import os
import sys
import json
import argparse
from datetime import datetime
from collections import defaultdict

ZEEK_LOG_DIR = "/var/log/zeek/current"

def tail_follow(filepath):
    with open(filepath, 'r') as f:
        f.seek(0, 2)
        while True:
            line = f.readline()
            if line:
                yield line.strip()
            else:
                import time
                time.sleep(0.1)

def monitor_lateral_movement(log_file):
    print(f"[*] 开始监控横向移动日志: {log_file}")
    print("[*] 按 Ctrl+C 停止监控\n")
    
    alert_count = defaultdict(int)
    
    try:
        for line in tail_follow(log_file):
            if not line or line.startswith('#'):
                continue
            
            try:
                fields = line.split('\t')
                if len(fields) < 9:
                    continue
                
                ts = datetime.fromtimestamp(float(fields[0]))
                attack_type = fields[6]
                severity = fields[7]
                description = fields[8]
                orig_h = fields[2]
                resp_h = fields[4]
                
                alert_count[attack_type] += 1
                
                severity_color = {
                    'CRITICAL': '\033[91m',
                    'HIGH': '\033[93m',
                    'MEDIUM': '\033[92m'
                }.get(severity, '\033[0m')
                
                print(f"{severity_color}[{severity}]\033[0m {ts.strftime('%Y-%m-%d %H:%M:%S')}")
                print(f"  类型: {attack_type}")
                print(f"  源IP: {orig_h} -> 目标IP: {resp_h}")
                print(f"  描述: {description}")
                print("-" * 80)
                
            except Exception as e:
                print(f"[!] 解析错误: {e}", file=sys.stderr)
    
    except KeyboardInterrupt:
        print("\n\n[*] 监控停止")
        print("\n=== 统计信息 ===")
        for attack_type, count in sorted(alert_count.items(), key=lambda x: x[1], reverse=True):
            print(f"  {attack_type}: {count} 次")

def analyze_logs(log_dir, hours=1):
    print(f"[*] 分析最近 {hours} 小时的日志...")
    
    lateral_log = os.path.join(log_dir, "lateral_movement.log")
    
    if not os.path.exists(lateral_log):
        print(f"[!] 日志文件不存在: {lateral_log}")
        return
    
    from datetime import timedelta
    cutoff_time = datetime.now() - timedelta(hours=hours)
    
    alerts = []
    stats = defaultdict(lambda: defaultdict(int))
    
    with open(lateral_log, 'r') as f:
        for line in f:
            if line.startswith('#'):
                continue
            
            fields = line.strip().split('\t')
            if len(fields) < 9:
                continue
            
            try:
                ts = datetime.fromtimestamp(float(fields[0]))
                if ts < cutoff_time:
                    continue
                
                attack_type = fields[6]
                severity = fields[7]
                orig_h = fields[2]
                
                alerts.append({
                    'timestamp': ts,
                    'type': attack_type,
                    'severity': severity,
                    'source': orig_h
                })
                
                stats[attack_type]['count'] += 1
                stats[attack_type]['sources'].add(orig_h)
                
            except Exception as e:
                continue
    
    print(f"\n总告警数: {len(alerts)}")
    print("\n=== 攻击类型分布 ===")
    for attack_type, data in sorted(stats.items(), key=lambda x: x[1]['count'], reverse=True):
        print(f"  {attack_type}: {data['count']} 次, 来源IP: {len(data['sources'])} 个")
    
    print("\n=== 最近10条告警 ===")
    for alert in sorted(alerts, key=lambda x: x['timestamp'], reverse=True)[:10]:
        print(f"  [{alert['severity']}] {alert['timestamp'].strftime('%H:%M:%S')} - {alert['type']} from {alert['source']}")

def main():
    parser = argparse.ArgumentParser(description='Zeek横向移动检测日志监控工具')
    parser.add_argument('-m', '--monitor', action='store_true', help='实时监控模式')
    parser.add_argument('-a', '--analyze', action='store_true', help='分析历史日志')
    parser.add_argument('-l', '--log-dir', default=ZEEK_LOG_DIR, help='Zeek日志目录')
    parser.add_argument('-H', '--hours', type=int, default=1, help='分析最近N小时的日志')
    
    args = parser.parse_args()
    
    if args.monitor:
        log_file = os.path.join(args.log_dir, "lateral_movement.log")
        monitor_lateral_movement(log_file)
    elif args.analyze:
        analyze_logs(args.log_dir, args.hours)
    else:
        parser.print_help()

if __name__ == '__main__':
    main()
