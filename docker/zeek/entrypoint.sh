#!/bin/bash
set -e

INTERFACE=${ZEEK_INTERFACE:-eth0}
BPF_FILTER=${ZEEK_BPF_FILTER:-""}
SCRIPTS_DIR=${ZEEK_SCRIPTS_DIR:-/opt/zeek/share/zeek/site/nta}

cat > /opt/zeek/etc/node.cfg <<EOF
[zeek]
type=standalone
host=localhost
interface=${INTERFACE}
EOF

cat > /opt/zeek/etc/networks.cfg <<EOF
10.0.0.0/8      Private IP space
172.16.0.0/12   Private IP space
192.168.0.0/16  Private IP space
EOF

cat > /opt/zeek/share/zeek/site/local.zeek <<EOF
@load base/frameworks/notice
@load base/frameworks/logging
@load protocols/conn/known-hosts
@load protocols/dns
@load protocols/http
@load protocols/ssl
@load protocols/ssh
@load protocols/smtp
@load protocols/ftp

# Load NTA custom scripts
@load nta/main.zeek

# Enable JSON logging
@load policy/tuning/json-logs.zeek

# Extend log retention
redef Log::default_rotation_interval = 1 day;
redef Log::default_rotation_postprocessor_cmd = "gzip";

# Connection logging
redef Conn::default_capture_loss_threshold = 0.1;

# Notice framework
redef Notice::emailed_types += {
    Scan::Port_Scan,
    Scan::Address_Scan,
};
EOF

if [ -n "$BPF_FILTER" ]; then
    cat >> /opt/zeek/etc/node.cfg <<EOF
bpf_filter=${BPF_FILTER}
EOF
fi

/opt/zeek/bin/zeekctl install

echo "Starting Zeek on interface ${INTERFACE}..."
exec /opt/zeek/bin/zeekctl start --foreground
