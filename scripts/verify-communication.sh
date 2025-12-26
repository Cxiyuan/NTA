#!/bin/bash
# 服务间通信验证脚本

echo "=== NTA 微服务通信测试 ==="
echo ""

# 颜色定义
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

check_service() {
    local name=$1
    local url=$2
    
    if curl -sf "$url" > /dev/null 2>&1; then
        echo -e "${GREEN}✓${NC} $name 可访问"
        return 0
    else
        echo -e "${RED}✗${NC} $name 不可访问"
        return 1
    fi
}

echo "1. 检查微服务健康状态"
echo "-----------------------------------"
check_service "auth-service" "http://localhost:8081/health"
check_service "asset-service" "http://localhost:8082/health"
check_service "detection-service" "http://localhost:8083/health"
check_service "alert-service" "http://localhost:8084/health"
check_service "report-service" "http://localhost:8085/health"
check_service "notification-service" "http://localhost:8086/health"
check_service "probe-service" "http://localhost:8087/health"
check_service "intel-service" "http://localhost:8088/health"
echo ""

echo "2. 检查 Traefik 路由"
echo "-----------------------------------"
check_service "Traefik Dashboard" "http://localhost:8888/dashboard/"
check_service "API Gateway (auth)" "http://localhost/api/v1/auth/users"
echo ""

echo "3. 检查基础设施"
echo "-----------------------------------"
check_service "PostgreSQL" "localhost:5432"
check_service "Redis" "localhost:6379"
check_service "Consul" "http://localhost:8500/v1/status/leader"
check_service "Jaeger" "http://localhost:16686"
check_service "Kafka" "localhost:9092"
check_service "Flink" "http://localhost:8081/overview"
echo ""

echo "4. 检查数据库初始化"
echo "-----------------------------------"
if docker exec nta-postgres psql -U nta -lqt 2>/dev/null | cut -d \| -f 1 | grep -qw auth_db; then
    echo -e "${GREEN}✓${NC} auth_db 已创建"
else
    echo -e "${RED}✗${NC} auth_db 未创建"
fi

if docker exec nta-postgres psql -U nta -lqt 2>/dev/null | cut -d \| -f 1 | grep -qw asset_db; then
    echo -e "${GREEN}✓${NC} asset_db 已创建"
else
    echo -e "${RED}✗${NC} asset_db 未创建"
fi
echo ""

echo "5. 检查 Docker 网络"
echo "-----------------------------------"
if docker network inspect nta_nta-network > /dev/null 2>&1; then
    echo -e "${GREEN}✓${NC} nta-network 已创建"
    echo "   容器数量: $(docker network inspect nta_nta-network | jq '.[0].Containers | length')"
else
    echo -e "${RED}✗${NC} nta-network 不存在"
fi
echo ""

echo "=== 测试完成 ==="
