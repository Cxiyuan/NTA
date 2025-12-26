FROM debian:bullseye-slim

ARG ZEEK_VERSION=6.0.3

RUN apt-get update && apt-get install -y \
    wget \
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
    libmaxminddb-dev \
    zlib1g-dev \
    libkrb5-dev \
    bsdmainutils \
    ca-certificates \
    swig \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /tmp

RUN wget https://download.zeek.org/zeek-${ZEEK_VERSION}.tar.gz && \
    tar -xzf zeek-${ZEEK_VERSION}.tar.gz && \
    cd zeek-${ZEEK_VERSION} && \
    ./configure --prefix=/opt/zeek && \
    make -j$(nproc) && \
    make install && \
    cd / && \
    rm -rf /tmp/zeek-${ZEEK_VERSION}*

ENV PATH="/opt/zeek/bin:${PATH}"

RUN groupadd -g 1000 zeek && \
    useradd -u 1000 -g zeek -s /bin/bash -m zeek

WORKDIR /opt/zeek

COPY scripts /opt/zeek/share/zeek/site/nta

RUN mkdir -p /opt/zeek/logs /opt/zeek/spool && \
    chown -R zeek:zeek /opt/zeek/logs /opt/zeek/spool

COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

USER zeek

EXPOSE 47760/tcp

ENTRYPOINT ["/entrypoint.sh"]
CMD ["zeekctl", "deploy"]