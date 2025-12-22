@load ./main
@load base/frameworks/files
@load base/files/hash-all-files

module DeepInspection;

export {
    redef enum Notice::Type += {
        High_Entropy_File,
        Shellcode_Detected,
        Malicious_Payload_Pattern,
        Suspicious_Certificate,
        JA3_Malware_Match
    };

    global malicious_ja3: table[string] of string = {
        ["a0e9f5d64349fb13191bc781f81f42e1"] = "Metasploit",
        ["6734f37431670b3ab4292b8f60f29984"] = "Trickbot",
        ["72a589da586844d7f0818ce684948eea"] = "Dridex",
        ["51c64c77e60f3980eea90869b68c58a8"] = "Cobalt Strike"
    };

    global malicious_patterns: pattern = 
        /sekurlsa::/ |
        /kerberos::/ |
        /lsadump::/ |
        /\x4d\x65\x74\x61\x73\x70\x6c\x6f\x69\x74/ |
        /\x90{50,}/;

    redef Files::hash_all_files = T;
}

function calculate_entropy(data: string): double {
    if (|data| == 0)
        return 0.0;

    local byte_counts: table[count] of count = table();
    local total = |data|;
    
    local i = 0;
    while (i < |data|) {
        local byte_val = bytestring_to_count(data[i:i+1]);
        if (byte_val !in byte_counts)
            byte_counts[byte_val] = 0;
        byte_counts[byte_val] += 1;
        i += 1;
    }
    
    local entropy = 0.0;
    for (byte_val in byte_counts) {
        local probability = byte_counts[byte_val] * 1.0 / total;
        if (probability > 0.0)
            entropy -= probability * log2(probability);
    }
    
    return entropy;
}

event file_over_new_connection(f: fa_file, c: connection, is_orig: bool) {
    if (c$id$resp_p != 445/tcp)
        return;

    Files::add_analyzer(f, Files::ANALYZER_MD5);
    Files::add_analyzer(f, Files::ANALYZER_SHA1);
    Files::add_analyzer(f, Files::ANALYZER_SHA256);
}

event file_state_remove(f: fa_file) {
    if (!f?$bof || |f$bof| < 100)
        return;

    local entropy = calculate_entropy(f$bof);
    
    if (entropy > 7.5 && f$total_bytes < 102400) {
        NOTICE([$note=High_Entropy_File,
                $f=f,
                $msg=fmt("高熵文件检测: 熵值=%.2f, 大小=%d", entropy, f$total_bytes),
                $sub="可能的加密或打包恶意软件"]);
    }

    if (malicious_patterns in f$bof) {
        NOTICE([$note=Malicious_Payload_Pattern,
                $f=f,
                $msg="检测到恶意工具特征码",
                $sub="文件包含Mimikatz或Metasploit特征"]);
    }

    if (|f$bof| >= 2 && f$bof[0:2] == "MZ") {
        local nop_count = 0;
        local i = 0;
        while (i < |f$bof| && i < 1000) {
            if (f$bof[i:i+1] == "\x90") {
                nop_count += 1;
                if (nop_count >= 50) {
                    NOTICE([$note=Shellcode_Detected,
                            $f=f,
                            $msg=fmt("检测到NOP滑坡 (长度=%d)", nop_count),
                            $sub="可能的shellcode注入"]);
                    break;
                }
            } else {
                nop_count = 0;
            }
            i += 1;
        }
    }
}

event ssl_established(c: connection) {
    if (!c?$ssl)
        return;

    if (c$ssl?$cert_chain && |c$ssl$cert_chain| > 0) {
        local cert = c$ssl$cert_chain[0];
        
        if (cert?$subject && cert?$issuer && cert$subject == cert$issuer) {
            if (Site::is_local_addr(c$id$orig_h) && Site::is_local_addr(c$id$resp_h)) {
                NOTICE([$note=Suspicious_Certificate,
                        $conn=c,
                        $msg="内网HTTPS使用自签名证书",
                        $sub=fmt("CN=%s", cert$subject)]);
            }
        }

        if (cert?$not_valid_after) {
            local days_left = (cert$not_valid_after - network_time()) / 1day;
            if (days_left < 7 && days_left > 0) {
                NOTICE([$note=Suspicious_Certificate,
                        $conn=c,
                        $msg=fmt("证书即将过期 (剩余%d天)", days_left),
                        $sub=fmt("目标: %s", c$id$resp_h)]);
            }
        }
    }
}

event zeek_init() {
    print "Zeek深度包检测模块已加载";
}
