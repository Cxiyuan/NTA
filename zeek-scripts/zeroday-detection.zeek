@load ./main

module ZeroDay;

export {
    redef enum Notice::Type += {
        Protocol_Violation,
        Oversized_Packet,
        Shellcode_NOP_Sled,
        Abnormal_Protocol_Behavior,
        Heap_Spray_Detected,
        Format_String_Attack
    };

    type Protocol_Baseline: record {
        avg_packet_size: double &default=0.0;
        std_deviation: double &default=0.0;
        packet_count: count &default=0;
    };

    global protocol_baselines: table[port] of Protocol_Baseline
        &create_expire=24hr;
}

event new_packet(c: connection, p: pkt_hdr) {
    if (p$ip$len > 8192) {
        NOTICE([$note=Oversized_Packet,
                $conn=c,
                $msg=fmt("异常大小数据包: %d bytes", p$ip$len),
                $sub="可能的缓冲区溢出攻击"]);
    }

    local port = c$id$resp_p;
    if (port !in protocol_baselines) {
        protocol_baselines[port] = Protocol_Baseline();
    }
    
    local baseline = protocol_baselines[port];
    local old_avg = baseline$avg_packet_size;
    local n = baseline$packet_count;
    
    baseline$packet_count += 1;
    baseline$avg_packet_size = (old_avg * n + p$ip$len) / (n + 1);
    
    if (n > 100) {
        local deviation = (p$ip$len - baseline$avg_packet_size);
        if (deviation < 0)
            deviation = -deviation;
        
        if (deviation > baseline$avg_packet_size * 3) {
            NOTICE([$note=Abnormal_Protocol_Behavior,
                    $conn=c,
                    $msg=fmt("数据包大小异常: %d bytes (平均%.0f bytes)",
                            p$ip$len, baseline$avg_packet_size),
                    $sub=fmt("端口%d的异常流量", port)]);
        }
    }
}

event tcp_packet(c: connection, is_orig: bool, flags: string,
                 seq: count, ack: count, len: count, payload: string) {
    
    if (len == 0)
        return;

    local nop_count = 0;
    local i = 0;
    
    while (i < |payload| && i < 1000) {
        if (payload[i:i+1] == "\x90" || payload[i:i+1] == "\x00") {
            nop_count += 1;
            if (nop_count >= 100) {
                NOTICE([$note=Shellcode_NOP_Sled,
                        $conn=c,
                        $msg=fmt("检测到NOP滑坡 (长度=%d)", nop_count),
                        $sub="可能的shellcode注入攻击"]);
                break;
            }
        } else {
            nop_count = 0;
        }
        i += 1;
    }

    if (|payload| >= 20) {
        if (/\x31\xc0\x50\x68/ in payload || 
            /\xeb.\x5e\x31\xc0/ in payload ||
            /\x31\xdb\x31\xc9\x31\xd2/ in payload) {
            
            NOTICE([$note=Shellcode_NOP_Sled,
                    $conn=c,
                    $msg="检测到shellcode特征码",
                    $sub="x86汇编指令特征"]);
        }
    }

    local format_string_patterns = vector(
        "%s%s%s%s",
        "%n%n%n",
        "%p%p%p%p",
        "AAAA%08x"
    );
    
    for (idx in format_string_patterns) {
        local pattern = format_string_patterns[idx];
        if (pattern in payload) {
            NOTICE([$note=Format_String_Attack,
                    $conn=c,
                    $msg="检测到格式化字符串攻击特征",
                    $sub=fmt("模式: %s", pattern)]);
        }
    }

    if (|payload| > 1000) {
        local spray_patterns = vector(
            "\x0c\x0c\x0c\x0c",
            "\x0d\x0d\x0d\x0d",
            "AAAA",
            "\x41\x41\x41\x41"
        );
        
        for (idx in spray_patterns) {
            local pattern = spray_patterns[idx];
            local pattern_count = 0;
            local j = 0;
            
            while (j < |payload| - 4) {
                if (payload[j:j+4] == pattern) {
                    pattern_count += 1;
                    if (pattern_count > 10) {
                        NOTICE([$note=Heap_Spray_Detected,
                                $conn=c,
                                $msg=fmt("检测到堆喷射特征 (重复模式%d次)", pattern_count),
                                $sub="可能的内存破坏攻击"]);
                        break;
                    }
                }
                j += 4;
            }
        }
    }
}

event udp_packet(c: connection, is_orig: bool, payload: string) {
    if (|payload| == 0)
        return;

    if (|payload| > 512 && c$id$resp_p == 53/udp) {
        NOTICE([$note=Abnormal_Protocol_Behavior,
                $conn=c,
                $msg=fmt("DNS请求异常大小: %d bytes", |payload|),
                $sub="可能的DNS放大攻击或隧道"]);
    }
}

event connection_established(c: connection) {
    if (!c?$service || |c$service| == 0) {
        if (c$id$resp_p !in [80/tcp, 443/tcp, 22/tcp, 3389/tcp]) {
            NOTICE([$note=Abnormal_Protocol_Behavior,
                    $conn=c,
                    $msg=fmt("无法识别的协议: 端口%d", c$id$resp_p),
                    $sub="可能的非标准服务或后门"]);
        }
    }
}

event zeek_init() {
    print "0day漏洞利用检测模块已加载";
}
