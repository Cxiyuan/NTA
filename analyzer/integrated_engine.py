#!/usr/bin/env python3

import sys
import json
import argparse
from datetime import datetime
from detector import LateralMovementDetector
from ml_detector import MLAnomalyDetector, BaselineLearner, CircadianAnalyzer
from graph_analyzer import LateralGraphAnalyzer
from threat_intel import ThreatIntelligence
from decision_engine import MultiLayerDecisionEngine
from report_generator import ReportGenerator

class IntegratedAnalysisEngine:
    def __init__(self, config_file='../config/detection.yaml'):
        print("[+] 初始化综合分析引擎...")
        
        self.detector = LateralMovementDetector(config_file)
        self.ml_detector = MLAnomalyDetector(config_file)
        self.baseline_learner = BaselineLearner()
        self.circadian_analyzer = CircadianAnalyzer()
        self.graph_analyzer = LateralGraphAnalyzer()
        self.threat_intel = ThreatIntelligence(config_file)
        self.decision_engine = MultiLayerDecisionEngine()
        self.report_generator = ReportGenerator()
        
        self.alerts = []
        self.statistics = {
            'zeek_scan': 0,
            'zeek_auth': 0,
            'zeek_exec': 0,
            'zeek_dpi': 0,
            'zeek_encrypted': 0,
            'zeek_zeroday': 0,
            'ml_anomaly': 0,
            'graph_analysis': 0,
            'threat_intel': 0,
            'baseline_deviation': 0,
            'total_processed': 0,
            'apt_campaigns': 0
        }
        
        print("[+] 引擎初始化完成")
    
    def process_zeek_log(self, log_line):
        try:
            record = json.loads(log_line)
            log_type = record.get('_path', '')
            
            if log_type == 'conn':
                self._process_connection(record)
            elif log_type == 'ntlm':
                self._process_ntlm(record)
            elif log_type == 'smb_files':
                self._process_smb(record)
            elif log_type == 'dce_rpc':
                self._process_dce_rpc(record)
            elif log_type == 'ssl':
                self._process_ssl(record)
            
            self.statistics['total_processed'] += 1
            
        except json.JSONDecodeError:
            pass
        except Exception as e:
            print(f"[!] 处理日志错误: {e}", file=sys.stderr)
    
    def _process_connection(self, record):
        src_ip = record.get('id.orig_h')
        dst_ip = record.get('id.resp_h')
        protocol = record.get('service', 'unknown')
        
        self.graph_analyzer.add_connection(
            src_ip, dst_ip, protocol,
            datetime.fromtimestamp(record.get('ts', 0))
        )
        
        detections = {}
        detection_scores = {}
        
        self.detector.process_conn_log(json.dumps(record))
        
        conn_data = {
            'connection_rate': 1.0,
            'target_count': 1,
            'port_diversity': 1,
            'failed_auth_ratio': 0.0,
            'avg_packet_size': record.get('orig_bytes', 0),
            'session_duration': record.get('duration', 0),
            'upload_download_ratio': 1.0,
            'inter_arrival_variance': 0.0
        }
        
        ml_result = self.ml_detector.predict(conn_data)
        if ml_result['anomaly']:
            detections['ml_anomaly'] = True
            detection_scores['ml_anomaly'] = ml_result['confidence']
            self.statistics['ml_anomaly'] += 1
        else:
            detections['ml_anomaly'] = False
        
        baseline_event = {
            'timestamp': datetime.fromtimestamp(record.get('ts', 0)),
            'connection_count': 1
        }
        is_baseline_anomaly, score = self.baseline_learner.is_anomalous(baseline_event, src_ip)
        detections['baseline_deviation'] = is_baseline_anomaly
        if is_baseline_anomaly:
            detection_scores['baseline_deviation'] = score / 10.0
            self.statistics['baseline_deviation'] += 1
        
        fanout_anomalies = self.graph_analyzer.detect_anomalous_fanout(threshold=20)
        detections['graph_analysis'] = len(fanout_anomalies) > 0
        if fanout_anomalies:
            detection_scores['graph_analysis'] = fanout_anomalies[0]['score']
            self.statistics['graph_analysis'] += 1
        
        intel_event = {
            'src_ip': src_ip,
            'dst_ip': dst_ip,
            'dst_port': record.get('id.resp_p', 0)
        }
        enrichment = self.threat_intel.enrich_event(intel_event)
        detections['threat_intel'] = enrichment['risk_score'] > 30
        if enrichment['risk_score'] > 30:
            detection_scores['threat_intel'] = enrichment['risk_score'] / 100.0
            self.statistics['threat_intel'] += 1
        
        if sum(detections.values()) >= 2:
            event = {
                'src_ip': src_ip,
                'dst_ip': dst_ip,
                'event_type': 'SUSPICIOUS_CONNECTION',
                'timestamp': datetime.fromtimestamp(record.get('ts', 0)).isoformat()
            }
            
            decision = self.decision_engine.process_event(event, detections, detection_scores)
            
            if decision['final_score'] >= 0.90:
                alert = self.decision_engine.generate_alert_report(decision, event)
                self.alerts.append(alert)
                print(f"[!] 告警: {json.dumps(alert, ensure_ascii=False)}")
    
    def _process_ntlm(self, record):
        self.detector.process_ntlm_log(json.dumps(record))
        self.statistics['zeek_auth'] += 1
    
    def _process_smb(self, record):
        self.detector.process_smb_log(json.dumps(record))
        self.statistics['zeek_exec'] += 1
    
    def _process_dce_rpc(self, record):
        self.detector.process_dce_rpc_log(json.dumps(record))
        self.statistics['zeek_exec'] += 1
    
    def _process_ssl(self, record):
        self.statistics['zeek_encrypted'] += 1
    
    def generate_report(self, output_file='/var/log/zeek/reports/report.html'):
        print(f"\n[+] 生成分析报告...")
        
        self.statistics['apt_campaigns'] = len([a for a in self.alerts if a.get('severity') == 'CRITICAL'])
        self.statistics['time_range'] = '实时分析'
        self.statistics['zeek_status'] = 'Running'
        self.statistics['traffic_processed'] = f"{self.statistics['total_processed']} 条记录"
        self.statistics['active_modules'] = '11/11'
        self.statistics['threat_intel_status'] = 'Active'
        
        report_file = self.report_generator.generate_html_report(
            self.alerts,
            self.statistics,
            output_file
        )
        
        print(f"[+] 报告已生成: {report_file}")
        return report_file
    
    def print_statistics(self):
        print("\n========== 检测统计 ==========")
        print(f"总处理记录: {self.statistics['total_processed']}")
        print(f"ML异常检测: {self.statistics['ml_anomaly']}")
        print(f"图分析告警: {self.statistics['graph_analysis']}")
        print(f"威胁情报匹配: {self.statistics['threat_intel']}")
        print(f"基线偏离: {self.statistics['baseline_deviation']}")
        print(f"总告警数: {len(self.alerts)}")
        print("==============================\n")


def main():
    parser = argparse.ArgumentParser(description='Zeek横向移动综合分析引擎')
    parser.add_argument('-i', '--input', help='输入日志文件路径')
    parser.add_argument('-r', '--report', default='/var/log/zeek/reports/report.html',
                       help='输出报告路径')
    parser.add_argument('--realtime', action='store_true', help='实时分析模式（从stdin读取）')
    
    args = parser.parse_args()
    
    engine = IntegratedAnalysisEngine()
    
    try:
        if args.realtime or not args.input:
            print("[+] 实时分析模式，从stdin读取...")
            for line in sys.stdin:
                line = line.strip()
                if line:
                    engine.process_zeek_log(line)
        else:
            print(f"[+] 分析日志文件: {args.input}")
            with open(args.input, 'r') as f:
                for line in f:
                    line = line.strip()
                    if line:
                        engine.process_zeek_log(line)
        
        engine.print_statistics()
        
        if engine.alerts:
            engine.generate_report(args.report)
        
    except KeyboardInterrupt:
        print("\n[!] 用户中断")
        engine.print_statistics()
        
        if engine.alerts:
            engine.generate_report(args.report)
    
    except Exception as e:
        print(f"[!] 运行错误: {e}", file=sys.stderr)
        import traceback
        traceback.print_exc()


if __name__ == '__main__':
    main()
