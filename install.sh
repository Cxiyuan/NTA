#!/bin/bash
# NTA 离线安装脚本 - 预编译包部署
# 版本: v2.0.0
# 支持系统: Ubuntu 24.04 LTS

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# 安装路径
INSTALL_DIR="/opt/nta"
DATA_DIR="/var/lib/nta"
LOG_DIR="/var/log/nta"
SERVICE_USER="nta"

# 当前脚本目录
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# 日志函数
log_info() {
    echo -e "${BLUE}[信息]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[成功]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[警告]${NC} $1"
}

log_error() {
    echo -e "${RED}[错误]${NC} $1"
}

# 检查root权限
check_root() {
    if [ "$EUID" -ne 0 ]; then
        log_error "请使用root权限运行此脚本"
        exit 1
    fi
}

# 检测操作系统
detect_os() {
    if [ ! -f /etc/os-release ]; then
        log_error "无法检测操作系统版本"
        exit 1
    fi
    
    . /etc/os-release
    
    if [ "$ID" != "ubuntu" ]; then
        log_error "本脚本仅支持 Ubuntu 系统"
        log_error "当前系统: $PRETTY_NAME"
        exit 1
    fi
    
    if [ "$VERSION_ID" != "24.04" ]; then
        log_error "本脚本仅支持 Ubuntu 24.04 LTS"
        log_error "当前版本: $VERSION_ID"
        exit 1
    fi
    
    log_success "检测到系统: Ubuntu 24.04 LTS"
}

# 检查系统要求
check_requirements() {
    log_info "检查系统要求..."
    
    # CPU
    cpu_cores=$(nproc)
    if [ "$cpu_cores" -lt 4 ]; then
        log_warn "CPU核心数不足，建议至少4核 (当前: ${cpu_cores}核)"
    else
        log_success "CPU核心数: ${cpu_cores}核"
    fi
    
    # 内存
    mem_total=$(free -g | awk '/^Mem:/{print $2}')
    if [ "$mem_total" -lt 8 ]; then
        log_warn "内存不足，建议至少8GB (当前: ${mem_total}GB)"
    else
        log_success "内存: ${mem_total}GB"
    fi
    
    # 磁盘
    disk_free=$(df -BG / | awk 'NR==2 {print $4}' | sed 's/G//')
    if [ "$disk_free" -lt 50 ]; then
        log_warn "磁盘空间不足，建议至少50GB (当前剩余: ${disk_free}GB)"
    else
        log_success "磁盘空间: ${disk_free}GB 可用"
    fi
}

# 安装系统依赖 (仅运行时库)
install_system_deps() {
    log_info "更新软件源..."
    apt-get update
    
    log_info "安装运行时依赖库..."
    DEBIAN_FRONTEND=noninteractive apt-get install -y \
        libpcap0.8 libssl3 zlib1g \
        libreadline8 libncurses6 \
        python3 \
        net-tools tcpdump iproute2 \
        systemd \
        libmaxminddb0 libkrb5-3 \
        default-jre-headless
    
    log_success "运行时依赖安装完成"
}

# 创建系统用户
create_user() {
    if id "$SERVICE_USER" &>/dev/null; then
        log_info "用户 $SERVICE_USER 已存在"
    else
        useradd -r -s /bin/bash -d $INSTALL_DIR -m $SERVICE_USER
        log_success "创建服务用户: $SERVICE_USER"
    fi
}

# 创建目录结构
create_directories() {
    log_info "创建目录结构..."
    
    mkdir -p $INSTALL_DIR/{bin,config,web}
    mkdir -p $DATA_DIR/{postgres,redis,kafka,zeek-logs,pcap,backups}
    mkdir -p $LOG_DIR/{nta,postgres,redis,kafka,zeek}
    
    # 确保日志目录权限正确
    chown -R $SERVICE_USER:$SERVICE_USER $DATA_DIR
    chown -R $SERVICE_USER:$SERVICE_USER $LOG_DIR
    
    log_success "目录创建完成"
}

# 安装PostgreSQL (预编译包)
install_postgres() {
    log_info "安装 PostgreSQL (预编译包)..."
    
    if [ -d "/opt/postgres" ]; then
        log_info "PostgreSQL 已安装，跳过"
        return
    fi
    
    cd $SCRIPT_DIR/depend
    
    log_info "解压 PostgreSQL..."
    tar -xzf postgresql-*-ubuntu24.04-amd64.tar.gz -C /
    
    # 初始化数据库
    chown -R $SERVICE_USER:$SERVICE_USER $DATA_DIR/postgres
    su - $SERVICE_USER -c "/opt/postgres/bin/initdb -D $DATA_DIR/postgres"
    
    # 配置PostgreSQL
    cat >> $DATA_DIR/postgres/postgresql.conf << EOF
listen_addresses = 'localhost'
port = 5432
max_connections = 200
shared_buffers = 256MB
EOF
    
    cat > $DATA_DIR/postgres/pg_hba.conf << EOF
local   all             all                                     trust
host    all             all             127.0.0.1/32            trust
host    all             all             ::1/128                 trust
EOF
    
    chown -R $SERVICE_USER:$SERVICE_USER $DATA_DIR/postgres
    
    log_success "PostgreSQL 安装完成"
}

# 安装Redis (预编译包)
install_redis() {
    log_info "安装 Redis (预编译包)..."
    
    if [ -d "/opt/redis" ]; then
        log_info "Redis 已安装，跳过"
        return
    fi
    
    cd $SCRIPT_DIR/depend
    
    log_info "解压 Redis..."
    tar -xzf redis-*-ubuntu24.04-amd64.tar.gz -C /
    
    # 配置Redis
    mkdir -p /opt/redis/etc
    cat > /opt/redis/etc/redis.conf << EOF
bind 127.0.0.1
port 6379
daemonize no
dir $DATA_DIR/redis
logfile $LOG_DIR/redis/redis.log
appendonly yes
appendfilename "appendonly.aof"
EOF
    
    chown -R $SERVICE_USER:$SERVICE_USER $DATA_DIR/redis
    
    log_success "Redis 安装完成"
}

# 安装Kafka (预编译包)
install_kafka() {
    log_info "安装 Kafka (预编译包)..."
    
    if [ -d "/opt/kafka" ]; then
        log_info "Kafka 已安装，跳过"
        return
    fi
    
    cd $SCRIPT_DIR/depend
    
    log_info "解压 Kafka..."
    tar -xzf kafka-*-bin.tar.gz -C /
    
    # 配置Kafka
    cat > /opt/kafka/config/server.properties << EOF
broker.id=0
listeners=PLAINTEXT://localhost:9092
log.dirs=$DATA_DIR/kafka
num.partitions=8
log.retention.hours=168
log.retention.bytes=10737418240
zookeeper.connect=localhost:2181
auto.create.topics.enable=true
EOF
    
    # 配置Zookeeper
    cat > /opt/kafka/config/zookeeper.properties << EOF
dataDir=$DATA_DIR/kafka/zookeeper
clientPort=2181
maxClientCnxns=0
admin.enableServer=false
EOF
    
    mkdir -p $DATA_DIR/kafka/zookeeper
    chown -R $SERVICE_USER:$SERVICE_USER /opt/kafka $DATA_DIR/kafka
    
    log_success "Kafka 安装完成"
}

# 安装Zeek (预编译包)
install_zeek() {
    log_info "安装 Zeek (预编译包)..."
    
    if [ -d "/opt/zeek" ]; then
        log_info "Zeek 已安装，跳过"
        return
    fi
    
    cd $SCRIPT_DIR/depend
    
    log_info "解压 Zeek..."
    tar -xzf zeek-*-ubuntu24.04-amd64.tar.gz -C /
    
    # 配置Zeek
    cat > /opt/zeek/etc/node.cfg << EOF
[zeek]
type=standalone
host=localhost
interface=eth0
EOF
    
    cat > /opt/zeek/etc/networks.cfg << EOF
10.0.0.0/8      Private IP space
172.16.0.0/12   Private IP space
192.168.0.0/16  Private IP space
EOF
    
    # 复制自定义脚本
    if [ -d "$SCRIPT_DIR/zeek-scripts" ]; then
        cp -r $SCRIPT_DIR/zeek-scripts/* /opt/zeek/share/zeek/site/
    fi
    
    echo "@load site" >> /opt/zeek/share/zeek/site/local.zeek
    
    chown -R root:root /opt/zeek
    chown -R $SERVICE_USER:$SERVICE_USER $DATA_DIR/zeek-logs
    
    log_success "Zeek 安装完成"
}

# 安装NTA应用
install_nta() {
    log_info "安装 NTA 应用..."
    
    # 复制二进制文件
    cp $SCRIPT_DIR/bin/nta-server $INSTALL_DIR/bin/
    cp $SCRIPT_DIR/bin/nta-kafka-consumer $INSTALL_DIR/bin/
    chmod +x $INSTALL_DIR/bin/*
    
    # 复制Web前端
    cp -r $SCRIPT_DIR/web/* $INSTALL_DIR/web/
    
    # 复制配置文件
    cp -r $SCRIPT_DIR/config/* $INSTALL_DIR/config/
    
    chown -R $SERVICE_USER:$SERVICE_USER $INSTALL_DIR
    
    log_success "NTA 应用安装完成"
}

# 创建systemd服务
create_services() {
    log_info "创建 systemd 服务..."
    
    # PostgreSQL服务
    cat > /etc/systemd/system/nta-postgres.service << EOF
[Unit]
Description=NTA PostgreSQL Database
After=network.target

[Service]
Type=simple
User=$SERVICE_USER
Group=$SERVICE_USER
Environment="PGDATA=$DATA_DIR/postgres"
ExecStart=/opt/postgres/bin/postgres -D $DATA_DIR/postgres
ExecReload=/bin/kill -HUP \$MAINPID
KillMode=mixed
KillSignal=SIGINT
TimeoutSec=infinity
StandardOutput=append:$LOG_DIR/postgres/postgres.log
StandardError=append:$LOG_DIR/postgres/postgres.log
Restart=on-failure

[Install]
WantedBy=multi-user.target
EOF
    
    # Redis服务
    cat > /etc/systemd/system/nta-redis.service << EOF
[Unit]
Description=NTA Redis Server
After=network.target

[Service]
Type=simple
User=$SERVICE_USER
Group=$SERVICE_USER
ExecStart=/opt/redis/bin/redis-server /opt/redis/etc/redis.conf
Restart=always

[Install]
WantedBy=multi-user.target
EOF
    
    # Zookeeper服务
    cat > /etc/systemd/system/nta-zookeeper.service << EOF
[Unit]
Description=NTA Zookeeper Service
After=network.target

[Service]
Type=simple
User=$SERVICE_USER
Group=$SERVICE_USER
Environment="KAFKA_HOME=/opt/kafka"
Environment="LOG_DIR=$LOG_DIR/kafka"
ExecStart=/opt/kafka/bin/zookeeper-server-start.sh /opt/kafka/config/zookeeper.properties
Restart=on-failure
TimeoutSec=300

[Install]
WantedBy=multi-user.target
EOF
    
    # Kafka服务
    cat > /etc/systemd/system/nta-kafka.service << EOF
[Unit]
Description=NTA Kafka Service
After=network.target nta-zookeeper.service
Requires=nta-zookeeper.service

[Service]
Type=simple
User=$SERVICE_USER
Group=$SERVICE_USER
Environment="KAFKA_HOME=/opt/kafka"
Environment="LOG_DIR=$LOG_DIR/kafka"
ExecStart=/opt/kafka/bin/kafka-server-start.sh /opt/kafka/config/server.properties
Restart=on-failure
TimeoutSec=300

[Install]
WantedBy=multi-user.target
EOF
    
    # Zeek服务
    cat > /etc/systemd/system/nta-zeek.service << EOF
[Unit]
Description=NTA Zeek Network Monitor
After=network.target nta-kafka.service
Requires=nta-kafka.service

[Service]
Type=forking
User=root
Environment="PATH=/opt/zeek/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"
ExecStart=/opt/zeek/bin/zeekctl deploy
ExecStop=/opt/zeek/bin/zeekctl stop
ExecReload=/opt/zeek/bin/zeekctl restart
Restart=on-failure
TimeoutSec=300

[Install]
WantedBy=multi-user.target
EOF
    
    # Kafka Consumer服务
    cat > /etc/systemd/system/nta-kafka-consumer.service << EOF
[Unit]
Description=NTA Kafka Consumer
After=network.target nta-kafka.service nta-postgres.service
Requires=nta-kafka.service nta-postgres.service

[Service]
Type=simple
User=$SERVICE_USER
Group=$SERVICE_USER
WorkingDirectory=$INSTALL_DIR
Environment="PATH=/opt/zeek/bin:/opt/postgres/bin:/usr/local/bin:/usr/bin:/bin"
ExecStart=$INSTALL_DIR/bin/nta-kafka-consumer
Restart=always

[Install]
WantedBy=multi-user.target
EOF
    
    # NTA主服务
    cat > /etc/systemd/system/nta-server.service << EOF
[Unit]
Description=NTA Server
After=network.target nta-postgres.service nta-redis.service nta-kafka.service
Requires=nta-postgres.service nta-redis.service

[Service]
Type=simple
User=$SERVICE_USER
Group=$SERVICE_USER
WorkingDirectory=$INSTALL_DIR
Environment="PATH=/opt/zeek/bin:/opt/postgres/bin:/usr/local/bin:/usr/bin:/bin"
ExecStart=$INSTALL_DIR/bin/nta-server -config $INSTALL_DIR/config/nta.yaml
Restart=always

[Install]
WantedBy=multi-user.target
EOF
    
    systemctl daemon-reload
    
    log_success "Systemd 服务创建完成"
}

# 初始化数据库
init_database() {
    log_info "初始化数据库..."
    
    # 启动PostgreSQL
    systemctl start nta-postgres
    
    # 等待PostgreSQL就绪，最多等待30秒
    local max_wait=30
    local waited=0
    while [ $waited -lt $max_wait ]; do
        if su - $SERVICE_USER -c "/opt/postgres/bin/pg_isready -q" 2>/dev/null; then
            log_success "PostgreSQL 已就绪"
            break
        fi
        sleep 1
        waited=$((waited + 1))
        if [ $((waited % 5)) -eq 0 ]; then
            log_info "等待 PostgreSQL 启动... (${waited}s/${max_wait}s)"
        fi
    done
    
    if [ $waited -ge $max_wait ]; then
        log_error "PostgreSQL 启动超时"
        systemctl status nta-postgres
        exit 1
    fi
    
    # 创建数据库和用户
    su - $SERVICE_USER -c "/opt/postgres/bin/createuser -s nta 2>/dev/null" || true
    su - $SERVICE_USER -c "/opt/postgres/bin/createdb -O nta nta 2>/dev/null" || true
    su - $SERVICE_USER -c "/opt/postgres/bin/psql -d nta -c \"ALTER USER nta WITH PASSWORD 'nta_password';\" 2>/dev/null" || true
    
    log_success "数据库初始化完成"
}

# 配置防火墙
configure_firewall() {
    log_info "配置防火墙..."
    
    if command -v ufw &> /dev/null; then
        ufw allow 8080/tcp comment 'NTA API Server' || true
        ufw allow 8090/tcp comment 'NTA Web UI' || true
        log_success "防火墙配置完成 (ufw)"
    else
        log_warn "未检测到 UFW 防火墙，请手动开放端口 8080, 8090"
    fi
}

# 启动服务
start_services() {
    log_info "启动服务..."
    
    # 启用服务自启动
    systemctl enable nta-postgres nta-redis nta-zookeeper nta-kafka nta-kafka-consumer nta-server
    
    # 按顺序启动服务
    systemctl start nta-postgres
    sleep 3
    
    systemctl start nta-redis
    sleep 2
    
    systemctl start nta-zookeeper
    sleep 5
    
    systemctl start nta-kafka
    sleep 10
    
    systemctl start nta-kafka-consumer
    sleep 3
    
    systemctl start nta-server
    sleep 3
    
    log_success "所有服务已启动"
}

# 检查服务状态
check_services() {
    log_info "检查服务状态..."
    echo ""
    
    services=("nta-postgres" "nta-redis" "nta-zookeeper" "nta-kafka" "nta-kafka-consumer" "nta-server")
    
    for service in "${services[@]}"; do
        if systemctl is-active --quiet $service; then
            echo -e "  ${GREEN}●${NC} $service: 运行中"
        else
            echo -e "  ${RED}●${NC} $service: 未运行"
        fi
    done
    
    echo ""
}

# 显示部署信息
show_info() {
    local server_ip=$(hostname -I | awk '{print $1}')
    
    echo ""
    echo "=========================================="
    echo "  NTA 系统部署完成！"
    echo "=========================================="
    echo ""
    echo "📊 访问地址:"
    echo "  - Web界面:  http://${server_ip}:8090"
    echo "  - API服务:  http://${server_ip}:8080"
    echo ""
    echo "🔑 默认账户:"
    echo "  - 用户名: admin"
    echo "  - 密码:   admin123"
    echo ""
    echo "📝 服务管理:"
    echo "  - 查看状态: systemctl status nta-server"
    echo "  - 查看日志: journalctl -u nta-server -f"
    echo "  - 重启服务: systemctl restart nta-server"
    echo "  - 停止服务: systemctl stop nta-server"
    echo ""
    echo "📂 安装目录:"
    echo "  - 程序目录: $INSTALL_DIR"
    echo "  - 数据目录: $DATA_DIR"
    echo "  - 日志目录: $LOG_DIR"
    echo ""
    echo "⚠️  重要提示:"
    echo "  1. 首次登录后请立即修改默认密码"
    echo "  2. 在 Web 界面配置 Zeek 监听网卡后启动探针"
    echo "  3. 配置路径: 系统管理 > 探针管理 > 内置探针"
    echo ""
    echo "🔧 常用命令:"
    echo "  - 查看所有服务: systemctl status 'nta-*'"
    echo "  - 卸载系统:     bash $SCRIPT_DIR/uninstall.sh"
    echo ""
    echo "✅ 预编译安装，总用时约 5-10 分钟"
    echo "=========================================="
}

# 主函数
main() {
    echo ""
    echo "╔════════════════════════════════════════╗"
    echo "║  NTA 网络流量分析系统 离线安装程序   ║"
    echo "║     Ubuntu 24.04 LTS 预编译版        ║"
    echo "╚════════════════════════════════════════╝"
    echo ""
    
    check_root
    detect_os
    check_requirements
    
    echo ""
    read -p "是否继续安装? (y/n): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        log_info "安装已取消"
        exit 0
    fi
    
    install_system_deps
    create_user
    create_directories
    install_postgres
    install_redis
    install_kafka
    install_zeek
    install_nta
    create_services
    init_database
    configure_firewall
    start_services
    sleep 5
    check_services
    show_info
    
    log_success "安装完成! 🎉"
}

main "$@"