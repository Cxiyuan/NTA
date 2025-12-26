#!/bin/bash
# PostgreSQL 多数据库初始化脚本

set -e

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
    -- 创建微服务数据库
    SELECT 'CREATE DATABASE auth_db' WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'auth_db')\gexec
    SELECT 'CREATE DATABASE asset_db' WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'asset_db')\gexec
    SELECT 'CREATE DATABASE alert_db' WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'alert_db')\gexec
    SELECT 'CREATE DATABASE report_db' WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'report_db')\gexec
    SELECT 'CREATE DATABASE notify_db' WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'notify_db')\gexec
    SELECT 'CREATE DATABASE probe_db' WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'probe_db')\gexec
    SELECT 'CREATE DATABASE intel_db' WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'intel_db')\gexec
    
    -- 授权
    GRANT ALL PRIVILEGES ON DATABASE auth_db TO nta;
    GRANT ALL PRIVILEGES ON DATABASE asset_db TO nta;
    GRANT ALL PRIVILEGES ON DATABASE alert_db TO nta;
    GRANT ALL PRIVILEGES ON DATABASE report_db TO nta;
    GRANT ALL PRIVILEGES ON DATABASE notify_db TO nta;
    GRANT ALL PRIVILEGES ON DATABASE probe_db TO nta;
    GRANT ALL PRIVILEGES ON DATABASE intel_db TO nta;
EOSQL

echo "✓ 微服务数据库初始化完成"
