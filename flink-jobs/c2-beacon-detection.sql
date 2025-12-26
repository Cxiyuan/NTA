-- Flink SQL作业：C2 Beacon检测
-- 通过滑动窗口检测规律性信标通信

CREATE CATALOG nta_catalog WITH (
  'type' = 'jdbc',
  'default-database' = 'nta',
  'username' = 'nta',
  'password' = 'nta_password',
  'base-url' = 'jdbc:postgresql://postgres:5432/'
);

USE CATALOG nta_catalog;

-- 创建Kafka源表：连接日志
CREATE TABLE zeek_conn (
  uid STRING,
  ts TIMESTAMP(3),
  orig_h STRING,
  orig_p INT,
  resp_h STRING,
  resp_p INT,
  proto STRING,
  service STRING,
  duration DOUBLE,
  orig_bytes BIGINT,
  resp_bytes BIGINT,
  conn_state STRING,
  WATERMARK FOR ts AS ts - INTERVAL '10' SECOND
) WITH (
  'connector' = 'kafka',
  'topic' = 'zeek-conn',
  'properties.bootstrap.servers' = 'kafka:9092',
  'properties.group.id' = 'flink-c2-detector',
  'scan.startup.mode' = 'latest-offset',
  'format' = 'json',
  'json.fail-on-missing-field' = 'false',
  'json.ignore-parse-errors' = 'true'
);

-- 创建告警输出表
CREATE TABLE alerts (
  alert_id STRING,
  alert_type STRING,
  severity STRING,
  src_ip STRING,
  dst_ip STRING,
  description STRING,
  confidence DOUBLE,
  timestamp TIMESTAMP(3),
  PRIMARY KEY (alert_id) NOT ENFORCED
) WITH (
  'connector' = 'jdbc',
  'url' = 'jdbc:postgresql://postgres:5432/nta',
  'table-name' = 'alerts',
  'username' = 'nta',
  'password' = 'nta_password',
  'driver' = 'org.postgresql.Driver'
);

-- 窗口聚合：10分钟滑动窗口检测C2 Beacon
INSERT INTO alerts
SELECT 
  CONCAT('c2-', orig_h, '-', resp_h, '-', CAST(window_end AS STRING)) AS alert_id,
  'c2_beacon' AS alert_type,
  CASE 
    WHEN interval_variance < 0.5 AND avg_interval > 60 THEN 'critical'
    WHEN interval_variance < 1.0 AND avg_interval > 30 THEN 'high'
    ELSE 'medium'
  END AS severity,
  orig_h AS src_ip,
  resp_h AS dst_ip,
  CONCAT('检测到C2 Beacon: 包数=', CAST(packet_count AS STRING), 
         ', 平均间隔=', CAST(ROUND(avg_interval, 2) AS STRING), 
         's, 方差=', CAST(ROUND(interval_variance, 2) AS STRING)) AS description,
  CASE 
    WHEN interval_variance < 0.5 THEN 0.95
    WHEN interval_variance < 1.0 THEN 0.85
    ELSE 0.75
  END AS confidence,
  window_end AS timestamp
FROM (
  SELECT 
    orig_h,
    resp_h,
    HOP_END(ts, INTERVAL '5' MINUTE, INTERVAL '10' MINUTE) as window_end,
    COUNT(*) as packet_count,
    AVG(duration) as avg_interval,
    STDDEV_SAMP(duration) as interval_variance,
    SUM(orig_bytes + resp_bytes) as total_bytes
  FROM zeek_conn
  WHERE duration > 0
  GROUP BY 
    orig_h, 
    resp_h,
    HOP(ts, INTERVAL '5' MINUTE, INTERVAL '10' MINUTE)
)
WHERE 
  interval_variance < 1.5 
  AND avg_interval > 20
  AND packet_count > 5
  AND total_bytes < 102400;
