#!/bin/bash

set -e

ZEEK_IFACE="${ZEEK_IFACE:-eth0}"
ZEEK_LOG_DIR="${ZEEK_LOG_DIR:-/var/spool/zeek}"
CAP_AGENT_DIR="/opt/cap-agent"

start_zeek() {
    echo "Starting Zeek on interface: $ZEEK_IFACE"
    mkdir -p "$ZEEK_LOG_DIR"
    zeekctl deploy
    zeekctl status
}

start_backend() {
    echo "Starting Backend API server..."
    cd "$CAP_AGENT_DIR/backend"
    python3 app.py &
}

start_analyzer() {
    echo "Starting Analyzer..."
    cd "$CAP_AGENT_DIR"
    python3 analyzer/detector.py &
}

case "$1" in
    zeek)
        start_zeek
        tail -f /dev/null
        ;;
    backend)
        start_backend
        tail -f /dev/null
        ;;
    analyzer)
        start_analyzer
        tail -f /dev/null
        ;;
    all)
        start_zeek
        sleep 5
        start_backend
        start_analyzer
        tail -f /dev/null
        ;;
    *)
        exec "$@"
        ;;
esac
