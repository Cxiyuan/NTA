@load ./main

module EncryptedTraffic;

export {
    redef enum Notice::Type += {
        Suspicious_TLS_Fingerprint,
        C2_Beacon_Behavior,
        Data_Exfiltration_Detected,
        DNS_Tunnel_Detected,
        Unexpected_Internal_HTTPS,
        Reverse_Shell_Detected
    };

    type Connection_Stats: record {
        packet_count: count &default=0;
        total_bytes: count &default=0;
        packet_sizes: vector of count &default=vector();
        inter_arrival_times: vector of interval &default=vector();
        last_packet_time: time &optional;
    };

    global connection_stats: table[addr, addr, port] of Connection_Stats
        &create_expire=10min;

    global known_internal_https: set[addr] = {
    };

    global upload_tracker: table[addr] of count 
        &create_expire=1hr &default=0;
}

event ssl_established(c: connection) {
    if (!c?$ssl)
        return;

    if (Site::is_local_addr(c$id$orig_h) && 
        Site::is_local_addr(c$id$resp_h)) {
        
        if (c$id$resp_p == 443/tcp || c$id$resp_p == 8443/tcp) {
            if (c$id$resp_h !in known_internal_https) {
                NOTICE([$note=Unexpected_Internal_HTTPS,
                        $conn=c,
                        $msg=fmt("意外的内网HTTPS连接: %s -> %s",
                                c$id$orig_h, c$id$resp_h),
                        $sub="可能的横向移动隧道或C2通信"]);
            }
        }
    }

    if (c$ssl?$server_name) {
        local domain = c$ssl$server_name;
        
        local subdomain_count = 0;
        local i = 0;
        while (i < |domain|) {
            if (domain[i:i+1] == ".")
                subdomain_count += 1;
            i += 1;
        }
        
        if (subdomain_count > 5 || |domain| > 100) {
            NOTICE([$note=DNS_Tunnel_Detected,
                    $conn=c,
                    $msg=fmt("可疑长域名: %s", domain),
                    $sub=fmt("长度=%d, 子域名=%d", |domain|, subdomain_count)]);
        }

        if (/[a-z0-9]{30,}\./ in domain) {
            NOTICE([$note=Suspicious_TLS_Fingerprint,
                    $conn=c,
                    $msg=fmt("可疑DGA域名: %s", domain),
                    $sub="随机生成的域名特征"]);
        }
    }
}

event new_packet(c: connection, p: pkt_hdr) {
    local key = [c$id$orig_h, c$id$resp_h, c$id$resp_p];
    
    if (key !in connection_stats) {
        connection_stats[key] = Connection_Stats();
    }
    
    local stats = connection_stats[key];
    stats$packet_count += 1;
    stats$total_bytes += p$ip$len;
    stats$packet_sizes += p$ip$len;
    
    if (stats?$last_packet_time) {
        stats$inter_arrival_times += network_time() - stats$last_packet_time;
    }
    stats$last_packet_time = network_time();
}

event connection_state_remove(c: connection) {
    if (!c?$ssl)
        return;

    local key = [c$id$orig_h, c$id$resp_h, c$id$resp_p];
    
    if (key !in connection_stats)
        return;

    local stats = connection_stats[key];
    
    if (|stats$inter_arrival_times| >= 5) {
        local sum: interval = 0sec;
        for (i in stats$inter_arrival_times) {
            sum += stats$inter_arrival_times[i];
        }
        local avg_interval = sum / |stats$inter_arrival_times|;
        
        local variance = 0.0;
        for (i in stats$inter_arrival_times) {
            local diff = interval_to_double(stats$inter_arrival_times[i] - avg_interval);
            variance += diff * diff;
        }
        variance = variance / |stats$inter_arrival_times|;
        
        if (variance < 1.0 && avg_interval > 30sec) {
            NOTICE([$note=C2_Beacon_Behavior,
                    $conn=c,
                    $msg=fmt("检测到规律性信标通信: 间隔=%.1fs, 方差=%.2f",
                            interval_to_double(avg_interval), variance),
                    $sub="可能的C2 Beacon通信"]);
        }
    }

    if (c$duration > 10min && stats$total_bytes < 10240) {
        NOTICE([$note=C2_Beacon_Behavior,
                $conn=c,
                $msg=fmt("长连接小流量行为: 持续%.0fs, 流量%d bytes",
                        interval_to_double(c$duration), stats$total_bytes),
                $sub="可能的C2保活连接"]);
    }

    if (Site::is_local_addr(c$id$orig_h) && 
        !Site::is_local_addr(c$id$resp_h)) {
        
        local uploaded = c$orig$num_bytes_ip;
        local downloaded = c$resp$num_bytes_ip;
        
        upload_tracker[c$id$orig_h] += uploaded;
        
        if (downloaded > 0) {
            local upload_ratio = uploaded * 1.0 / downloaded;
            
            if (upload_ratio > 5.0 && uploaded > 1048576) {
                NOTICE([$note=Data_Exfiltration_Detected,
                        $conn=c,
                        $msg=fmt("异常上传流量: %d bytes (上传/下载比=%.1f)",
                                uploaded, upload_ratio),
                        $sub=fmt("目标: %s", c$ssl?$server_name ? 
                                c$ssl$server_name : cat(c$id$resp_h))]);
            }
        }
        
        if (upload_tracker[c$id$orig_h] > 104857600) {
            NOTICE([$note=Data_Exfiltration_Detected,
                    $conn=c,
                    $msg=fmt("大量数据上传: %s 上传了 %d MB",
                            c$id$orig_h, 
                            upload_tracker[c$id$orig_h] / 1048576),
                    $sub=fmt("目标域名: %s", 
                            c$ssl?$server_name ? c$ssl$server_name : "unknown")]);
        }
    }
}

global outbound_connections: table[addr] of set[port] 
    &create_expire=5min;

event connection_established(c: connection) {
    if (Site::is_local_addr(c$id$orig_h) && 
        !Site::is_local_addr(c$id$resp_h)) {
        
        local server_subnets: set[subnet] = set(
        );
        
        local is_server = F;
        for (subnet in server_subnets) {
            if (c$id$orig_h in subnet) {
                is_server = T;
                break;
            }
        }
        
        if (is_server) {
            local src = c$id$orig_h;
            local dst_port = c$id$resp_p;
            
            if (src !in outbound_connections) {
                outbound_connections[src] = set();
            }
            
            add outbound_connections[src][dst_port];
            
            if (dst_port == 4444/tcp || 
                dst_port == 5555/tcp || 
                dst_port == 1337/tcp) {
                
                NOTICE([$note=Reverse_Shell_Detected,
                        $conn=c,
                        $msg=fmt("检测到反弹shell: %s -> %s:%d",
                                c$id$orig_h, c$id$resp_h, dst_port),
                        $sub="服务器主动外连可疑端口"]);
            }
            
            if (|outbound_connections[src]| >= 5) {
                NOTICE([$note=Reverse_Shell_Detected,
                        $conn=c,
                        $msg=fmt("服务器异常外连行为: %s (连接%d个端口)",
                                src, |outbound_connections[src]|),
                        $sub="可能已被入侵"]);
            }
        }
    }
}

event zeek_init() {
    print "加密流量分析模块已加载";
}
