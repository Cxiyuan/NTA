#!/bin/bash

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

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

detect_os() {
    if [ -f /etc/os-release ]; then
        . /etc/os-release
        OS=$ID
        OS_VERSION=$VERSION_ID
    else
        log_error "无法检测操作系统类型"
        exit 1
    fi
    log_info "检测到操作系统: $OS $OS_VERSION"
}

check_docker() {
    if command -v docker &> /dev/null; then
        DOCKER_VERSION=$(docker --version | awk '{print $3}' | sed 's/,//')
        log_info "Docker 已安装: $DOCKER_VERSION"
        return 0
    else
        log_warn "Docker 未安装"
        return 1
    fi
}

check_docker_compose() {
    if command -v docker-compose &> /dev/null; then
        COMPOSE_VERSION=$(docker-compose --version | awk '{print $4}' | sed 's/,//')
        log_info "Docker Compose 已安装: $COMPOSE_VERSION"
        return 0
    elif docker compose version &> /dev/null; then
        COMPOSE_VERSION=$(docker compose version --short)
        log_info "Docker Compose Plugin 已安装: $COMPOSE_VERSION"
        return 0
    else
        log_warn "Docker Compose 未安装"
        return 1
    fi
}

install_docker_centos() {
    log_info "开始在 CentOS/RHEL 系统上安装 Docker..."
    
    yum install -y yum-utils device-mapper-persistent-data lvm2
    
    yum-config-manager --add-repo https://download.docker.com/linux/centos/docker-ce.repo
    
    yum install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin
    
    systemctl start docker
    systemctl enable docker
    
    log_info "Docker 安装完成"
}

install_docker_ubuntu() {
    log_info "开始在 Ubuntu/Debian 系统上安装 Docker..."
    
    apt-get update
    apt-get install -y ca-certificates curl gnupg lsb-release
    
    install -m 0755 -d /etc/apt/keyrings
    curl -fsSL https://download.docker.com/linux/ubuntu/gpg | gpg --dearmor -o /etc/apt/keyrings/docker.gpg
    chmod a+r /etc/apt/keyrings/docker.gpg
    
    echo \
      "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu \
      $(lsb_release -cs) stable" | tee /etc/apt/sources.list.d/docker.list > /dev/null
    
    apt-get update
    apt-get install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin
    
    systemctl start docker
    systemctl enable docker
    
    log_info "Docker 安装完成"
}

install_docker_rocky() {
    log_info "开始在 Rocky Linux/AlmaLinux 系统上安装 Docker..."
    
    dnf install -y yum-utils
    yum-config-manager --add-repo https://download.docker.com/linux/centos/docker-ce.repo
    
    dnf install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin
    
    systemctl start docker
    systemctl enable docker
    
    log_info "Docker 安装完成"
}

install_docker() {
    case "$OS" in
        centos|rhel|anolis)
            install_docker_centos
            ;;
        ubuntu|debian)
            install_docker_ubuntu
            ;;
        rocky|almalinux)
            install_docker_rocky
            ;;
        *)
            log_error "不支持的操作系统: $OS"
            log_info "请手动安装 Docker: https://docs.docker.com/engine/install/"
            exit 1
            ;;
    esac
}

install_docker_compose_standalone() {
    log_info "安装 Docker Compose 独立版本..."
    
    COMPOSE_VERSION="v2.24.5"
    curl -L "https://github.com/docker/compose/releases/download/${COMPOSE_VERSION}/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
    
    chmod +x /usr/local/bin/docker-compose
    
    ln -sf /usr/local/bin/docker-compose /usr/bin/docker-compose
    
    log_info "Docker Compose 安装完成"
}

verify_docker() {
    log_info "验证 Docker 安装..."
    
    if ! docker run --rm hello-world &> /dev/null; then
        log_error "Docker 安装验证失败"
        exit 1
    fi
    
    log_info "Docker 安装验证成功"
}

check_network_interface() {
    log_info "检测网络接口..."
    
    interfaces=$(ip -o link show | awk -F': ' '{print $2}' | grep -v lo)
    
    echo "可用的网络接口:"
    echo "$interfaces"
    
    default_iface=$(ip route | grep default | awk '{print $5}' | head -n1)
    
    if [ -z "$default_iface" ]; then
        log_warn "未检测到默认网络接口，请手动配置 ZEEK_IFACE"
    else
        log_info "检测到默认网络接口: $default_iface"
        
        if grep -q "ZEEK_IFACE=eth0" docker-compose.yml; then
            sed -i "s/ZEEK_IFACE=eth0/ZEEK_IFACE=$default_iface/g" docker-compose.yml
            log_info "已自动更新 docker-compose.yml 中的网络接口为: $default_iface"
        fi
    fi
}

create_directories() {
    log_info "创建必要的目录..."
    
    mkdir -p logs reports config
    
    log_info "目录创建完成"
}

extract_archive() {
    log_info "检查项目文件..."
    if [ ! -d "analyzer" ] || [ ! -d "backend" ]; then
        log_warn "项目文件不完整，这是正常的（使用 Docker 镜像部署）"
    fi
}

check_system_resources() {
    log_info "检查系统资源..."
    
    total_mem=$(free -g | awk '/^Mem:/{print $2}')
    if [ "$total_mem" -lt 8 ]; then
        log_warn "系统内存小于 8GB (当前: ${total_mem}GB)，可能影响性能"
    else
        log_info "系统内存: ${total_mem}GB ✓"
    fi
    
    total_disk=$(df -BG . | awk 'NR==2 {print $4}' | sed 's/G//')
    if [ "$total_disk" -lt 50 ]; then
        log_warn "可用磁盘空间小于 50GB (当前: ${total_disk}GB)，可能不足"
    else
        log_info "可用磁盘空间: ${total_disk}GB ✓"
    fi
    
    cpu_cores=$(nproc)
    if [ "$cpu_cores" -lt 4 ]; then
        log_warn "CPU 核心数小于 4 (当前: ${cpu_cores})，可能影响性能"
    else
        log_info "CPU 核心数: ${cpu_cores} ✓"
    fi
}

load_docker_image() {
    log_info "加载 Docker 镜像..."
    
    if [ -f "cap-agent-latest.tar.gz" ]; then
        if docker images cap-agent:latest | grep -q "cap-agent"; then
            log_info "镜像 cap-agent:latest 已存在"
        else
            docker load -i cap-agent-latest.tar.gz
            log_info "镜像加载完成"
        fi
    else
        log_error "未找到镜像文件 cap-agent-latest.tar.gz"
        exit 1
    fi
}

build_images() {
    log_info "检查 Docker 镜像..."
    
    if docker images cap-agent:latest | grep -q "cap-agent"; then
        log_info "镜像 cap-agent:latest 已存在，跳过构建"
    else
        log_warn "未找到镜像，尝试从 tar.gz 加载..."
        load_docker_image
    fi
}

start_services() {
    log_info "启动服务..."
    
    if command -v docker-compose &> /dev/null; then
        docker-compose up -d
    else
        docker compose up -d
    fi
    
    log_info "服务启动完成"
}

show_status() {
    log_info "服务状态:"
    
    if command -v docker-compose &> /dev/null; then
        docker-compose ps
    else
        docker compose ps
    fi
}

show_help() {
    echo ""
    echo "常用命令:"
    echo "  查看服务状态:    docker compose ps"
    echo "  查看日志:        docker compose logs -f"
    echo "  停止服务:        docker compose stop"
    echo "  启动服务:        docker compose start"
    echo "  重启服务:        docker compose restart"
    echo "  删除服务:        docker compose down"
    echo ""
    echo "验证 Zeek 运行:"
    echo "  docker compose exec cap-agent zeekctl status"
    echo ""
    echo "访问 Web 管理界面:"
    echo "  http://$(hostname -I | awk '{print $1}'):5000"
    echo ""
    echo "详细文档请参考: DOCKER_DEPLOYMENT.md"
    echo ""
}

main() {
    echo "=========================================="
    echo "   Cap Agent Docker 自动部署脚本"
    echo "=========================================="
    echo ""
    
    check_root
    detect_os
    check_system_resources
    
    echo ""
    echo "=========================================="
    echo "   检查 Docker 环境"
    echo "=========================================="
    echo ""
    
    if ! check_docker; then
        read -p "是否安装 Docker? (y/n): " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            install_docker
            verify_docker
        else
            log_error "Docker 是必需的，退出安装"
            exit 1
        fi
    fi
    
    if ! check_docker_compose; then
        if ! docker compose version &> /dev/null; then
            read -p "是否安装 Docker Compose? (y/n): " -n 1 -r
            echo
            if [[ $REPLY =~ ^[Yy]$ ]]; then
                install_docker_compose_standalone
            else
                log_error "Docker Compose 是必需的，退出安装"
                exit 1
            fi
        fi
    fi
    
    echo ""
    echo "=========================================="
    echo "   准备部署环境"
    echo "=========================================="
    echo ""
    
    extract_archive
    load_docker_image
    create_directories
    check_network_interface
    
    echo ""
    echo "=========================================="
    echo "   构建和启动服务"
    echo "=========================================="
    echo ""
    
    read -p "是否立即启动服务? (y/n): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        start_services
        
        log_info "等待服务启动..."
        sleep 10
        
        show_status
        show_help
        
        log_info "部署完成! 🎉"
    else
        log_info "已完成环境准备，您可以手动执行:"
        echo "  docker compose up -d"
    fi
}

main "$@"