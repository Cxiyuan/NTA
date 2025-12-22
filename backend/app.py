#!/usr/bin/env python3

from flask import Flask, jsonify, request
from flask_cors import CORS
from flask_socketio import SocketIO, emit
import sys
import os
import yaml
from datetime import datetime, timedelta
import random

sys.path.append(os.path.join(os.path.dirname(__file__), '..', 'analyzer'))

from detector import LateralMovementDetector
from ml_detector import MLAnomalyDetector
from graph_analyzer import LateralGraphAnalyzer
from threat_intel import ThreatIntelligence
from decision_engine import MultiLayerDecisionEngine

app = Flask(__name__)
CORS(app)
socketio = SocketIO(app, cors_allowed_origins="*")

detector = LateralMovementDetector()
ml_detector = MLAnomalyDetector()
graph_analyzer = LateralGraphAnalyzer()
threat_intel = ThreatIntelligence()
decision_engine = MultiLayerDecisionEngine()

@app.route('/api/stats', methods=['GET'])
def get_stats():
    """获取统计信息"""
    return jsonify({
        'critical': random.randint(10, 20),
        'high': random.randint(30, 50),
        'medium': random.randint(50, 80),
        'low': random.randint(70, 100),
        'apt': random.randint(1, 5),
        'traffic': f'{random.uniform(5, 15):.1f}GB'
    })

@app.route('/api/stats/trend', methods=['GET'])
def get_trend():
    """获取趋势数据"""
    time_range = request.args.get('range', '24h')
    
    hours = 24 if time_range == '24h' else 7 if time_range == '7d' else 1
    data = []
    
    for i in range(hours):
        data.append({
            'time': f'{i:02d}:00',
            'critical': random.randint(5, 20),
            'high': random.randint(20, 50),
            'medium': random.randint(30, 60),
            'low': random.randint(40, 80)
        })
    
    return jsonify(data)

@app.route('/api/alerts', methods=['GET'])
def get_alerts():
    """获取告警列表"""
    page = int(request.args.get('page', 1))
    page_size = int(request.args.get('page_size', 20))
    severity = request.args.get('severity', '')
    
    alerts = []
    for i in range(100):
        alerts.append({
            'id': 1000 + i,
            'timestamp': (datetime.now() - timedelta(minutes=i*5)).strftime('%Y-%m-%d %H:%M:%S'),
            'severity': random.choice(['CRITICAL', 'HIGH', 'MEDIUM', 'LOW']),
            'type': random.choice(['PTH攻击', '横向扫描', 'PSExec', 'WMI执行', 'RDP跳板']),
            'source': f'192.168.1.{random.randint(100, 200)}',
            'target': f'10.0.1.{random.randint(1, 100)}',
            'confidence': random.uniform(0.7, 0.99),
            'description': '检测到横向移动攻击',
            'detector': random.choice(['lateral-auth.zeek', 'lateral-exec.zeek', 'ml_detector'])
        })
    
    if severity:
        alerts = [a for a in alerts if a['severity'] == severity]
    
    start = (page - 1) * page_size
    end = start + page_size
    
    return jsonify({
        'data': alerts[start:end],
        'total': len(alerts),
        'page': page,
        'page_size': page_size
    })

@app.route('/api/alerts/<int:alert_id>', methods=['GET'])
def get_alert_detail(alert_id):
    """获取告警详情"""
    return jsonify({
        'id': alert_id,
        'timestamp': datetime.now().strftime('%Y-%m-%d %H:%M:%S'),
        'severity': 'CRITICAL',
        'type': 'PTH攻击',
        'source': '192.168.1.100',
        'target': '10.0.1.50',
        'confidence': 0.95,
        'description': 'Pass-the-Hash攻击检测',
        'detector': 'lateral-auth.zeek',
        'evidence': 'NTLM Hash重用于3台主机:\n  - 10.0.1.50\n  - 10.0.1.51\n  - 10.0.1.52',
        'recommended_actions': [
            '立即隔离源IP 192.168.1.100',
            '检查受影响主机的进程列表',
            '重置相关账户密码',
            '检查域控制器日志'
        ]
    })

@app.route('/api/alerts/<int:alert_id>/handle', methods=['POST'])
def handle_alert(alert_id):
    """处置告警"""
    data = request.json
    action = data.get('action')
    
    return jsonify({
        'success': True,
        'message': f'告警 {alert_id} 已{action}'
    })

@app.route('/api/config', methods=['GET'])
def get_config():
    """获取配置"""
    try:
        with open('../config/detection.yaml', 'r', encoding='utf-8') as f:
            config = yaml.safe_load(f)
        return jsonify(config)
    except Exception as e:
        return jsonify({'error': str(e)}), 500

@app.route('/api/config', methods=['PUT'])
def update_config():
    """更新配置"""
    try:
        data = request.json
        with open('../config/detection.yaml', 'w', encoding='utf-8') as f:
            yaml.dump(data, f, allow_unicode=True)
        return jsonify({'success': True})
    except Exception as e:
        return jsonify({'error': str(e)}), 500

@app.route('/api/threat-intel/iocs', methods=['GET'])
def get_iocs():
    """获取IOC列表"""
    ioc_type = request.args.get('type', 'ip')
    
    return jsonify({
        'data': [
            {
                'id': 1,
                'value': '8.8.8.8',
                'type': 'ip',
                'source': 'abuse.ch',
                'category': 'C2服务器',
                'confidence': 0.95,
                'first_seen': '2025-12-20 10:00:00',
                'description': 'Cobalt Strike C2服务器'
            }
        ],
        'total': 1
    })

@app.route('/api/threat-intel/iocs', methods=['POST'])
def add_ioc():
    """添加IOC"""
    data = request.json
    
    return jsonify({
        'success': True,
        'id': random.randint(1000, 9999)
    })

@app.route('/api/threat-intel/iocs/<int:ioc_id>', methods=['DELETE'])
def delete_ioc(ioc_id):
    """删除IOC"""
    return jsonify({'success': True})

@app.route('/api/threat-intel/update', methods=['POST'])
def update_threat_intel():
    """更新威胁情报"""
    return jsonify({
        'success': True,
        'message': '威胁情报更新成功'
    })

@app.route('/api/topology/graph', methods=['GET'])
def get_topology_graph():
    """获取网络拓扑图"""
    return jsonify({
        'nodes': [
            {'id': '192.168.1.100', 'name': '192.168.1.100', 'type': 'attacker'},
            {'id': '10.0.1.50', 'name': '10.0.1.50', 'type': 'victim'},
            {'id': '10.0.1.51', 'name': '10.0.1.51', 'type': 'victim'},
        ],
        'edges': [
            {'source': '192.168.1.100', 'target': '10.0.1.50', 'protocol': 'SMB', 'count': 125},
            {'source': '192.168.1.100', 'target': '10.0.1.51', 'protocol': 'RDP', 'count': 45},
        ]
    })

@app.route('/api/topology/anomalies', methods=['GET'])
def get_topology_anomalies():
    """获取拓扑异常"""
    fanout = graph_analyzer.detect_anomalous_fanout(threshold=20)
    chains = graph_analyzer.find_multi_hop_chains(min_hops=3)
    
    return jsonify({
        'fanout': fanout,
        'chains': chains
    })

@app.route('/api/reports', methods=['GET'])
def get_reports():
    """获取报告列表"""
    return jsonify({
        'data': [
            {
                'id': 1,
                'title': '2025-12-22 安全检测日报',
                'type': '日报',
                'time_range': '2025-12-22 00:00 - 23:59',
                'alerts_count': 145,
                'created_at': '2025-12-22 23:30:00',
                'status': '已完成'
            }
        ],
        'total': 1
    })

@app.route('/api/reports/generate', methods=['POST'])
def generate_report():
    """生成报告"""
    data = request.json
    
    return jsonify({
        'success': True,
        'id': random.randint(1, 100),
        'message': '报告生成中'
    })

@app.route('/api/reports/<int:report_id>/download', methods=['GET'])
def download_report(report_id):
    """下载报告"""
    return jsonify({'url': f'/reports/{report_id}.html'})

@socketio.on('connect')
def handle_connect():
    """WebSocket连接"""
    print('Client connected')
    emit('connected', {'status': 'ok'})

@socketio.on('disconnect')
def handle_disconnect():
    """WebSocket断开"""
    print('Client disconnected')

def emit_new_alert():
    """发送新告警"""
    alert = {
        'id': random.randint(1000, 9999),
        'timestamp': datetime.now().strftime('%Y-%m-%d %H:%M:%S'),
        'severity': random.choice(['CRITICAL', 'HIGH', 'MEDIUM']),
        'type': random.choice(['PTH攻击', '横向扫描', 'PSExec']),
        'source': f'192.168.1.{random.randint(100, 200)}',
        'target': f'10.0.1.{random.randint(1, 100)}',
        'description': '检测到横向移动攻击'
    }
    socketio.emit('new_alert', alert)

if __name__ == '__main__':
    print('🚀 Cap Agent Backend API Server')
    print('📡 Listening on http://0.0.0.0:5000')
    print('🔌 WebSocket enabled')
    
    socketio.run(app, host='0.0.0.0', port=5000, debug=True, allow_unsafe_werkzeug=True)
