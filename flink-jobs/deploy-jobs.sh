#!/bin/bash
# Flink作业部署脚本

set -e

FLINK_HOST=${FLINK_HOST:-localhost}
FLINK_PORT=${FLINK_PORT:-8081}
FLINK_SQL_CLIENT="docker exec -it nta-flink-jobmanager ./bin/sql-client.sh"

echo "==> 等待Flink集群就绪..."
until curl -sf http://${FLINK_HOST}:${FLINK_PORT}/overview > /dev/null; do
  echo "等待Flink启动..."
  sleep 5
done

echo "==> Flink集群已就绪"

# 提交SQL作业
JOBS_DIR="/opt/flink/jobs"

echo "==> 提交C2 Beacon检测作业..."
${FLINK_SQL_CLIENT} -f ${JOBS_DIR}/c2-beacon-detection.sql

echo "==> 提交DGA检测作业..."
${FLINK_SQL_CLIENT} -f ${JOBS_DIR}/dga-detection.sql

echo "==> 提交数据渗出检测作业..."
${FLINK_SQL_CLIENT} -f ${JOBS_DIR}/data-exfiltration-detection.sql

echo "==> 所有Flink作业已提交"
