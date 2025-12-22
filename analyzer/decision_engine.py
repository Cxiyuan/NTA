#!/usr/bin/env python3

import json
import numpy as np
from datetime import datetime
from collections import defaultdict

class BayesianFusion:
    def __init__(self):
        self.prior_attack = 0.001
        
        self.detector_accuracy = {
            'zeek_scan': {'tpr': 0.90, 'fpr': 0.10},
            'zeek_auth': {'tpr': 0.90, 'fpr': 0.08},
            'zeek_exec': {'tpr': 0.85, 'fpr': 0.12},
            'zeek_dpi': {'tpr': 0.80, 'fpr': 0.15},
            'zeek_encrypted': {'tpr': 0.75, 'fpr': 0.20},
            'zeek_zeroday': {'tpr': 0.70, 'fpr': 0.25},
            'ml_anomaly': {'tpr': 0.85, 'fpr': 0.10},
            'graph_analysis': {'tpr': 0.80, 'fpr': 0.12},
            'threat_intel': {'tpr': 0.95, 'fpr': 0.02},
            'baseline_deviation': {'tpr': 0.75, 'fpr': 0.18}
        }
        
        self.detector_weights = {
            'zeek_scan': 1.0,
            'zeek_auth': 1.2,
            'zeek_exec': 1.3,
            'zeek_dpi': 0.9,
            'zeek_encrypted': 0.8,
            'zeek_zeroday': 0.7,
            'ml_anomaly': 1.1,
            'graph_analysis': 1.0,
            'threat_intel': 1.5,
            'baseline_deviation': 0.9
        }
    
    def calculate_posterior(self, detections):
        likelihood = 1.0
        evidence = 1.0
        
        for detector, triggered in detections.items():
            if detector not in self.detector_accuracy:
                continue
            
            accuracy = self.detector_accuracy[detector]
            
            if triggered:
                likelihood *= accuracy['tpr']
                evidence *= (accuracy['tpr'] * self.prior_attack + 
                           accuracy['fpr'] * (1 - self.prior_attack))
            else:
                likelihood *= (1 - accuracy['tpr'])
                evidence *= ((1 - accuracy['tpr']) * self.prior_attack + 
                           (1 - accuracy['fpr']) * (1 - self.prior_attack))
        
        if evidence == 0:
            return 0.0
        
        posterior = (likelihood * self.prior_attack) / evidence
        
        return posterior
    
    def weighted_vote(self, detection_scores):
        total_weight = 0.0
        weighted_sum = 0.0
        
        for detector, score in detection_scores.items():
            if detector in self.detector_weights:
                weight = self.detector_weights[detector]
                weighted_sum += score * weight
                total_weight += weight
        
        if total_weight == 0:
            return 0.0
        
        return weighted_sum / total_weight
    
    def decide(self, detections, detection_scores=None):
        bayes_prob = self.calculate_posterior(detections)
        
        if detection_scores:
            vote_score = self.weighted_vote(detection_scores)
            final_score = bayes_prob * 0.6 + vote_score * 0.4
        else:
            final_score = bayes_prob
        
        decision = {
            'bayesian_probability': float(bayes_prob),
            'final_score': float(final_score),
            'action': self._determine_action(final_score),
            'confidence': self._calculate_confidence(detections),
            'detections': detections,
            'timestamp': datetime.now().isoformat()
        }
        
        return decision
    
    def _determine_action(self, score):
        if score >= 0.9999:
            return 'BLOCK_IMMEDIATELY'
        elif score >= 0.99:
            return 'ALERT_SOC_URGENT'
        elif score >= 0.95:
            return 'ALERT_SOC_HIGH'
        elif score >= 0.90:
            return 'ALERT_SOC_NORMAL'
        elif score >= 0.80:
            return 'MONITOR_CLOSELY'
        else:
            return 'LOG_ONLY'
    
    def _calculate_confidence(self, detections):
        triggered_count = sum(1 for v in detections.values() if v)
        total_detectors = len(detections)
        
        if total_detectors == 0:
            return 0.0
        
        agreement_ratio = triggered_count / total_detectors
        
        if triggered_count >= 5:
            return 0.95
        elif triggered_count >= 3:
            return 0.85
        elif triggered_count >= 2:
            return 0.70
        elif triggered_count == 1:
            return 0.50
        else:
            return 0.20


class MultiLayerDecisionEngine:
    def __init__(self):
        self.bayesian = BayesianFusion()
        self.alert_history = defaultdict(list)
        self.threshold_config = {
            'auto_block': 0.9999,
            'urgent_alert': 0.99,
            'high_alert': 0.95,
            'normal_alert': 0.90,
            'monitor': 0.80
        }
    
    def process_event(self, event, detections, detection_scores=None):
        decision = self.bayesian.decide(detections, detection_scores)
        
        src_ip = event.get('src_ip', 'unknown')
        self.alert_history[src_ip].append({
            'timestamp': datetime.now(),
            'score': decision['final_score'],
            'action': decision['action']
        })
        
        context = self._add_context(event, src_ip)
        decision['context'] = context
        
        decision = self._apply_business_rules(decision, event)
        
        if decision['action'] in ['BLOCK_IMMEDIATELY', 'ALERT_SOC_URGENT']:
            self._enrich_critical_alert(decision, event)
        
        return decision
    
    def _add_context(self, event, src_ip):
        context = {
            'previous_alerts': len(self.alert_history[src_ip]),
            'is_repeat_offender': len(self.alert_history[src_ip]) >= 3,
            'alert_frequency': 0.0
        }
        
        if len(self.alert_history[src_ip]) > 1:
            recent_alerts = [a for a in self.alert_history[src_ip] 
                           if (datetime.now() - a['timestamp']).total_seconds() < 3600]
            context['alert_frequency'] = len(recent_alerts) / 60.0
        
        return context
    
    def _apply_business_rules(self, decision, event):
        src_ip = event.get('src_ip')
        dst_ip = event.get('dst_ip')
        
        vip_hosts = set(['10.0.1.1', '10.0.2.1'])
        critical_servers = set(['10.0.3.1', '10.0.3.2'])
        
        if dst_ip in vip_hosts or dst_ip in critical_servers:
            decision['final_score'] = min(decision['final_score'] * 1.3, 1.0)
            decision['action'] = self.bayesian._determine_action(decision['final_score'])
            decision['context']['target_criticality'] = 'HIGH'
        
        if decision['context'].get('is_repeat_offender'):
            decision['final_score'] = min(decision['final_score'] * 1.2, 1.0)
            decision['action'] = self.bayesian._determine_action(decision['final_score'])
        
        working_hours = 9 <= datetime.now().hour <= 17
        if not working_hours and decision['final_score'] > 0.80:
            decision['final_score'] = min(decision['final_score'] * 1.15, 1.0)
            decision['action'] = self.bayesian._determine_action(decision['final_score'])
            decision['context']['off_hours'] = True
        
        return decision
    
    def _enrich_critical_alert(self, decision, event):
        decision['investigation'] = {
            'recommended_actions': [
                "隔离源IP地址",
                "检查受影响主机的进程列表",
                "收集网络流量PCAP",
                "检查相关主机的登录日志",
                "扫描受影响系统的IOC"
            ],
            'ioc_collection': {
                'src_ip': event.get('src_ip'),
                'dst_ip': event.get('dst_ip'),
                'timestamp': event.get('timestamp'),
                'protocols': event.get('protocols', []),
                'files_transferred': event.get('files', [])
            }
        }
    
    def generate_alert_report(self, decision, event):
        report = {
            'alert_id': f"ALERT-{datetime.now().strftime('%Y%m%d%H%M%S')}",
            'timestamp': decision['timestamp'],
            'severity': self._map_action_to_severity(decision['action']),
            'confidence': decision['confidence'],
            'score': decision['final_score'],
            'event_summary': {
                'source': event.get('src_ip'),
                'destination': event.get('dst_ip'),
                'type': event.get('event_type'),
                'description': event.get('description', 'Unknown')
            },
            'detections': decision['detections'],
            'context': decision.get('context', {}),
            'recommended_action': decision['action'],
            'investigation': decision.get('investigation', {})
        }
        
        return report
    
    def _map_action_to_severity(self, action):
        mapping = {
            'BLOCK_IMMEDIATELY': 'CRITICAL',
            'ALERT_SOC_URGENT': 'CRITICAL',
            'ALERT_SOC_HIGH': 'HIGH',
            'ALERT_SOC_NORMAL': 'MEDIUM',
            'MONITOR_CLOSELY': 'LOW',
            'LOG_ONLY': 'INFO'
        }
        return mapping.get(action, 'UNKNOWN')


if __name__ == '__main__':
    print("多层决策融合引擎测试\n")
    
    engine = MultiLayerDecisionEngine()
    
    test_event = {
        'src_ip': '192.168.1.100',
        'dst_ip': '10.0.1.50',
        'event_type': 'LATERAL_MOVEMENT',
        'description': 'PSExec远程执行检测',
        'timestamp': datetime.now().isoformat()
    }
    
    test_detections = {
        'zeek_scan': True,
        'zeek_auth': True,
        'zeek_exec': True,
        'zeek_dpi': False,
        'zeek_encrypted': True,
        'zeek_zeroday': False,
        'ml_anomaly': True,
        'graph_analysis': True,
        'threat_intel': False,
        'baseline_deviation': True
    }
    
    test_scores = {
        'zeek_scan': 0.85,
        'zeek_auth': 0.90,
        'zeek_exec': 0.92,
        'zeek_encrypted': 0.75,
        'ml_anomaly': 0.88,
        'graph_analysis': 0.82,
        'baseline_deviation': 0.78
    }
    
    decision = engine.process_event(test_event, test_detections, test_scores)
    
    print("决策结果:")
    print(json.dumps(decision, indent=2, ensure_ascii=False))
    
    print("\n告警报告:")
    report = engine.generate_alert_report(decision, test_event)
    print(json.dumps(report, indent=2, ensure_ascii=False))
