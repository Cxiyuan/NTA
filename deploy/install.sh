#!/bin/bash

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

INSTALL_DIR="/opt/nta-probe"
SERVICE_USER="nta"
ZEEK_VERSION="6.0.3"
GO_VERSION="1.21.5"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

DEPLOY_MODE=""

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

check_root() {
    if [[ $EUID -ne 0 ]]; then
        log_error "此脚本需要 root 权限运行"
        exit 1
    fi
}

select_deploy_mode() {
    echo "请选择部署模式:"
    echo "  1) Docker 部署 (推荐，适用于所有系统)"
    echo "  2) 原生部署 (仅支持 Ubuntu 24.04)"
    echo ""
    read -p "请输入选项 [1-2]: " -n 1 -r
    echo ""
    
    case $REPLY in
        1)
            DEPLOY_MODE="docker"
            log_info "已选择: Docker 部署模式"
            ;;
        2)
            DEPLOY_MODE="native"
            log_info "已选择: 原生部署模式"
            ;;
        *)
            log_error "无效选项"
            exit 1
            ;;
    esac
}

check_docker() {
    if ! command -v docker &> /dev/null; then
        log_error "Docker 未安装，请先安装 Docker"
        log_info "安装命令: curl -fsSL https://get.docker.com | sh"
        exit 1
    fi
    log_info "✓ Docker 已安装: $(docker --version)"
}

check_docker_compose() {
    if ! command -v docker-compose &> /dev/null && ! docker compose version &> /dev/null; then
        log_error "Docker Compose 未安装，请先安装 Docker Compose"
        exit 1
    fi
    log_info "✓ Docker Compose 已安装"
}

check_ubuntu_24() {
    if [ "$DEPLOY_MODE" != "native" ]; then
        return 0
    fi
    
    if [ ! -f /etc/os-release ]; then
        log_error "无法检测操作系统"
        exit 1
    fi
    
    . /etc/os-release
    
    if [ "$ID" != "ubuntu" ]; then
        log_error "仅支持 Ubuntu 系统，当前系统: $ID"
        exit 1
    fi
    
    if [ "$VERSION_ID" != "24.04" ]; then
        log_error "仅支持 Ubuntu 24.04，当前版本: $VERSION_ID"
        exit 1
    fi
    
    log_info "✓ 检测到 Ubuntu 24.04"
}

check_system_requirements() {
    if [ "$DEPLOY_MODE" != "native" ]; then
        log_info "跳过系统要求检查 (Docker 模式)"
        return 0
    fi
    
    log_info "检查系统要求..."
    
    if [ "$(uname -m)" != "x86_64" ]; then
        log_error "仅支持 x86_64 架构"
        exit 1
    fi
    
    total_mem=$(free -g | awk '/^Mem:/{print $2}')
    if [ "$total_mem" -lt 4 ]; then
        log_warn "系统内存小于 4GB (当前: ${total_mem}GB)，可能影响性能"
    else
        log_info "系统内存: ${total_mem}GB ✓"
    fi
    
    cpu_cores=$(nproc)
    if [ "$cpu_cores" -lt 4 ]; then
        log_warn "CPU 核心数小于 4 (当前: ${cpu_cores})，可能影响性能"
    else
        log_info "CPU 核心数: ${cpu_cores} ✓"
    fi
}

install_dependencies() {
    log_info "安装 Ubuntu 24.04 依赖包..."
    
    export DEBIAN_FRONTEND=noninteractive
    
    apt-get update
    apt-get install -y \
        cmake make gcc g++ flex bison \
        libpcap-dev libssl-dev \
        swig zlib1g-dev git wget curl \
        tcpdump net-tools redis-server
    
    log_info "依赖包安装完成"
}

install_golang() {
    if command -v go &> /dev/null; then
        GO_INSTALLED_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
        log_info "Go 已安装: $GO_INSTALLED_VERSION"
        return 0
    fi

    log_info "安装 Go ${GO_VERSION}..."
    
    cd /tmp
    wget https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz
    rm -rf /usr/local/go
    tar -C /usr/local -xzf go${GO_VERSION}.linux-amd64.tar.gz
    
    echo 'export PATH=$PATH:/usr/local/go/bin' >> /etc/profile.d/golang.sh
    export PATH=$PATH:/usr/local/go/bin
    
    rm -f go${GO_VERSION}.linux-amd64.tar.gz
    
    log_info "Go 安装完成: $(go version)"
}

install_zeek() {
    if command -v zeek &> /dev/null; then
        ZEEK_INSTALLED=$(zeek --version 2>&1 | head -n1 | awk '{print $2}')
        log_info "Zeek 已安装: $ZEEK_INSTALLED"
        return 0
    fi

    log_info "编译安装 Zeek ${ZEEK_VERSION}..."
    
    cd /tmp
    wget https://download.zeek.org/zeek-${ZEEK_VERSION}.tar.gz
    tar -xzf zeek-${ZEEK_VERSION}.tar.gz
    cd zeek-${ZEEK_VERSION}
    
    ./configure --prefix=/opt/zeek
    make -j$(nproc)
    make install
    
    echo 'export PATH="/opt/zeek/bin:$PATH"' >> /etc/profile.d/zeek.sh
    export PATH="/opt/zeek/bin:$PATH"
    
    cd /tmp
    rm -rf zeek-${ZEEK_VERSION} zeek-${ZEEK_VERSION}.tar.gz
    
    log_info "Zeek 安装完成"
}

create_service_user() {
    if id "$SERVICE_USER" &>/dev/null; then
        log_info "用户 $SERVICE_USER 已存在"
    else
        useradd -r -s /bin/bash -d "$INSTALL_DIR" "$SERVICE_USER"
        log_info "创建服务用户: $SERVICE_USER"
    fi
}

build_nta() {
    log_info "编译 NTA 探针..."
    
    SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
    cd "$SCRIPT_DIR/.."
    
    # Build Go binary
    export PATH=$PATH:/usr/local/go/bin
    go build -o nta-server ./cmd/nta-server
    
    log_info "NTA 编译完成"
}

install_nta_probe() {
    log_info "安装 NTA 探针..."
    
    SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
    
    mkdir -p "$INSTALL_DIR"/{bin,config,zeek-scripts,logs,reports,data}
    
    # Copy binary
    cp "$SCRIPT_DIR/../nta-server" "$INSTALL_DIR/bin/"
    chmod +x "$INSTALL_DIR/bin/nta-server"
    
    # Copy Zeek scripts
    cp -r "$SCRIPT_DIR/../zeek-scripts"/* "$INSTALL_DIR/zeek-scripts/"
    
    # Copy config files
    if [ -d "$SCRIPT_DIR/../config" ]; then
        cp -r "$SCRIPT_DIR/../config"/* "$INSTALL_DIR/config/" 2>/dev/null || true
    fi
    
    # Create default config
    cat > "$INSTALL_DIR/config/nta.yaml" <<EOF
server:
  host: 0.0.0.0
  port: 8080
  mode: release

zeek:
  log_dir: /var/spool/zeek
  script_dir: $INSTALL_DIR/zeek-scripts
  interface: eth0

redis:
  addr: localhost:6379
  password: ""
  db: 0

database:
  type: sqlite
  dsn: $INSTALL_DIR/data/nta.db

detection:
  scan:
    threshold: 20
    time_window: 300
    min_fail_rate: 0.6
  auth:
    fail_threshold: 5
    pth_window: 3600
  ml:
    enabled: true
    contamination: 0.01

threat_intel:
  sources:
    - name: threatfox
      url: https://threatfox-api.abuse.ch/api/v1/
      enabled: true
  update_interval: 3600
  local_feed_path: $INSTALL_DIR/config/threat_feed.json

license:
  license_file: $INSTALL_DIR/config/license.key
  public_key_file: $INSTALL_DIR/config/public.pem
EOF
    
    chown -R $SERVICE_USER:$SERVICE_USER "$INSTALL_DIR"
    
    log_info "NTA 探针安装完成"
}

configure_zeek() {
    log_info "配置 Zeek..."
    
    echo "@load $INSTALL_DIR/zeek-scripts/main.zeek" >> /opt/zeek/share/zeek/site/local.zeek
    
    default_iface=$(ip route | grep default | awk '{print $5}' | head -n1)
    if [ -n "$default_iface" ]; then
        log_info "配置 Zeek 监听接口: $default_iface"
        sed -i "s/interface=eth0/interface=$default_iface/" /opt/zeek/etc/node.cfg
        sed -i "s/interface: eth0/interface: $default_iface/" "$INSTALL_DIR/config/nta.yaml"
    fi
    
    zeekctl deploy
}

# ============================================
# Docker 部署相关函数
# ============================================

check_docker_config() {
    log_info "检查 Docker 配置文件..."
    
    if [ ! -f "$PROJECT_ROOT/config/nta.yaml" ]; then
        log_warn "配置文件不存在，从示例文件创建..."
        if [ -f "$PROJECT_ROOT/config/nta.yaml.example" ]; then
            cp "$PROJECT_ROOT/config/nta.yaml.example" "$PROJECT_ROOT/config/nta.yaml"
            log_info "✓ 已创建配置文件: config/nta.yaml"
        else
            log_error "示例配置文件不存在: config/nta.yaml.example"
            exit 1
        fi
    else
        log_info "✓ 配置文件已存在: config/nta.yaml"
    fi
    
    if ! grep -q "nta-postgres" "$PROJECT_ROOT/config/nta.yaml"; then
        log_warn "配置文件中数据库地址可能不正确，正在自动修复..."
        sed -i 's/host=localhost/host=nta-postgres/g' "$PROJECT_ROOT/config/nta.yaml"
        sed -i 's/host=postgres /host=nta-postgres /g' "$PROJECT_ROOT/config/nta.yaml"
    fi
    
    if ! grep -q "nta-redis" "$PROJECT_ROOT/config/nta.yaml"; then
        log_warn "配置文件中 Redis 地址可能不正确，正在自动修复..."
        sed -i 's/addr: localhost:6379/addr: nta-redis:6379/g' "$PROJECT_ROOT/config/nta.yaml"
        sed -i 's/addr: redis:6379/addr: nta-redis:6379/g' "$PROJECT_ROOT/config/nta.yaml"
    fi
    
    log_info "✓ 配置文件检查完成"
}

check_docker_images() {
    log_info "检查 Docker 镜像..."
    
    if ! docker images | grep -q "nta-server.*v1.0.0"; then
        log_error "nta-server:v1.0.0 镜像不存在"
        log_info "请先通过以下方式之一获取镜像:"
        log_info "  1. 从 GitHub Actions 下载并导入: docker load -i nta-server-v1.0.0.tar"
        log_info "  2. 或在本地构建: docker build -t nta-server:v1.0.0 -f Dockerfile ."
        exit 1
    fi
    
    if ! docker images | grep -q "nta-web.*v1.0.0"; then
        log_warn "nta-web:v1.0.0 镜像不存在，将跳过 Web UI 部署"
        log_info "如需部署 Web UI，请先获取镜像:"
        log_info "  1. 从 GitHub Actions 下载并导入: docker load -i nta-web-v1.0.0.tar"
        log_info "  2. 或在本地构建: docker build -t nta-web:v1.0.0 -f web/Dockerfile web/"
    fi
    
    log_info "✓ 必需镜像已存在"
}

cleanup_old_containers() {
    log_info "清理旧容器..."
    
    cd "$PROJECT_ROOT"
    if command -v docker-compose &> /dev/null; then
        docker-compose down 2>/dev/null || true
    else
        docker compose down 2>/dev/null || true
    fi
    
    docker ps -a | grep "nta-" | awk '{print $1}' | xargs -r docker rm -f 2>/dev/null || true
    
    log_info "✓ 清理完成"
}

start_docker_containers() {
    log_info "启动 Docker 容器..."
    
    cd "$PROJECT_ROOT"
    
    if command -v docker-compose &> /dev/null; then
        docker-compose up -d
    else
        docker compose up -d
    fi
    
    log_info "✓ 容器已启动"
}

wait_for_docker_services() {
    log_info "等待服务启动..."
    
    local max_attempts=60
    local attempt=0
    
    while [ $attempt -lt $max_attempts ]; do
        if docker ps | grep -q "nta-postgres.*Up"; then
            if docker ps | grep -q "nta-redis.*Up"; then
                if docker ps | grep -q "nta-server.*Up"; then
                    log_info "✓ 所有服务已启动"
                    return 0
                fi
            fi
        fi
        attempt=$((attempt + 1))
        sleep 2
        echo -n "."
    done
    
    echo ""
    log_error "服务启动超时"
    return 1
}

check_docker_status() {
    log_info "检查容器状态..."
    echo ""
    docker ps -a --filter "name=nta-" --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"
    echo ""
    
    if docker ps | grep -q "nta-server.*Restarting"; then
        log_error "nta-server 容器持续重启，查看日志:"
        docker logs nta-server --tail 50
        return 1
    fi
    
    if ! docker ps | grep -q "nta-server.*Up"; then
        log_error "nta-server 容器未正常运行"
        docker logs nta-server --tail 50
        return 1
    fi
    
    log_info "✓ 容器状态正常"
    return 0
}

show_docker_logs() {
    log_info "显示 nta-server 日志 (最近 20 行):"
    echo ""
    docker logs nta-server --tail 20 2>&1 || true
    echo ""
}

deploy_docker() {
    log_info "开始 Docker 部署..."
    
    check_docker
    check_docker_compose
    check_docker_config
    check_docker_images
    cleanup_old_containers
    start_docker_containers
    
    sleep 5
    
    if wait_for_docker_services; then
        sleep 3
        if check_docker_status; then
            show_docker_logs
            return 0
        else
            log_error "容器状态异常"
            return 1
        fi
    else
        log_error "服务启动失败"
        show_docker_logs
        return 1
    fi
}

# ============================================
# 原生部署相关函数
# ============================================

deploy_native() {
    log_info "开始原生部署..."
    
    install_dependencies
    install_golang
    install_zeek
    create_service_user
    build_nta
    install_nta_probe
    configure_zeek
    create_systemd_services
    start_services
    
    sleep 5
    show_status
}

create_systemd_services() {
    log_info "创建 systemd 服务..."
    
    cat > /etc/systemd/system/nta-zeek.service <<EOF
[Unit]
Description=NTA Zeek Service
After=network.target

[Service]
Type=forking
User=root
ExecStart=/opt/zeek/bin/zeekctl deploy
ExecStop=/opt/zeek/bin/zeekctl stop
ExecReload=/opt/zeek/bin/zeekctl restart
Restart=on-failure

[Install]
WantedBy=multi-user.target
EOF

    cat > /etc/systemd/system/nta-server.service <<EOF
[Unit]
Description=NTA Server
After=network.target nta-zeek.service redis.service

[Service]
Type=simple
User=$SERVICE_USER
WorkingDirectory=$INSTALL_DIR
ExecStart=$INSTALL_DIR/bin/nta-server -config $INSTALL_DIR/config/nta.yaml
Restart=always
RestartSec=10
Environment="PATH=/usr/local/go/bin:/opt/zeek/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin"

[Install]
WantedBy=multi-user.target
EOF

    systemctl daemon-reload
}

start_services() {
    log_info "启动服务..."
    
    systemctl enable redis
    systemctl start redis
    
    systemctl enable nta-zeek
    systemctl start nta-zeek
    
    sleep 5
    
    systemctl enable nta-server
    systemctl start nta-server
}

show_status() {
    log_info "服务状态:"
    echo ""
    systemctl status nta-zeek --no-pager -l || true
    echo ""
    systemctl status nta-server --no-pager -l || true
}

show_help() {
    if [ "$DEPLOY_MODE" = "docker" ]; then
        show_docker_help
    else
        show_native_help
    fi
}

show_native_help() {
    echo ""
    echo "=========================================="
    echo "   NTA 探针安装完成 (原生模式)"
    echo "=========================================="
    echo ""
    echo "服务管理命令:"
    echo "  启动服务:    systemctl start nta-zeek nta-server"
    echo "  停止服务:    systemctl stop nta-zeek nta-server"
    echo "  重启服务:    systemctl restart nta-zeek nta-server"
    echo "  查看状态:    systemctl status nta-zeek nta-server"
    echo ""
    echo "日志查看:"
    echo "  系统日志:    journalctl -u nta-server -f"
    echo "  Zeek日志:    tail -f /var/spool/zeek/current/*.log"
    echo ""
    echo "访问地址:"
    echo "  API Server: http://$(hostname -I | awk '{print $1}'):8080"
    echo "  Health:     http://$(hostname -I | awk '{print $1}'):8080/health"
    echo ""
    echo "配置文件:    $INSTALL_DIR/config/nta.yaml"
    echo "安装目录:    $INSTALL_DIR"
    echo ""
}

show_docker_help() {
    echo ""
    echo "=========================================="
    echo "   NTA Docker 部署完成"
    echo "=========================================="
    echo ""
    echo "服务访问地址:"
    echo "  API Server:  http://$(hostname -I | awk '{print $1}'):8080"
    echo "  Web UI:      http://$(hostname -I | awk '{print $1}'):80"
    echo "  Grafana:     http://$(hostname -I | awk '{print $1}'):3000  (admin/admin)"
    echo "  Prometheus:  http://$(hostname -I | awk '{print $1}'):9090"
    echo ""
    echo "服务管理命令:"
    if command -v docker-compose &> /dev/null; then
        echo "  启动所有服务:  cd $PROJECT_ROOT && docker-compose up -d"
        echo "  停止所有服务:  cd $PROJECT_ROOT && docker-compose down"
        echo "  重启服务:      cd $PROJECT_ROOT && docker-compose restart nta-server"
        echo "  查看日志:      docker logs -f nta-server"
        echo "  查看状态:      docker-compose ps"
    else
        echo "  启动所有服务:  cd $PROJECT_ROOT && docker compose up -d"
        echo "  停止所有服务:  cd $PROJECT_ROOT && docker compose down"
        echo "  重启服务:      cd $PROJECT_ROOT && docker compose restart nta-server"
        echo "  查看日志:      docker logs -f nta-server"
        echo "  查看状态:      docker compose ps"
    fi
    echo ""
    echo "容器管理:"
    echo "  进入容器:      docker exec -it nta-server sh"
    echo "  重启容器:      docker restart nta-server"
    echo ""
    echo "配置文件位置:  $PROJECT_ROOT/config/nta.yaml"
    echo "查看数据卷:    docker volume ls | grep nta"
    echo ""
}

main() {
    echo "=========================================="
    echo "   NTA 自动安装脚本"
    echo "=========================================="
    echo ""
    
    check_root
    select_deploy_mode
    
    if [ "$DEPLOY_MODE" = "native" ]; then
        check_ubuntu_24
    fi
    
    check_system_requirements
    
    echo ""
    read -p "是否继续安装? (y/n): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        log_info "安装已取消"
        exit 0
    fi
    
    if [ "$DEPLOY_MODE" = "docker" ]; then
        if deploy_docker; then
            show_help
            log_info "部署完成! 🎉"
        else
            log_error "部署失败，请查看上方错误信息"
            exit 1
        fi
    else
        deploy_native
        show_help
        log_info "安装完成! 🎉"
    fi
}

main "$@"