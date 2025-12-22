#!/usr/bin/env python3

import networkx as nx
import json
from collections import defaultdict
from datetime import datetime, timedelta

class LateralGraphAnalyzer:
    def __init__(self):
        self.G = nx.DiGraph()
        self.normal_paths = set()
        self.communication_history = defaultdict(list)
        self.graph_file = '/var/log/zeek/graph_state.json'
    
    def add_connection(self, src, dst, protocol, timestamp, metadata=None):
        if not self.G.has_edge(src, dst):
            self.G.add_edge(src, dst, 
                          weight=1,
                          protocols=set([protocol]),
                          first_seen=timestamp,
                          last_seen=timestamp,
                          count=1,
                          metadata=metadata or {})
        else:
            edge = self.G[src][dst]
            edge['weight'] += 1
            edge['count'] += 1
            edge['protocols'].add(protocol)
            edge['last_seen'] = timestamp
    
    def detect_anomalous_fanout(self, threshold=20):
        anomalies = []
        
        for node in self.G.nodes():
            out_degree = self.G.out_degree(node)
            
            if out_degree > threshold:
                targets = list(self.G.successors(node))
                
                anomalies.append({
                    'type': 'ABNORMAL_FANOUT',
                    'node': node,
                    'target_count': out_degree,
                    'targets': targets[:10],
                    'score': min(out_degree / threshold, 1.0),
                    'severity': 'HIGH' if out_degree > threshold * 2 else 'MEDIUM'
                })
        
        return anomalies
    
    def find_multi_hop_chains(self, min_hops=3, max_hops=6):
        chains = []
        
        for source in self.G.nodes():
            try:
                paths = nx.single_source_shortest_path(self.G, source, cutoff=max_hops)
                
                for target, path in paths.items():
                    if len(path) >= min_hops:
                        if self._is_abnormal_chain(path):
                            chains.append({
                                'path': path,
                                'length': len(path),
                                'score': self._calculate_chain_score(path)
                            })
            except Exception:
                continue
        
        return chains
    
    def _is_abnormal_chain(self, path):
        if len(path) < 3:
            return False
        
        internal_hops = 0
        for node in path[1:-1]:
            if self._is_internal_node(node):
                internal_hops += 1
        
        return internal_hops >= 2
    
    def _calculate_chain_score(self, path):
        score = len(path) * 10
        
        for i in range(len(path) - 1):
            src = path[i]
            dst = path[i + 1]
            
            if self.G.has_edge(src, dst):
                edge = self.G[src][dst]
                
                if edge['count'] == 1:
                    score += 5
                
                if 'SMB' in edge['protocols'] or 'RDP' in edge['protocols']:
                    score += 10
        
        return score
    
    def detect_rare_communications(self, rarity_threshold=0.95):
        anomalies = []
        
        total_edges = self.G.number_of_edges()
        
        for src, dst, data in self.G.edges(data=True):
            edge_tuple = (src, dst)
            
            if edge_tuple not in self.normal_paths:
                frequency = data['count']
                rarity = 1.0 - (frequency / max(total_edges, 1))
                
                if rarity > rarity_threshold:
                    anomalies.append({
                        'type': 'RARE_COMMUNICATION',
                        'edge': (src, dst),
                        'protocols': list(data['protocols']),
                        'count': frequency,
                        'rarity': rarity,
                        'score': rarity
                    })
        
        return anomalies
    
    def detect_pivot_points(self):
        pivots = []
        
        for node in self.G.nodes():
            in_degree = self.G.in_degree(node)
            out_degree = self.G.out_degree(node)
            
            if in_degree >= 1 and out_degree >= 3:
                betweenness = nx.betweenness_centrality(self.G).get(node, 0)
                
                if betweenness > 0.1:
                    pivots.append({
                        'node': node,
                        'in_degree': in_degree,
                        'out_degree': out_degree,
                        'betweenness': betweenness,
                        'type': 'PIVOT_POINT',
                        'severity': 'CRITICAL' if out_degree > 5 else 'HIGH'
                    })
        
        return pivots
    
    def detect_circular_paths(self):
        try:
            cycles = list(nx.simple_cycles(self.G))
            
            anomalous_cycles = []
            for cycle in cycles:
                if len(cycle) >= 3:
                    anomalous_cycles.append({
                        'type': 'CIRCULAR_PATH',
                        'cycle': cycle,
                        'length': len(cycle),
                        'score': len(cycle) * 5
                    })
            
            return anomalous_cycles
        except Exception:
            return []
    
    def get_attack_path_summary(self, attacker_ip):
        if attacker_ip not in self.G:
            return None
        
        summary = {
            'attacker': attacker_ip,
            'direct_targets': list(self.G.successors(attacker_ip)),
            'total_targets': len(list(nx.descendants(self.G, attacker_ip))),
            'max_hop_depth': 0,
            'protocols_used': set(),
            'pivot_hosts': []
        }
        
        for target in nx.descendants(self.G, attacker_ip):
            try:
                path = nx.shortest_path(self.G, attacker_ip, target)
                if len(path) > summary['max_hop_depth']:
                    summary['max_hop_depth'] = len(path)
            except Exception:
                pass
        
        for src, dst, data in self.G.edges(data=True):
            if src == attacker_ip or dst == attacker_ip:
                summary['protocols_used'].update(data['protocols'])
        
        summary['protocols_used'] = list(summary['protocols_used'])
        
        return summary
    
    def _is_internal_node(self, node):
        return True
    
    def export_graph(self):
        data = {
            'nodes': list(self.G.nodes()),
            'edges': [
                {
                    'source': src,
                    'target': dst,
                    'protocols': list(data['protocols']),
                    'count': data['count'],
                    'first_seen': str(data['first_seen']),
                    'last_seen': str(data['last_seen'])
                }
                for src, dst, data in self.G.edges(data=True)
            ],
            'timestamp': datetime.now().isoformat()
        }
        
        return json.dumps(data, indent=2, default=str)
    
    def save_graph(self):
        try:
            with open(self.graph_file, 'w') as f:
                f.write(self.export_graph())
        except Exception as e:
            print(f"保存图状态失败: {e}")
    
    def load_graph(self):
        try:
            with open(self.graph_file, 'r') as f:
                data = json.load(f)
            
            self.G.clear()
            
            for edge in data['edges']:
                self.G.add_edge(
                    edge['source'],
                    edge['target'],
                    protocols=set(edge['protocols']),
                    count=edge['count'],
                    first_seen=datetime.fromisoformat(edge['first_seen']),
                    last_seen=datetime.fromisoformat(edge['last_seen']),
                    weight=edge['count']
                )
            
            print(f"图状态加载成功: {self.G.number_of_nodes()} 节点, {self.G.number_of_edges()} 边")
        except Exception as e:
            print(f"加载图状态失败: {e}")


if __name__ == '__main__':
    print("图分析引擎示例")
    
    analyzer = LateralGraphAnalyzer()
    
    analyzer.add_connection('10.0.1.100', '10.0.1.101', 'SMB', datetime.now())
    analyzer.add_connection('10.0.1.100', '10.0.1.102', 'RDP', datetime.now())
    analyzer.add_connection('10.0.1.101', '10.0.1.103', 'SMB', datetime.now())
    
    print("\n检测异常扇出:")
    fanout = analyzer.detect_anomalous_fanout(threshold=2)
    for item in fanout:
        print(f"  - {item}")
    
    print("\n检测多跳链路:")
    chains = analyzer.find_multi_hop_chains(min_hops=2)
    for chain in chains:
        print(f"  - {chain}")
    
    print("\n导出图结构:")
    print(analyzer.export_graph())
