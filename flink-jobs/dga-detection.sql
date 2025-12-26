-- Flink SQL作业：DGA域名检测
-- 通过统计特征检测算法生成的恶意域名

CREATE TABLE zeek_dns (
  uid STRING,
  ts TIMESTAMP(3),
  orig_h STRING,
  resp_h STRING,
  query STRING,
  qclass_name STRING,
  qtype_name STRING,
  rcode_name STRING,
  WATERMARK FOR ts AS ts - INTERVAL '5' SECOND
) WITH (
  'connector' = 'kafka',
  'topic' = 'zeek-dns',
  'properties.bootstrap.servers' = 'kafka:9092',
  'properties.group.id' = 'flink-dga-detector',
  'scan.startup.mode' = 'latest-offset',
  'format' = 'json'
);

-- DGA检测：长域名、高熵值、数字比例异常
INSERT INTO alerts
SELECT 
  CONCAT('dga-', orig_h, '-', query, '-', CAST(ts AS STRING)) AS alert_id,
  'dga_domain' AS alert_type,
  'medium' AS severity,
  orig_h AS src_ip,
  resp_h AS dst_ip,
  CONCAT('可疑DGA域名: ', query, ', 长度=', CAST(CHAR_LENGTH(query) AS STRING)) AS description,
  0.80 AS confidence,
  ts AS timestamp
FROM zeek_dns
WHERE 
  -- 域名长度异常
  CHAR_LENGTH(query) > 20
  -- 包含大量数字
  AND (CHAR_LENGTH(query) - CHAR_LENGTH(REGEXP_REPLACE(query, '[0-9]', ''))) > 5
  -- 子域名层级过多
  AND (CHAR_LENGTH(query) - CHAR_LENGTH(REGEXP_REPLACE(query, '\.', ''))) > 4
  -- 非常见顶级域名
  AND query NOT LIKE '%.com' 
  AND query NOT LIKE '%.cn'
  AND query NOT LIKE '%.net'
  AND query NOT LIKE '%.org';
