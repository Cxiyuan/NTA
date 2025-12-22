#!/bin/bash

set -e

echo "==================================="
echo "Zeek横向渗透检测探针部署脚本"
echo "==================================="

ZEEK_DIR="/usr/local/zeek"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/.."
ZEEK_SCRIPTS_DIR="$SCRIPT_DIR/zeek-scripts"
INSTALL_DIR="/opt/cap_agent"

check_root() {
    if [ "$EUID" -ne 0 ]; then
        echo "错误: 请使用root权限运行此脚本"
        exit 1
    fi
}

check_zeek_installed() {
    echo "[1/6] 检查Zeek安装状态..."
    
    if command -v zeek &> /dev/null; then
        ZEEK_VERSION=$(zeek --version | head -n1)
        echo "✓ Zeek已安装: $ZEEK_VERSION"
        ZEEK_BIN=$(which zeek)
        ZEEK_DIR=$(dirname $(dirname $ZEEK_BIN))
    else
        echo "✗ Zeek未安装，开始安装..."
        install_zeek
    fi
}

install_zeek() {
    if [ -f /etc/redhat-release ]; then
        echo "检测到CentOS/RHEL系统"
        yum install -y epel-release
        yum install -y cmake make gcc gcc-c++ flex bison libpcap-devel openssl-devel python3 python3-pip swig zlib-devel
        
        cd /tmp
        wget https://download.zeek.org/zeek-5.0.0.tar.gz
        tar -xzf zeek-5.0.0.tar.gz
        cd zeek-5.0.0
        ./configure --prefix=/usr/local/zeek
        make -j$(nproc)
        make install
        
        echo 'export PATH=/usr/local/zeek/bin:$PATH' >> /etc/profile.d/zeek.sh
        source /etc/profile.d/zeek.sh
        
    elif [ -f /etc/debian_version ]; then
        echo "检测到Debian/Ubuntu系统"
        apt-get update
        apt-get install -y cmake make gcc g++ flex bison libpcap-dev libssl-dev python3 python3-pip swig zlib1g-dev
        
        cd /tmp
        wget https://download.zeek.org/zeek-5.0.0.tar.gz
        tar -xzf zeek-5.0.0.tar.gz
        cd zeek-5.0.0
        ./configure --prefix=/usr/local/zeek
        make -j$(nproc)
        make install
        
        echo 'export PATH=/usr/local/zeek/bin:$PATH' >> /etc/profile
        source /etc/profile
    fi
    
    echo "✓ Zeek安装完成"
}

install_python_deps() {
    echo "[2/6] 安装Python依赖..."
    
    pip3 install --upgrade pip
    pip3 install pyyaml
    
    echo "✓ Python依赖安装完成"
}

deploy_scripts() {
    echo "[3/6] 部署Zeek检测脚本..."
    
    mkdir -p $INSTALL_DIR
    cp -r $SCRIPT_DIR/* $INSTALL_DIR/
    
    ZEEK_SITE_DIR="$ZEEK_DIR/share/zeek/site"
    mkdir -p $ZEEK_SITE_DIR/lateral-movement
    
    cp $ZEEK_SCRIPTS_DIR/*.zeek $ZEEK_SITE_DIR/lateral-movement/
    
    if ! grep -q "lateral-movement" $ZEEK_SITE_DIR/local.zeek 2>/dev/null; then
        echo "@load ./lateral-movement/main" >> $ZEEK_SITE_DIR/local.zeek
        echo "✓ 已添加检测模块到local.zeek"
    fi
    
    echo "✓ 脚本部署完成"
}

configure_zeek() {
    echo "[4/6] 配置Zeek..."
    
    ZEEK_CONFIG_DIR="$ZEEK_DIR/etc"
    
    read -p "请输入监听网卡名称 (默认: eth0): " INTERFACE
    INTERFACE=${INTERFACE:-eth0}
    
    if [ ! -f $ZEEK_CONFIG_DIR/node.cfg.bak ]; then
        cp $ZEEK_CONFIG_DIR/node.cfg $ZEEK_CONFIG_DIR/node.cfg.bak
    fi
    
    cat > $ZEEK_CONFIG_DIR/node.cfg <<EOF
[zeek]
type=standalone
host=localhost
interface=$INTERFACE
EOF
    
    if [ ! -f $ZEEK_CONFIG_DIR/networks.cfg.bak ]; then
        cp $ZEEK_CONFIG_DIR/networks.cfg $ZEEK_CONFIG_DIR/networks.cfg.bak 2>/dev/null || true
    fi
    
    cat > $ZEEK_CONFIG_DIR/networks.cfg <<EOF
10.0.0.0/8          Private IP space
172.16.0.0/12       Private IP space
192.168.0.0/16      Private IP space
EOF
    
    echo "✓ Zeek配置完成 (监听网卡: $INTERFACE)"
}

setup_systemd() {
    echo "[5/6] 配置系统服务..."
    
    cat > /etc/systemd/system/zeek-lateral-detector.service <<EOF
[Unit]
Description=Zeek Lateral Movement Detector
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=$INSTALL_DIR
ExecStart=/usr/local/zeek/bin/zeekctl deploy
ExecReload=/usr/local/zeek/bin/zeekctl restart
ExecStop=/usr/local/zeek/bin/zeekctl stop
Restart=on-failure
RestartSec=10

[Install]
WantedBy=multi-user.target
EOF
    
    systemctl daemon-reload
    
    echo "✓ 系统服务配置完成"
}

start_zeek() {
    echo "[6/6] 启动Zeek探针..."
    
    $ZEEK_DIR/bin/zeekctl install
    $ZEEK_DIR/bin/zeekctl deploy
    
    sleep 3
    
    if $ZEEK_DIR/bin/zeekctl status | grep -q "running"; then
        echo "✓ Zeek探针启动成功"
    else
        echo "✗ Zeek探针启动失败，请检查配置"
        $ZEEK_DIR/bin/zeekctl status
        exit 1
    fi
}

show_usage() {
    echo ""
    echo "==================================="
    echo "部署完成！"
    echo "==================================="
    echo ""
    echo "常用命令:"
    echo "  启动探针: zeekctl start"
    echo "  停止探针: zeekctl stop"
    echo "  重启探针: zeekctl restart"
    echo "  查看状态: zeekctl status"
    echo "  查看日志: tail -f /var/log/zeek/current/lateral_movement.log"
    echo ""
    echo "检测日志位置:"
    echo "  告警日志: /var/log/zeek/current/notice.log"
    echo "  横向移动: /var/log/zeek/current/lateral_movement.log"
    echo "  连接日志: /var/log/zeek/current/conn.log"
    echo ""
    echo "Python分析引擎:"
    echo "  实时分析: zeek-cut < /var/log/zeek/current/conn.log | python3 $INSTALL_DIR/analyzer/detector.py"
    echo ""
}

main() {
    check_root
    check_zeek_installed
    install_python_deps
    deploy_scripts
    configure_zeek
    setup_systemd
    start_zeek
    show_usage
}

main
