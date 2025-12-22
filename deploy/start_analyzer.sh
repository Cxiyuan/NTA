#!/bin/bash

ZEEK_LOG_DIR="/var/log/zeek/current"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/.."

echo "正在从Zeek日志读取数据并启动实时分析..."
echo "日志目录: $ZEEK_LOG_DIR"
echo ""

if [ ! -d "$ZEEK_LOG_DIR" ]; then
    echo "错误: Zeek日志目录不存在: $ZEEK_LOG_DIR"
    exit 1
fi

tail -F $ZEEK_LOG_DIR/*.log 2>/dev/null | \
    grep -v '^#' | \
    python3 $SCRIPT_DIR/analyzer/detector.py
