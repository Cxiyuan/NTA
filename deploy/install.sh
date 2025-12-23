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

check_ubuntu_24() {
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
    echo ""
    echo "=========================================="
    echo "   NTA 探针安装完成"
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

main() {
    echo "=========================================="
    echo "   NTA 探针自动安装脚本"
    echo "   仅支持 Ubuntu 24.04 LTS"
    echo "=========================================="
    echo ""
    
    check_root
    check_ubuntu_24
    check_system_requirements
    
    read -p "是否继续安装? (y/n): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        log_info "安装已取消"
        exit 0
    fi
    
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
    show_help
    
    log_info "安装完成! 🎉"
}

main "$@"