#!/usr/bin/env python3

import requests
import json
import hashlib
from datetime import datetime, timedelta
from collections import defaultdict

class ThreatIntelligence:
    def __init__(self, config_file='../config/detection.yaml'):
        self.ioc_cache = {
            'ip': {},
            'domain': {},
            'hash': {},
            'url': {}
        }
        
        self.malicious_ips = set()
        self.malicious_domains = set()
        self.malicious_hashes = set()
        
        self.cache_ttl = timedelta(hours=24)
        self.cache_file = '/var/log/zeek/threat_intel_cache.json'
        
        self.load_cache()
        self.load_builtin_iocs()
    
    def load_builtin_iocs(self):
        self.malicious_ja3 = {
            "a0e9f5d64349fb13191bc781f81f42e1": {
                "name": "Metasploit",
                "type": "C2_Framework",
                "severity": "CRITICAL"
            },
            "6734f37431670b3ab4292b8f60f29984": {
                "name": "Trickbot",
                "type": "Banking_Trojan",
                "severity": "CRITICAL"
            },
            "72a589da586844d7f0818ce684948eea": {
                "name": "Dridex",
                "type": "Banking_Trojan",
                "severity": "CRITICAL"
            },
            "51c64c77e60f3980eea90869b68c58a8": {
                "name": "Cobalt Strike",
                "type": "C2_Framework",
                "severity": "CRITICAL"
            }
        }
        
        self.malicious_user_agents = {
            "python-requests": "Automated_Script",
            "curl": "Command_Line_Tool",
            "Metasploit": "Exploitation_Framework",
            "Nmap": "Network_Scanner",
            "sqlmap": "SQL_Injection_Tool",
            "masscan": "Port_Scanner"
        }
        
        self.suspicious_domain_patterns = [
            r'[a-z0-9]{20,}\.com$',
            r'[a-z0-9]{15,}\.(ru|cn|tk)$',
            r'.*-[0-9]{8,}\..*',
            r'.*\.(bit|onion)$'
        ]
        
        self.c2_port_signatures = {
            4444: "Metasploit_Default",
            5555: "Common_Backdoor",
            6666: "Common_Backdoor",
            7777: "Common_Backdoor",
            8888: "Common_Proxy",
            9999: "Common_Backdoor",
            1337: "Leet_Port",
            31337: "Back_Orifice"
        }
    
    def check_ip(self, ip_address):
        if ip_address in self.ioc_cache['ip']:
            cached = self.ioc_cache['ip'][ip_address]
            if datetime.now() - cached['timestamp'] < self.cache_ttl:
                return cached['result']
        
        result = {
            'ip': ip_address,
            'is_malicious': False,
            'confidence': 0.0,
            'sources': [],
            'categories': [],
            'details': {}
        }
        
        if ip_address in self.malicious_ips:
            result['is_malicious'] = True
            result['confidence'] = 0.95
            result['sources'].append('Local_Blacklist')
        
        self.ioc_cache['ip'][ip_address] = {
            'timestamp': datetime.now(),
            'result': result
        }
        
        return result
    
    def check_domain(self, domain):
        if domain in self.ioc_cache['domain']:
            cached = self.ioc_cache['domain'][domain]
            if datetime.now() - cached['timestamp'] < self.cache_ttl:
                return cached['result']
        
        result = {
            'domain': domain,
            'is_malicious': False,
            'confidence': 0.0,
            'sources': [],
            'categories': [],
            'details': {}
        }
        
        if domain in self.malicious_domains:
            result['is_malicious'] = True
            result['confidence'] = 0.95
            result['sources'].append('Local_Blacklist')
            result['categories'].append('Known_Malicious')
        
        import re
        for pattern in self.suspicious_domain_patterns:
            if re.search(pattern, domain):
                result['is_malicious'] = True
                result['confidence'] = max(result['confidence'], 0.7)
                result['sources'].append('Pattern_Match')
                result['categories'].append('Suspicious_Pattern')
                result['details']['pattern'] = pattern
                break
        
        self.ioc_cache['domain'][domain] = {
            'timestamp': datetime.now(),
            'result': result
        }
        
        return result
    
    def check_file_hash(self, file_hash, hash_type='md5'):
        if file_hash in self.ioc_cache['hash']:
            cached = self.ioc_cache['hash'][file_hash]
            if datetime.now() - cached['timestamp'] < self.cache_ttl:
                return cached['result']
        
        result = {
            'hash': file_hash,
            'hash_type': hash_type,
            'is_malicious': False,
            'confidence': 0.0,
            'sources': [],
            'malware_family': None,
            'details': {}
        }
        
        if file_hash in self.malicious_hashes:
            result['is_malicious'] = True
            result['confidence'] = 0.99
            result['sources'].append('Local_Blacklist')
        
        self.ioc_cache['hash'][file_hash] = {
            'timestamp': datetime.now(),
            'result': result
        }
        
        return result
    
    def check_ja3_fingerprint(self, ja3_hash):
        if ja3_hash in self.malicious_ja3:
            info = self.malicious_ja3[ja3_hash]
            return {
                'ja3': ja3_hash,
                'is_malicious': True,
                'confidence': 0.95,
                'tool_name': info['name'],
                'tool_type': info['type'],
                'severity': info['severity']
            }
        
        return {
            'ja3': ja3_hash,
            'is_malicious': False,
            'confidence': 0.0
        }
    
    def check_user_agent(self, user_agent):
        for pattern, category in self.malicious_user_agents.items():
            if pattern.lower() in user_agent.lower():
                return {
                    'user_agent': user_agent,
                    'is_suspicious': True,
                    'confidence': 0.8,
                    'category': category,
                    'matched_pattern': pattern
                }
        
        return {
            'user_agent': user_agent,
            'is_suspicious': False,
            'confidence': 0.0
        }
    
    def check_port(self, port_number):
        if port_number in self.c2_port_signatures:
            return {
                'port': port_number,
                'is_suspicious': True,
                'signature': self.c2_port_signatures[port_number],
                'confidence': 0.7
            }
        
        return {
            'port': port_number,
            'is_suspicious': False,
            'confidence': 0.0
        }
    
    def enrich_event(self, event):
        enrichment = {
            'threat_intel': {},
            'risk_score': 0
        }
        
        if 'src_ip' in event:
            ip_intel = self.check_ip(event['src_ip'])
            if ip_intel['is_malicious']:
                enrichment['threat_intel']['src_ip'] = ip_intel
                enrichment['risk_score'] += 50
        
        if 'dst_ip' in event:
            ip_intel = self.check_ip(event['dst_ip'])
            if ip_intel['is_malicious']:
                enrichment['threat_intel']['dst_ip'] = ip_intel
                enrichment['risk_score'] += 30
        
        if 'domain' in event:
            domain_intel = self.check_domain(event['domain'])
            if domain_intel['is_malicious']:
                enrichment['threat_intel']['domain'] = domain_intel
                enrichment['risk_score'] += 40
        
        if 'file_hash' in event:
            hash_intel = self.check_file_hash(event['file_hash'])
            if hash_intel['is_malicious']:
                enrichment['threat_intel']['file_hash'] = hash_intel
                enrichment['risk_score'] += 60
        
        if 'ja3' in event:
            ja3_intel = self.check_ja3_fingerprint(event['ja3'])
            if ja3_intel['is_malicious']:
                enrichment['threat_intel']['ja3'] = ja3_intel
                enrichment['risk_score'] += 45
        
        if 'user_agent' in event:
            ua_intel = self.check_user_agent(event['user_agent'])
            if ua_intel['is_suspicious']:
                enrichment['threat_intel']['user_agent'] = ua_intel
                enrichment['risk_score'] += 20
        
        if 'dst_port' in event:
            port_intel = self.check_port(event['dst_port'])
            if port_intel['is_suspicious']:
                enrichment['threat_intel']['port'] = port_intel
                enrichment['risk_score'] += 15
        
        return enrichment
    
    def add_ioc(self, ioc_type, value, metadata=None):
        if ioc_type == 'ip':
            self.malicious_ips.add(value)
        elif ioc_type == 'domain':
            self.malicious_domains.add(value)
        elif ioc_type == 'hash':
            self.malicious_hashes.add(value)
        
        print(f"添加IOC: {ioc_type} = {value}")
    
    def load_ioc_feed(self, feed_url):
        try:
            response = requests.get(feed_url, timeout=10)
            if response.status_code == 200:
                data = response.json()
                
                for item in data.get('ips', []):
                    self.add_ioc('ip', item)
                
                for item in data.get('domains', []):
                    self.add_ioc('domain', item)
                
                for item in data.get('hashes', []):
                    self.add_ioc('hash', item)
                
                print(f"成功加载IOC feed: {feed_url}")
            else:
                print(f"加载IOC feed失败: HTTP {response.status_code}")
        except Exception as e:
            print(f"加载IOC feed异常: {e}")
    
    def save_cache(self):
        cache_data = {
            'malicious_ips': list(self.malicious_ips),
            'malicious_domains': list(self.malicious_domains),
            'malicious_hashes': list(self.malicious_hashes),
            'timestamp': datetime.now().isoformat()
        }
        
        try:
            with open(self.cache_file, 'w') as f:
                json.dump(cache_data, f, indent=2)
        except Exception as e:
            print(f"保存缓存失败: {e}")
    
    def load_cache(self):
        try:
            with open(self.cache_file, 'r') as f:
                cache_data = json.load(f)
            
            self.malicious_ips = set(cache_data.get('malicious_ips', []))
            self.malicious_domains = set(cache_data.get('malicious_domains', []))
            self.malicious_hashes = set(cache_data.get('malicious_hashes', []))
            
            print(f"威胁情报缓存加载成功: {len(self.malicious_ips)} IPs, "
                  f"{len(self.malicious_domains)} domains, {len(self.malicious_hashes)} hashes")
        except FileNotFoundError:
            print("威胁情报缓存文件不存在，使用默认配置")
        except Exception as e:
            print(f"加载缓存失败: {e}")


if __name__ == '__main__':
    print("威胁情报集成模块")
    
    intel = ThreatIntelligence()
    
    print("\n测试IP检查:")
    result = intel.check_ip("192.168.1.100")
    print(json.dumps(result, indent=2))
    
    print("\n测试JA3指纹检查:")
    result = intel.check_ja3_fingerprint("a0e9f5d64349fb13191bc781f81f42e1")
    print(json.dumps(result, indent=2))
    
    print("\n测试事件增强:")
    event = {
        'src_ip': '10.0.1.100',
        'dst_ip': '8.8.8.8',
        'domain': 'example.com',
        'ja3': 'a0e9f5d64349fb13191bc781f81f42e1'
    }
    enrichment = intel.enrich_event(event)
    print(json.dumps(enrichment, indent=2))
