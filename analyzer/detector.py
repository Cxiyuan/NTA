#!/usr/bin/env python3

import json
import sys
import time
from collections import defaultdict
from datetime import datetime, timedelta
from typing import Dict, List, Set
import re

class LateralMovementDetector:
    def __init__(self, config_file='../config/detection.yaml'):
        self.scan_threshold = 20
        self.auth_fail_threshold = 5
        self.time_window = 300
        
        self.connection_tracker = defaultdict(lambda: defaultdict(set))
        self.auth_failures = defaultdict(lambda: defaultdict(int))
        self.ntlm_hashes = defaultdict(set)
        self.wmi_activities = defaultdict(set)
        self.alerts = []

    def process_conn_log(self, line: str):
        try:
            record = json.loads(line)
            
            if not self._is_internal(record.get('id.orig_h', '')) or \
               not self._is_internal(record.get('id.resp_h', '')):
                return
            
            orig_h = record['id.orig_h']
            resp_h = record['id.resp_h']
            resp_p = record.get('id.resp_p', 0)
            
            if resp_p in [135, 139, 445, 3389, 22, 5985, 5986]:
                self.connection_tracker[orig_h]['targets'].add(resp_h)
                self.connection_tracker[orig_h]['ports'].add(resp_p)
                self.connection_tracker[orig_h]['timestamp'] = record['ts']
                
                self._check_lateral_scan(orig_h, record)
        
        except json.JSONDecodeError:
            pass
        except Exception as e:
            print(f"Error processing conn.log: {e}", file=sys.stderr)

    def process_ntlm_log(self, line: str):
        try:
            record = json.loads(line)
            
            if 'ntlm_response' in record:
                ntlm_hash = record['ntlm_response']
                orig_h = record['id.orig_h']
                
                self.ntlm_hashes[ntlm_hash].add(orig_h)
                
                if len(self.ntlm_hashes[ntlm_hash]) >= 3:
                    self._alert_pth_attack(ntlm_hash, record)
        
        except json.JSONDecodeError:
            pass
        except Exception as e:
            print(f"Error processing ntlm.log: {e}", file=sys.stderr)

    def process_smb_log(self, line: str):
        try:
            record = json.loads(line)
            
            if record.get('action') == 'SMB::FILE_OPEN':
                path = record.get('path', '')
                
                if any(x in path for x in ['ADMIN$', 'C$', 'IPC$']):
                    orig_h = record['id.orig_h']
                    resp_h = record['id.resp_h']
                    
                    key = f"{orig_h}->{resp_h}"
                    self.connection_tracker[key]['admin_shares'].add(path)
                    
                    if len(self.connection_tracker[key]['admin_shares']) >= 2:
                        self._alert_psexec(orig_h, resp_h, record)
            
            if record.get('status') and record['status'] != 'STATUS_SUCCESS':
                orig_h = record['id.orig_h']
                resp_h = record['id.resp_h']
                self.auth_failures[orig_h][resp_h] += 1
                
                if self.auth_failures[orig_h][resp_h] >= self.auth_fail_threshold:
                    self._alert_bruteforce(orig_h, resp_h, 'SMB', record)
        
        except json.JSONDecodeError:
            pass
        except Exception as e:
            print(f"Error processing smb.log: {e}", file=sys.stderr)

    def process_dce_rpc_log(self, line: str):
        try:
            record = json.loads(line)
            
            endpoint = record.get('endpoint', '')
            wmi_endpoints = ['IWbemServices', 'ISystemActivator', 'IWbemLevel1Login']
            
            if any(wmi in endpoint for wmi in wmi_endpoints):
                orig_h = record['id.orig_h']
                resp_h = record['id.resp_h']
                key = f"{orig_h}->{resp_h}"
                
                self.wmi_activities[key].add(endpoint)
                
                if len(self.wmi_activities[key]) >= 2:
                    self._alert_wmi_execution(orig_h, resp_h, record)
        
        except json.JSONDecodeError:
            pass
        except Exception as e:
            print(f"Error processing dce_rpc.log: {e}", file=sys.stderr)

    def process_rdp_log(self, line: str):
        try:
            record = json.loads(line)
            
            if 'cookie' in record:
                orig_h = record['id.orig_h']
                resp_h = record['id.resp_h']
                
                self.connection_tracker[orig_h]['rdp_targets'].add(resp_h)
                
                if len(self.connection_tracker[orig_h]['rdp_targets']) >= 5:
                    self._alert_rdp_hopping(orig_h, record)
        
        except json.JSONDecodeError:
            pass
        except Exception as e:
            print(f"Error processing rdp.log: {e}", file=sys.stderr)

    def _is_internal(self, ip: str) -> bool:
        if not ip:
            return False
        
        internal_ranges = [
            r'^10\.',
            r'^172\.(1[6-9]|2[0-9]|3[0-1])\.',
            r'^192\.168\.'
        ]
        
        for pattern in internal_ranges:
            if re.match(pattern, ip):
                return True
        return False

    def _check_lateral_scan(self, orig_h: str, record: dict):
        tracker = self.connection_tracker[orig_h]
        target_count = len(tracker['targets'])
        port_count = len(tracker['ports'])
        
        if target_count >= self.scan_threshold:
            alert = {
                'timestamp': datetime.now().isoformat(),
                'type': 'LATERAL_SCAN',
                'severity': 'HIGH',
                'source_ip': orig_h,
                'target_count': target_count,
                'port_count': port_count,
                'description': f'横向扫描检测: {orig_h} 扫描了 {target_count} 个内网主机',
                'targets': list(tracker['targets'])[:10],
                'ports': list(tracker['ports'])
            }
            self.alerts.append(alert)
            print(json.dumps(alert, ensure_ascii=False))

    def _alert_pth_attack(self, ntlm_hash: str, record: dict):
        hosts = list(self.ntlm_hashes[ntlm_hash])
        alert = {
            'timestamp': datetime.now().isoformat(),
            'type': 'PASS_THE_HASH',
            'severity': 'CRITICAL',
            'ntlm_hash': ntlm_hash[:16] + '...',
            'affected_hosts': hosts,
            'host_count': len(hosts),
            'description': f'Pass-the-Hash攻击: 相同NTLM Hash在 {len(hosts)} 台主机上使用',
            'evidence': f'涉及主机: {", ".join(hosts)}'
        }
        self.alerts.append(alert)
        print(json.dumps(alert, ensure_ascii=False))

    def _alert_psexec(self, orig_h: str, resp_h: str, record: dict):
        alert = {
            'timestamp': datetime.now().isoformat(),
            'type': 'PSEXEC',
            'severity': 'CRITICAL',
            'source_ip': orig_h,
            'target_ip': resp_h,
            'description': f'PSExec远程执行: {orig_h} -> {resp_h}',
            'evidence': '检测到管理共享写入行为'
        }
        self.alerts.append(alert)
        print(json.dumps(alert, ensure_ascii=False))

    def _alert_wmi_execution(self, orig_h: str, resp_h: str, record: dict):
        key = f"{orig_h}->{resp_h}"
        endpoints = list(self.wmi_activities[key])
        
        alert = {
            'timestamp': datetime.now().isoformat(),
            'type': 'WMI_EXECUTION',
            'severity': 'CRITICAL',
            'source_ip': orig_h,
            'target_ip': resp_h,
            'description': f'WMI远程执行: {orig_h} -> {resp_h}',
            'evidence': f'调用接口: {", ".join(endpoints)}'
        }
        self.alerts.append(alert)
        print(json.dumps(alert, ensure_ascii=False))

    def _alert_bruteforce(self, orig_h: str, resp_h: str, protocol: str, record: dict):
        fail_count = self.auth_failures[orig_h][resp_h]
        
        alert = {
            'timestamp': datetime.now().isoformat(),
            'type': f'{protocol}_BRUTEFORCE',
            'severity': 'CRITICAL',
            'source_ip': orig_h,
            'target_ip': resp_h,
            'protocol': protocol,
            'fail_count': fail_count,
            'description': f'{protocol}暴力破解: {orig_h} -> {resp_h}',
            'evidence': f'失败认证 {fail_count} 次'
        }
        self.alerts.append(alert)
        print(json.dumps(alert, ensure_ascii=False))

    def _alert_rdp_hopping(self, orig_h: str, record: dict):
        targets = list(self.connection_tracker[orig_h]['rdp_targets'])
        
        alert = {
            'timestamp': datetime.now().isoformat(),
            'type': 'RDP_HOPPING',
            'severity': 'HIGH',
            'source_ip': orig_h,
            'target_count': len(targets),
            'description': f'RDP跳板行为: {orig_h} 连接了 {len(targets)} 台主机',
            'targets': targets[:10]
        }
        self.alerts.append(alert)
        print(json.dumps(alert, ensure_ascii=False))

    def get_statistics(self) -> dict:
        return {
            'total_alerts': len(self.alerts),
            'alert_types': defaultdict(int, {
                alert['type']: self.alerts.count(alert)
                for alert in self.alerts
            }),
            'monitored_hosts': len(self.connection_tracker),
            'tracked_ntlm_hashes': len(self.ntlm_hashes)
        }


def main():
    detector = LateralMovementDetector()
    
    print("横向移动检测分析引擎已启动...", file=sys.stderr)
    print("支持日志类型: conn.log, ntlm.log, smb.log, dce_rpc.log, rdp.log", file=sys.stderr)
    
    log_handlers = {
        'conn': detector.process_conn_log,
        'ntlm': detector.process_ntlm_log,
        'smb_files': detector.process_smb_log,
        'smb_mapping': detector.process_smb_log,
        'dce_rpc': detector.process_dce_rpc_log,
        'rdp': detector.process_rdp_log
    }
    
    for line in sys.stdin:
        line = line.strip()
        if not line:
            continue
        
        try:
            record = json.loads(line)
            log_type = record.get('_path', '')
            
            handler = log_handlers.get(log_type)
            if handler:
                handler(line)
        
        except Exception as e:
            print(f"Error processing line: {e}", file=sys.stderr)
    
    stats = detector.get_statistics()
    print("\n=== 检测统计 ===", file=sys.stderr)
    print(json.dumps(stats, ensure_ascii=False, indent=2), file=sys.stderr)


if __name__ == '__main__':
    main()
