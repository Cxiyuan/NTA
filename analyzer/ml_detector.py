#!/usr/bin/env python3

import json
import yaml
import numpy as np
from collections import defaultdict
from datetime import datetime, timedelta
from sklearn.ensemble import IsolationForest
from sklearn.preprocessing import StandardScaler
import pickle
import os

class MLAnomalyDetector:
    def __init__(self, config_file='../config/detection.yaml'):
        self.model = IsolationForest(
            contamination=0.01,
            random_state=42,
            n_estimators=100
        )
        self.scaler = StandardScaler()
        self.is_trained = False
        self.model_file = '/var/log/zeek/ml_model.pkl'
        
        self.load_model()
    
    def extract_features(self, connection_data):
        features = []
        
        conn_rate = connection_data.get('connection_rate', 0)
        target_count = connection_data.get('target_count', 0)
        port_diversity = connection_data.get('port_diversity', 0)
        failed_auth_ratio = connection_data.get('failed_auth_ratio', 0)
        avg_packet_size = connection_data.get('avg_packet_size', 0)
        session_duration = connection_data.get('session_duration', 0)
        upload_download_ratio = connection_data.get('upload_download_ratio', 0)
        inter_arrival_variance = connection_data.get('inter_arrival_variance', 0)
        
        features = [
            conn_rate,
            target_count,
            port_diversity,
            failed_auth_ratio,
            avg_packet_size,
            session_duration,
            upload_download_ratio,
            inter_arrival_variance
        ]
        
        return np.array(features).reshape(1, -1)
    
    def train(self, historical_logs):
        print(f"训练ML模型，样本数: {len(historical_logs)}")
        
        X = []
        for log in historical_logs:
            features = self.extract_features(log)
            X.append(features[0])
        
        X = np.array(X)
        X_scaled = self.scaler.fit_transform(X)
        
        self.model.fit(X_scaled)
        self.is_trained = True
        
        self.save_model()
        print("模型训练完成")
    
    def predict(self, connection_data):
        if not self.is_trained:
            return {'anomaly': False, 'score': 0.0, 'confidence': 0.0}
        
        features = self.extract_features(connection_data)
        features_scaled = self.scaler.transform(features)
        
        prediction = self.model.predict(features_scaled)[0]
        score = self.model.decision_function(features_scaled)[0]
        
        is_anomaly = prediction == -1
        confidence = abs(score)
        
        return {
            'anomaly': is_anomaly,
            'score': float(score),
            'confidence': float(confidence)
        }
    
    def save_model(self):
        model_data = {
            'model': self.model,
            'scaler': self.scaler,
            'is_trained': self.is_trained
        }
        
        with open(self.model_file, 'wb') as f:
            pickle.dump(model_data, f)
    
    def load_model(self):
        if os.path.exists(self.model_file):
            try:
                with open(self.model_file, 'rb') as f:
                    model_data = pickle.load(f)
                
                self.model = model_data['model']
                self.scaler = model_data['scaler']
                self.is_trained = model_data['is_trained']
                
                print("ML模型加载成功")
            except Exception as e:
                print(f"加载模型失败: {e}")


class BaselineLearner:
    def __init__(self):
        self.ip_baselines = {}
        self.global_baseline = None
        self.baseline_file = '/var/log/zeek/baseline.json'
        
        self.load_baseline()
    
    def calculate_baseline(self, logs, time_window='1h'):
        baseline = {
            'connection_rate': [],
            'target_count': [],
            'port_diversity': [],
            'avg_packet_size': [],
            'session_duration': []
        }
        
        for log in logs:
            for key in baseline.keys():
                if key in log:
                    baseline[key].append(log[key])
        
        result = {}
        for key, values in baseline.items():
            if values:
                result[key] = {
                    'mean': np.mean(values),
                    'std': np.std(values),
                    'min': np.min(values),
                    'max': np.max(values),
                    'p95': np.percentile(values, 95)
                }
        
        return result
    
    def is_anomalous(self, event, ip_address):
        if ip_address not in self.ip_baselines:
            return False, 0.0
        
        baseline = self.ip_baselines[ip_address]
        anomaly_score = 0.0
        
        for metric, value in event.items():
            if metric in baseline:
                b = baseline[metric]
                
                if b['std'] > 0:
                    z_score = abs((value - b['mean']) / b['std'])
                    if z_score > 3:
                        anomaly_score += z_score
        
        is_anomaly = anomaly_score > 10.0
        return is_anomaly, anomaly_score
    
    def update_baseline(self, ip_address, event):
        if ip_address not in self.ip_baselines:
            self.ip_baselines[ip_address] = {}
        
        for metric, value in event.items():
            if metric not in self.ip_baselines[ip_address]:
                self.ip_baselines[ip_address][metric] = {
                    'mean': value,
                    'std': 0.0,
                    'count': 1
                }
            else:
                b = self.ip_baselines[ip_address][metric]
                n = b['count']
                old_mean = b['mean']
                
                b['mean'] = (old_mean * n + value) / (n + 1)
                b['count'] += 1
    
    def save_baseline(self):
        try:
            with open(self.baseline_file, 'w') as f:
                json.dump(self.ip_baselines, f, indent=2, default=str)
        except Exception as e:
            print(f"保存基线失败: {e}")
    
    def load_baseline(self):
        if os.path.exists(self.baseline_file):
            try:
                with open(self.baseline_file, 'r') as f:
                    self.ip_baselines = json.load(f)
                print(f"基线加载成功，包含{len(self.ip_baselines)}个IP")
            except Exception as e:
                print(f"加载基线失败: {e}")


class CircadianAnalyzer:
    def __init__(self):
        self.hourly_baselines = {}
        
        for hour in range(24):
            self.hourly_baselines[hour] = {
                'avg_connections': 0.0,
                'std_connections': 0.0,
                'common_activities': set(),
                'sample_count': 0
            }
    
    def is_anomalous_time(self, event):
        timestamp = event.get('timestamp', datetime.now())
        hour = timestamp.hour
        
        baseline = self.hourly_baselines[hour]
        
        conn_count = event.get('connection_count', 0)
        
        if baseline['sample_count'] < 10:
            return False, 0.0
        
        if baseline['std_connections'] > 0:
            z_score = abs((conn_count - baseline['avg_connections']) / 
                         baseline['std_connections'])
        else:
            z_score = 0.0
        
        if 2 <= hour <= 6:
            threshold = 2.0
        elif 9 <= hour <= 17:
            threshold = 5.0
        else:
            threshold = 3.0
        
        is_anomaly = z_score > threshold
        
        return is_anomaly, z_score
    
    def update_hourly_baseline(self, event):
        timestamp = event.get('timestamp', datetime.now())
        hour = timestamp.hour
        
        baseline = self.hourly_baselines[hour]
        conn_count = event.get('connection_count', 0)
        
        n = baseline['sample_count']
        old_avg = baseline['avg_connections']
        
        baseline['avg_connections'] = (old_avg * n + conn_count) / (n + 1)
        baseline['sample_count'] += 1


if __name__ == '__main__':
    print("ML异常检测引擎")
    print("使用示例:")
    print("  detector = MLAnomalyDetector()")
    print("  result = detector.predict(connection_data)")
