#!/usr/bin/env python3

import json
from datetime import datetime
from collections import defaultdict
import matplotlib
matplotlib.use('Agg')
import matplotlib.pyplot as plt
import matplotlib.dates as mdates

class AlertVisualizer:
    def __init__(self, output_dir='/var/log/zeek/reports'):
        self.output_dir = output_dir
        self.alert_data = defaultdict(list)
    
    def add_alert(self, alert):
        timestamp = alert.get('timestamp', datetime.now())
        severity = alert.get('severity', 'UNKNOWN')
        
        self.alert_data[severity].append({
            'timestamp': timestamp,
            'source': alert.get('source'),
            'type': alert.get('type'),
            'score': alert.get('score', 0)
        })
    
    def generate_timeline_chart(self, hours=24):
        fig, ax = plt.subplots(figsize=(12, 6))
        
        for severity, alerts in self.alert_data.items():
            timestamps = [a['timestamp'] for a in alerts]
            values = [1] * len(timestamps)
            
            ax.scatter(timestamps, values, label=severity, alpha=0.6, s=100)
        
        ax.set_xlabel('时间')
        ax.set_ylabel('告警')
        ax.set_title(f'最近{hours}小时告警时间线')
        ax.legend()
        ax.xaxis.set_major_formatter(mdates.DateFormatter('%H:%M'))
        plt.xticks(rotation=45)
        plt.tight_layout()
        
        output_file = f"{self.output_dir}/timeline_{datetime.now().strftime('%Y%m%d%H%M%S')}.png"
        plt.savefig(output_file)
        plt.close()
        
        return output_file
    
    def generate_severity_distribution(self):
        fig, ax = plt.subplots(figsize=(10, 6))
        
        severities = list(self.alert_data.keys())
        counts = [len(self.alert_data[s]) for s in severities]
        
        colors = {
            'CRITICAL': 'red',
            'HIGH': 'orange',
            'MEDIUM': 'yellow',
            'LOW': 'green',
            'INFO': 'blue'
        }
        
        bar_colors = [colors.get(s, 'gray') for s in severities]
        
        ax.bar(severities, counts, color=bar_colors)
        ax.set_xlabel('严重级别')
        ax.set_ylabel('数量')
        ax.set_title('告警严重级别分布')
        plt.tight_layout()
        
        output_file = f"{self.output_dir}/severity_{datetime.now().strftime('%Y%m%d%H%M%S')}.png"
        plt.savefig(output_file)
        plt.close()
        
        return output_file


class ReportGenerator:
    def __init__(self):
        self.report_template = """
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>横向移动检测报告</title>
    <style>
        body {{ font-family: Arial, sans-serif; margin: 20px; }}
        h1 {{ color: #333; }}
        h2 {{ color: #666; border-bottom: 2px solid #ddd; padding-bottom: 5px; }}
        table {{ border-collapse: collapse; width: 100%; margin: 20px 0; }}
        th, td {{ border: 1px solid #ddd; padding: 12px; text-align: left; }}
        th {{ background-color: #4CAF50; color: white; }}
        tr:nth-child(even) {{ background-color: #f2f2f2; }}
        .critical {{ background-color: #ff4444; color: white; }}
        .high {{ background-color: #ff8800; color: white; }}
        .medium {{ background-color: #ffbb33; }}
        .low {{ background-color: #00C851; }}
        .summary {{ background-color: #f5f5f5; padding: 15px; border-radius: 5px; margin: 20px 0; }}
        .metric {{ display: inline-block; margin: 10px 20px; }}
        .metric-value {{ font-size: 36px; font-weight: bold; color: #333; }}
        .metric-label {{ font-size: 14px; color: #666; }}
    </style>
</head>
<body>
    <h1>Zeek横向移动检测报告</h1>
    
    <div class="summary">
        <h2>执行摘要</h2>
        <p>报告生成时间: {timestamp}</p>
        <p>监控时段: {time_range}</p>
        
        <div class="metric">
            <div class="metric-value">{total_alerts}</div>
            <div class="metric-label">总告警数</div>
        </div>
        
        <div class="metric">
            <div class="metric-value">{critical_count}</div>
            <div class="metric-label">严重告警</div>
        </div>
        
        <div class="metric">
            <div class="metric-value">{unique_ips}</div>
            <div class="metric-label">涉及IP</div>
        </div>
        
        <div class="metric">
            <div class="metric-value">{apt_campaigns}</div>
            <div class="metric-label">APT活动</div>
        </div>
    </div>
    
    <h2>严重告警详情</h2>
    <table>
        <tr>
            <th>时间</th>
            <th>严重级别</th>
            <th>源IP</th>
            <th>目标IP</th>
            <th>攻击类型</th>
            <th>置信度</th>
            <th>描述</th>
        </tr>
        {critical_alerts_rows}
    </table>
    
    <h2>攻击类型统计</h2>
    <table>
        <tr>
            <th>攻击类型</th>
            <th>数量</th>
            <th>占比</th>
        </tr>
        {attack_type_rows}
    </table>
    
    <h2>TOP 10 攻击源</h2>
    <table>
        <tr>
            <th>排名</th>
            <th>IP地址</th>
            <th>告警数</th>
            <th>最高严重级别</th>
            <th>攻击类型</th>
        </tr>
        {top_attackers_rows}
    </table>
    
    <h2>检测能力统计</h2>
    <table>
        <tr>
            <th>检测模块</th>
            <th>触发次数</th>
            <th>准确率估计</th>
        </tr>
        {detector_stats_rows}
    </table>
    
    <h2>推荐行动</h2>
    <ul>
        {recommended_actions}
    </ul>
    
    <div class="summary">
        <h2>系统状态</h2>
        <p>Zeek运行状态: {zeek_status}</p>
        <p>处理流量: {traffic_processed}</p>
        <p>检测模块: {active_modules}</p>
        <p>威胁情报: {threat_intel_status}</p>
    </div>
</body>
</html>
"""
    
    def generate_html_report(self, alerts, statistics, output_file):
        critical_alerts = [a for a in alerts if a.get('severity') == 'CRITICAL']
        
        critical_alerts_rows = ""
        for alert in critical_alerts[:20]:
            critical_alerts_rows += f"""
                <tr class="critical">
                    <td>{alert.get('timestamp', 'N/A')}</td>
                    <td>{alert.get('severity', 'N/A')}</td>
                    <td>{alert.get('source', 'N/A')}</td>
                    <td>{alert.get('destination', 'N/A')}</td>
                    <td>{alert.get('type', 'N/A')}</td>
                    <td>{alert.get('confidence', 0):.2%}</td>
                    <td>{alert.get('description', 'N/A')}</td>
                </tr>
            """
        
        attack_types = defaultdict(int)
        for alert in alerts:
            attack_types[alert.get('type', 'UNKNOWN')] += 1
        
        total_alerts = len(alerts)
        attack_type_rows = ""
        for attack_type, count in sorted(attack_types.items(), key=lambda x: x[1], reverse=True):
            percentage = (count / total_alerts * 100) if total_alerts > 0 else 0
            attack_type_rows += f"""
                <tr>
                    <td>{attack_type}</td>
                    <td>{count}</td>
                    <td>{percentage:.1f}%</td>
                </tr>
            """
        
        attacker_stats = defaultdict(lambda: {'count': 0, 'max_severity': 'LOW', 'types': set()})
        for alert in alerts:
            src = alert.get('source', 'unknown')
            attacker_stats[src]['count'] += 1
            attacker_stats[src]['types'].add(alert.get('type', 'UNKNOWN'))
            
            severity = alert.get('severity', 'LOW')
            severity_levels = {'CRITICAL': 4, 'HIGH': 3, 'MEDIUM': 2, 'LOW': 1, 'INFO': 0}
            if severity_levels.get(severity, 0) > severity_levels.get(attacker_stats[src]['max_severity'], 0):
                attacker_stats[src]['max_severity'] = severity
        
        top_attackers = sorted(attacker_stats.items(), key=lambda x: x[1]['count'], reverse=True)[:10]
        top_attackers_rows = ""
        for rank, (ip, stats) in enumerate(top_attackers, 1):
            top_attackers_rows += f"""
                <tr>
                    <td>{rank}</td>
                    <td>{ip}</td>
                    <td>{stats['count']}</td>
                    <td>{stats['max_severity']}</td>
                    <td>{', '.join(list(stats['types'])[:3])}</td>
                </tr>
            """
        
        detector_stats_rows = """
            <tr><td>Zeek横向扫描</td><td>{scan_count}</td><td>90%</td></tr>
            <tr><td>Zeek认证分析</td><td>{auth_count}</td><td>90%</td></tr>
            <tr><td>Zeek远程执行</td><td>{exec_count}</td><td>85%</td></tr>
            <tr><td>深度包检测</td><td>{dpi_count}</td><td>80%</td></tr>
            <tr><td>加密流量分析</td><td>{encrypted_count}</td><td>75%</td></tr>
            <tr><td>ML异常检测</td><td>{ml_count}</td><td>85%</td></tr>
            <tr><td>图分析</td><td>{graph_count}</td><td>80%</td></tr>
            <tr><td>威胁情报</td><td>{intel_count}</td><td>95%</td></tr>
        """.format(
            scan_count=statistics.get('zeek_scan', 0),
            auth_count=statistics.get('zeek_auth', 0),
            exec_count=statistics.get('zeek_exec', 0),
            dpi_count=statistics.get('zeek_dpi', 0),
            encrypted_count=statistics.get('zeek_encrypted', 0),
            ml_count=statistics.get('ml_anomaly', 0),
            graph_count=statistics.get('graph_analysis', 0),
            intel_count=statistics.get('threat_intel', 0)
        )
        
        recommended_actions = ""
        if critical_alerts:
            recommended_actions += "<li>立即隔离以下高风险IP: " + ", ".join([a.get('source', 'N/A') for a in critical_alerts[:5]]) + "</li>"
            recommended_actions += "<li>对受影响主机进行深度取证分析</li>"
            recommended_actions += "<li>检查域控制器和关键服务器的完整性</li>"
        recommended_actions += "<li>更新威胁情报库</li>"
        recommended_actions += "<li>审查白名单配置，确保无误报</li>"
        recommended_actions += "<li>定期审查告警趋势，优化检测规则</li>"
        
        unique_ips = len(set(a.get('source') for a in alerts))
        apt_campaigns = statistics.get('apt_campaigns', 0)
        
        html_content = self.report_template.format(
            timestamp=datetime.now().strftime('%Y-%m-%d %H:%M:%S'),
            time_range=statistics.get('time_range', '最近24小时'),
            total_alerts=total_alerts,
            critical_count=len(critical_alerts),
            unique_ips=unique_ips,
            apt_campaigns=apt_campaigns,
            critical_alerts_rows=critical_alerts_rows,
            attack_type_rows=attack_type_rows,
            top_attackers_rows=top_attackers_rows,
            detector_stats_rows=detector_stats_rows,
            recommended_actions=recommended_actions,
            zeek_status=statistics.get('zeek_status', 'Running'),
            traffic_processed=statistics.get('traffic_processed', 'N/A'),
            active_modules=statistics.get('active_modules', '11/11'),
            threat_intel_status=statistics.get('threat_intel_status', 'Active')
        )
        
        with open(output_file, 'w', encoding='utf-8') as f:
            f.write(html_content)
        
        print(f"HTML报告已生成: {output_file}")
        return output_file


if __name__ == '__main__':
    print("报告生成模块测试\n")
    
    test_alerts = [
        {
            'timestamp': '2025-12-22 10:30:00',
            'severity': 'CRITICAL',
            'source': '192.168.1.100',
            'destination': '10.0.1.50',
            'type': 'PTH_ATTACK',
            'confidence': 0.95,
            'description': 'Pass-the-Hash攻击'
        },
        {
            'timestamp': '2025-12-22 10:35:00',
            'severity': 'HIGH',
            'source': '192.168.1.100',
            'destination': '10.0.1.51',
            'type': 'LATERAL_SCAN',
            'confidence': 0.90,
            'description': '横向扫描'
        }
    ]
    
    test_statistics = {
        'time_range': '最近24小时',
        'apt_campaigns': 1,
        'zeek_scan': 15,
        'zeek_auth': 8,
        'zeek_exec': 5,
        'zeek_dpi': 3,
        'zeek_encrypted': 12,
        'ml_anomaly': 10,
        'graph_analysis': 7,
        'threat_intel': 2,
        'zeek_status': 'Running',
        'traffic_processed': '250GB',
        'active_modules': '11/11',
        'threat_intel_status': 'Active'
    }
    
    generator = ReportGenerator()
    output_file = '/tmp/test_report.html'
    generator.generate_html_report(test_alerts, test_statistics, output_file)
    
    print(f"测试报告已生成: {output_file}")
