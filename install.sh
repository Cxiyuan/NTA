#!/bin/bash
# NTA 离线安装脚本 - 支持 Kafka/Flink 流处理架构
# 版本: v2.0.0

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# 日志函数
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查root权限
check_root() {
    if [ "$EUID" -ne 0 ]; then
        log_error "请使用root权限运行此脚本"
        exit 1
    fi
}

# 检查系统
check_system() {
    log_info "检查系统环境..."
    
    # 检查CPU核心数
    cpu_cores=$(nproc)
    if [ "$cpu_cores" -lt 2 ]; then
        log_warn "CPU核心数不足，建议至少2核 (当前: ${cpu_cores}核)"
    fi
    
    # 检查内存
    mem_total=$(free -g | awk '/^Mem:/{print $2}')
    if [ "$mem_total" -lt 4 ]; then
        log_warn "内存不足，建议至少4GB (当前: ${mem_total}GB)"
    fi
    
    # 检查磁盘空间
    disk_free=$(df -BG / | awk 'NR==2 {print $4}' | sed 's/G//')
    if [ "$disk_free" -lt 50 ]; then
        log_warn "磁盘空间不足，建议至少50GB (当前剩余: ${disk_free}GB)"
    fi
    
    log_success "系统检查完成"
}

# 安装Docker
install_docker() {
    if command -v docker &> /dev/null; then
        log_info "Docker已安装，版本: $(docker --version)"
        return
    fi
    
    log_info "安装Docker..."
    
    if [ -f "docker/docker-24.0.7.tgz" ]; then
        tar -xzf docker/docker-24.0.7.tgz
        cp docker/* /usr/bin/
        
        # 创建systemd服务
        cat > /etc/systemd/system/docker.service << 'EOF'
[Unit]
Description=Docker Application Container Engine
After=network-online.target firewalld.service
Wants=network-online.target

[Service]
Type=notify
ExecStart=/usr/bin/dockerd
ExecReload=/bin/kill -s HUP $MAINPID
LimitNOFILE=infinity
LimitNPROC=infinity
LimitCORE=infinity
TimeoutStartSec=0
Delegate=yes
KillMode=process
Restart=on-failure
StartLimitBurst=3
StartLimitInterval=60s

[Install]
WantedBy=multi-user.target
EOF
        
        systemctl daemon-reload
        systemctl enable docker
        systemctl start docker
        
        log_success "Docker安装完成"
    else
        log_error "Docker安装包不存在"
        exit 1
    fi
}

# 安装Docker Compose
install_docker_compose() {
    if command -v docker-compose &> /dev/null; then
        log_info "Docker Compose已安装，版本: $(docker-compose --version)"
        return
    fi
    
    log_info "安装Docker Compose..."
    
    if [ -f "docker-compose" ]; then
        cp docker-compose /usr/local/bin/
        chmod +x /usr/local/bin/docker-compose
        log_success "Docker Compose安装完成"
    else
        log_error "Docker Compose文件不存在"
        exit 1
    fi
}

# 加载镜像
load_images() {
    log_info "加载Docker镜像..."
    
    if [ ! -d "images" ]; then
        log_error "镜像目录不存在"
        exit 1
    fi
    
    cd images
    
    # 基础镜像
    log_info "加载基础组件镜像..."
    docker load -i postgres.tar
    docker load -i redis.tar
    
    # 应用镜像
    log_info "加载应用镜像..."
    docker load -i nta-server.tar
    docker load -i nta-web.tar
    docker load -i nta-zeek.tar
    
    # 微服务镜像 (新增)
    log_info "加载微服务镜像..."
    docker load -i nta-auth-service.tar
    docker load -i nta-asset-service.tar
    docker load -i nta-detection-service.tar
    docker load -i nta-alert-service.tar
    docker load -i nta-report-service.tar
    docker load -i nta-notification-service.tar
    docker load -i nta-probe-service.tar
    docker load -i nta-intel-service.tar
    
    # API网关和基础设施 (新增)
    log_info "加载基础设施镜像..."
    docker load -i nta-traefik.tar
    docker load -i consul.tar
    docker load -i jaeger.tar
    
    # 流处理镜像
    log_info "加载流处理组件镜像..."
    docker load -i zookeeper.tar
    docker load -i kafka.tar
    docker load -i flink.tar
    docker load -i nta-kafka-consumer.tar
    
    # 监控镜像
    log_info "加载监控组件镜像..."
    docker load -i prometheus.tar
    docker load -i grafana.tar
    
    cd ..
    
    log_success "所有镜像加载完成"
    
    # 显示镜像列表
    log_info "已加载的镜像："
    docker images | grep -E "nta-|postgres|redis|zookeeper|kafka|flink|prometheus|grafana|consul|jaeger|traefik"
}

# 配置系统参数 (针对Kafka/Flink优化)
configure_system() {
    log_info "优化系统参数..."
    
    # 文件描述符限制
    if ! grep -q "* soft nofile 65536" /etc/security/limits.conf; then
        cat >> /etc/security/limits.conf << EOF
* soft nofile 65536
* hard nofile 65536
* soft nproc 32000
* hard nproc 32000
EOF
    fi
    
    # 内核参数优化 (Kafka需要)
    if ! grep -q "vm.max_map_count" /etc/sysctl.conf; then
        cat >> /etc/sysctl.conf << EOF
# Kafka/Flink 优化
vm.max_map_count=262144
vm.swappiness=1
net.core.somaxconn=1024
net.ipv4.tcp_max_syn_backlog=2048
EOF
        sysctl -p
    fi
    
    log_success "系统参数配置完成"
}

# 启动服务
start_services() {
    log_info "启动NTA服务..."
    
    # 设置环境变量
    export VERSION=$(cat VERSION 2>/dev/null || echo "v1.0.0")
    export BUILD_TIME=$(date -u +'%Y-%m-%dT%H:%M:%SZ')
    export GIT_COMMIT=$(cat GIT_COMMIT 2>/dev/null || echo "unknown")
    
    # 启动Docker Compose
    docker-compose up -d
    
    log_info "等待服务启动..."
    sleep 10
    
    # 检查服务状态
    log_info "检查服务状态..."
    docker-compose ps
    
    # 等待Kafka就绪
    log_info "等待Kafka集群启动..."
    local max_wait=60
    local wait_count=0
    
    while [ $wait_count -lt $max_wait ]; do
        if docker exec nta-kafka kafka-broker-api-versions.sh --bootstrap-server localhost:9092 &>/dev/null; then
            log_success "Kafka集群已就绪"
            break
        fi
        sleep 2
        wait_count=$((wait_count + 1))
    done
    
    if [ $wait_count -eq $max_wait ]; then
        log_warn "Kafka启动超时，请检查日志: docker logs nta-kafka"
    fi
    
    # 部署Flink作业
    if [ -f "flink-jobs/deploy-jobs.sh" ]; then
        log_info "部署Flink流处理作业..."
        
        # 等待Flink就绪
        sleep 15
        
        # 注意：Flink作业部署需要等待JobManager完全启动
        log_info "等待Flink JobManager启动..."
        local flink_wait=0
        while [ $flink_wait -lt 30 ]; do
            if curl -sf http://localhost:8081/overview &>/dev/null; then
                log_success "Flink JobManager已就绪"
                bash flink-jobs/deploy-jobs.sh || log_warn "Flink作业部署失败，请手动部署"
                break
            fi
            sleep 2
            flink_wait=$((flink_wait + 1))
        done
        
        if [ $flink_wait -eq 30 ]; then
            log_warn "Flink启动超时，请稍后手动部署作业"
            log_info "手动部署命令: bash flink-jobs/deploy-jobs.sh"
        fi
    fi
    
    log_success "NTA服务启动完成"
}

# 健康检查
health_check() {
    log_info "执行健康检查..."
    
    local services=(
        "nta-postgres:5432"
        "nta-redis:6379"
        "nta-consul:8500"
        "nta-zookeeper:2181"
        "nta-kafka:9092"
        "nta-flink-jobmanager:8081"
        "nta-traefik:80"
        "nta-auth-service:8081"
        "nta-asset-service:8082"
        "nta-detection-service:8083"
        "nta-alert-service:8084"
        "nta-jaeger:16686"
    )
    
    for service in "${services[@]}"; do
        local name="${service%%:*}"
        local port="${service##*:}"
        
        if docker ps | grep -q "$name"; then
            log_success "$name 运行中"
        else
            log_warn "$name 未运行"
        fi
    done
    
    # 检查微服务健康状态
    log_info "检查微服务健康状态..."
    sleep 5
    
    if curl -sf http://localhost/api/v1/auth/users &>/dev/null 2>&1 || curl -sf http://localhost:8081/health &>/dev/null 2>&1; then
        log_success "微服务API可访问"
    else
        log_warn "微服务API未就绪，可能正在启动中"
    fi
}

# 显示访问信息
show_info() {
    echo ""
    echo "=========================================="
    echo "  NTA 系统部署完成！"
    echo "=========================================="
    echo ""
    echo "📊 访问地址:"
    echo "  - Web界面:      http://$(hostname -I | awk '{print $1}')"
    echo "  - API服务:      http://$(hostname -I | awk '{print $1}'):8080"
    echo "  - Prometheus:   http://$(hostname -I | awk '{print $1}'):9090"
    echo "  - Grafana:      http://$(hostname -I | awk '{print $1}'):3000"
    echo "  - Flink Web UI: http://$(hostname -I | awk '{print $1}'):8081"
    echo ""
    echo "🔑 默认账户:"
    echo "  - 用户名: admin"
    echo "  - 密码:   admin123"
    echo ""
    echo "📝 常用命令:"
    echo "  - 查看日志:   docker-compose logs -f [service]"
    echo "  - 重启服务:   docker-compose restart [service]"
    echo "  - 停止服务:   docker-compose stop"
    echo "  - 启动服务:   docker-compose start"
    echo "  - 查看状态:   docker-compose ps"
    echo ""
    echo "🔧 流处理监控:"
    echo "  - Kafka Topic: docker exec nta-kafka kafka-topics.sh --list --bootstrap-server localhost:9092"
    echo "  - Flink Jobs:  curl http://localhost:8081/jobs"
    echo ""
    echo "=========================================="
}

# 主函数
main() {
    echo ""
    echo "╔════════════════════════════════════════╗"
    echo "║  NTA 网络流量分析系统 离线安装程序   ║"
    echo "║     支持 Kafka/Flink 流处理架构      ║"
    echo "╚════════════════════════════════════════╝"
    echo ""
    
    check_root
    check_system
    configure_system
    install_docker
    install_docker_compose
    load_images
    start_services
    sleep 5
    health_check
    show_info
}

main "$@"