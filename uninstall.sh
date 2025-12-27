#!/bin/bash
# NTA 卸载脚本
# 版本: v2.0.0

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

INSTALL_DIR="/opt/nta"
DATA_DIR="/var/lib/nta"
LOG_DIR="/var/log/nta"
SERVICE_USER="nta"

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

check_root() {
    if [ "$EUID" -ne 0 ]; then
        log_error "请使用root权限运行此脚本"
        exit 1
    fi
}

stop_services() {
    log_info "停止所有 NTA 服务..."
    
    systemctl stop nta-server || true
    systemctl stop nta-kafka-consumer || true
    systemctl stop nta-zeek || true
    systemctl stop nta-kafka || true
    systemctl stop nta-zookeeper || true
    systemctl stop nta-redis || true
    systemctl stop nta-postgres || true
    
    log_success "服务已停止"
}

disable_services() {
    log_info "禁用服务自启动..."
    
    systemctl disable nta-server || true
    systemctl disable nta-kafka-consumer || true
    systemctl disable nta-zeek || true
    systemctl disable nta-kafka || true
    systemctl disable nta-zookeeper || true
    systemctl disable nta-redis || true
    systemctl disable nta-postgres || true
    
    log_success "服务自启动已禁用"
}

remove_services() {
    log_info "删除 systemd 服务文件..."
    
    rm -f /etc/systemd/system/nta-server.service
    rm -f /etc/systemd/system/nta-kafka-consumer.service
    rm -f /etc/systemd/system/nta-zeek.service
    rm -f /etc/systemd/system/nta-kafka.service
    rm -f /etc/systemd/system/nta-zookeeper.service
    rm -f /etc/systemd/system/nta-redis.service
    rm -f /etc/systemd/system/nta-postgres.service
    
    systemctl daemon-reload
    
    log_success "服务文件已删除"
}

remove_files() {
    log_info "删除程序文件..."
    
    echo ""
    read -p "是否删除数据文件 (包含数据库、日志等)? (y/n): " -n 1 -r
    echo
    
    rm -rf $INSTALL_DIR
    rm -rf /opt/postgres
    rm -rf /opt/redis
    rm -rf /opt/kafka
    rm -rf /opt/zeek
    
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        rm -rf $DATA_DIR
        rm -rf $LOG_DIR
        log_warn "数据文件已删除"
    else
        log_info "保留数据文件: $DATA_DIR"
    fi
    
    log_success "程序文件已删除"
}

remove_user() {
    log_info "删除服务用户..."
    
    if id "$SERVICE_USER" &>/dev/null; then
        userdel $SERVICE_USER || true
        log_success "用户 $SERVICE_USER 已删除"
    fi
}

main() {
    echo ""
    echo "╔════════════════════════════════════════╗"
    echo "║      NTA 卸载程序                    ║"
    echo "╚════════════════════════════════════════╝"
    echo ""
    
    check_root
    
    echo ""
    log_warn "此操作将卸载 NTA 系统的所有组件"
    read -p "是否继续卸载? (y/n): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        log_info "卸载已取消"
        exit 0
    fi
    
    stop_services
    disable_services
    remove_services
    remove_files
    remove_user
    
    echo ""
    log_success "NTA 卸载完成！"
    echo ""
}

main "$@"
