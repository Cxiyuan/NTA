-- Flink SQL作业：数据渗出检测
-- 检测异常大量上传流量

CREATE TABLE conn_stats (
  orig_h STRING,
  window_end TIMESTAMP(3),
  total_upload BIGINT,
  total_download BIGINT,
  upload_ratio DOUBLE,
  conn_count BIGINT,
  PRIMARY KEY (orig_h, window_end) NOT ENFORCED
) WITH (
  'connector' = 'upsert-kafka',
  'topic' = 'conn-stats',
  'properties.bootstrap.servers' = 'kafka:9092',
  'key.format' = 'json',
  'value.format' = 'json'
);

-- 计算5分钟内每个源IP的上传统计
INSERT INTO conn_stats
SELECT 
  orig_h,
  TUMBLE_END(ts, INTERVAL '5' MINUTE) as window_end,
  SUM(orig_bytes) as total_upload,
  SUM(resp_bytes) as total_download,
  CASE 
    WHEN SUM(resp_bytes) > 0 THEN CAST(SUM(orig_bytes) AS DOUBLE) / SUM(resp_bytes)
    ELSE 0
  END as upload_ratio,
  COUNT(*) as conn_count
FROM zeek_conn
GROUP BY 
  orig_h,
  TUMBLE(ts, INTERVAL '5' MINUTE);

-- 检测数据渗出异常
INSERT INTO alerts
SELECT 
  CONCAT('exfil-', orig_h, '-', CAST(window_end AS STRING)) AS alert_id,
  'data_exfiltration' AS alert_type,
  CASE 
    WHEN total_upload > 104857600 THEN 'critical'  -- >100MB
    WHEN total_upload > 52428800 THEN 'high'       -- >50MB
    ELSE 'medium'
  END AS severity,
  orig_h AS src_ip,
  '' AS dst_ip,
  CONCAT('检测到大量数据上传: ', CAST(total_upload / 1048576 AS STRING), 
         'MB, 上传/下载比=', CAST(ROUND(upload_ratio, 2) AS STRING)) AS description,
  CASE 
    WHEN upload_ratio > 10 THEN 0.90
    WHEN upload_ratio > 5 THEN 0.80
    ELSE 0.70
  END AS confidence,
  window_end AS timestamp
FROM conn_stats
WHERE 
  total_upload > 10485760  -- >10MB
  AND (
    upload_ratio > 5       -- 上传远大于下载
    OR total_upload > 52428800  -- 绝对值过大
  );
