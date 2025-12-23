# Build timestamp: 2025-12-23
FROM ubuntu:latest

LABEL maintainer="cap-agent"
LABEL description="Cap Agent - Zeek-based Lateral Movement Detection Probe"

ENV ZEEK_VERSION=6.0.3
ENV PYTHONUNBUFFERED=1
ENV DEBIAN_FRONTEND=noninteractive

# Install build dependencies and runtime tools
RUN apt-get update && \
    apt-get install -y \
        cmake \
        make \
        gcc \
        g++ \
        flex \
        bison \
        libpcap-dev \
        libssl-dev \
        python3 \
        python3-pip \
        python3-dev \
        swig \
        zlib1g-dev \
        git \
        wget \
        curl \
        tcpdump \
        net-tools && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

# Build and install Zeek
WORKDIR /tmp
RUN wget https://download.zeek.org/zeek-${ZEEK_VERSION}.tar.gz && \
    tar -xzf zeek-${ZEEK_VERSION}.tar.gz && \
    cd zeek-${ZEEK_VERSION} && \
    ./configure --prefix=/opt/zeek && \
    make -j$(nproc) && \
    make install && \
    cd /tmp && \
    rm -rf zeek-${ZEEK_VERSION} zeek-${ZEEK_VERSION}.tar.gz

ENV PATH="/opt/zeek/bin:${PATH}"

WORKDIR /opt/cap-agent

# Install Python dependencies (cache layer)
COPY requirements.txt /opt/cap-agent/
RUN pip3 install --no-cache-dir --break-system-packages -r requirements.txt

COPY backend/requirements.txt /opt/cap-agent/backend-requirements.txt
RUN pip3 install --no-cache-dir --break-system-packages -r backend-requirements.txt

# Copy application code
COPY . /opt/cap-agent/

# Create directories and set permissions
RUN mkdir -p /opt/cap-agent/logs /opt/cap-agent/reports /var/spool/zeek && \
    chmod +x /opt/cap-agent/deploy/*.sh && \
    chmod +x /opt/cap-agent/analyzer/*.py && \
    chmod +x /opt/cap-agent/backend/*.py

# Configure Zeek to load custom scripts
RUN echo "@load /opt/cap-agent/zeek-scripts/main.zeek" >> /opt/zeek/share/zeek/site/local.zeek

EXPOSE 5000 5001

VOLUME ["/opt/cap-agent/logs", "/opt/cap-agent/reports", "/var/spool/zeek"]

COPY deploy/docker-entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
CMD ["all"]