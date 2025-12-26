@load ./main

module AttackChain;

export {
    redef enum Notice::Type += {
        APT_Campaign_Detected,
        Multi_Stage_Attack,
        Kill_Chain_Progress
    };

    type AttackEvent: record {
        timestamp: time;
        event_type: string;
        source: addr;
        target: addr;
        details: string;
        severity: count;
    };

    type APT_Campaign: record {
        attacker_ip: addr;
        first_seen: time;
        stages_detected: set[string];
        victims: set[addr];
        c2_domains: set[string];
        iocs: vector of string;
        risk_score: count;
        last_activity: time;
    };

    global active_campaigns: table[addr] of APT_Campaign 
        &create_expire=7days;

    global attack_events: table[addr] of vector of AttackEvent
        &create_expire=24hr;
}

function log_attack_event(src: addr, dst: addr, event_type: string, 
                          details: string, severity: count) {
    if (src !in attack_events) {
        attack_events[src] = vector();
    }
    
    attack_events[src] += AttackEvent(
        $timestamp=network_time(),
        $event_type=event_type,
        $source=src,
        $target=dst,
        $details=details,
        $severity=severity
    );

    if (src !in active_campaigns) {
        active_campaigns[src] = APT_Campaign(
            $attacker_ip=src,
            $first_seen=network_time(),
            $stages_detected=set(),
            $victims=set(),
            $c2_domains=set(),
            $iocs=vector(),
            $risk_score=0,
            $last_activity=network_time()
        );
    }
    
    local campaign = active_campaigns[src];
    add campaign$stages_detected[event_type];
    add campaign$victims[dst];
    campaign$risk_score += severity;
    campaign$last_activity = network_time();
    
    if (|campaign$stages_detected| >= 2) {
        NOTICE([$note=Kill_Chain_Progress,
                $src=src,
                $msg=fmt("攻击链进展: %s (阶段%d/7)",
                        src, |campaign$stages_detected|),
                $sub=fmt("检测到: %s", event_type),
                $identifier=cat(src, event_type)]);
    }

    if (|campaign$stages_detected| >= 4) {
        NOTICE([$note=APT_Campaign_Detected,
                $src=src,
                $msg=fmt("确认APT攻击活动 (风险评分=%d)", campaign$risk_score),
                $sub=fmt("攻击阶段: %s", campaign$stages_detected),
                $identifier=cat(src)]);
        
        generate_apt_report(campaign);
    }
}

function generate_apt_report(campaign: APT_Campaign) {
    local report = fmt("========== APT攻击报告 ==========\n");
    report += fmt("攻击者IP: %s\n", campaign$attacker_ip);
    report += fmt("首次发现: %s\n", strftime("%Y-%m-%d %H:%M:%S", campaign$first_seen));
    report += fmt("最后活动: %s\n", strftime("%Y-%m-%d %H:%M:%S", campaign$last_activity));
    report += fmt("持续时间: %.0f 小时\n", 
                 interval_to_double(campaign$last_activity - campaign$first_seen) / 3600);
    report += fmt("攻击阶段数: %d\n", |campaign$stages_detected|);
    report += fmt("受害主机: %d 台\n", |campaign$victims|);
    report += fmt("风险评分: %d / 100\n", campaign$risk_score);
    report += "================================\n";
    
    print report;
}

function analyze_attack_pattern(src: addr): string {
    if (src !in attack_events || |attack_events[src]| < 3)
        return "UNKNOWN";

    local events = attack_events[src];
    local has_recon = F;
    local has_exploit = F;
    local has_lateral = F;
    local has_c2 = F;
    local has_exfil = F;

    for (idx in events) {
        local evt = events[idx];
        
        if (evt$event_type == "RECONNAISSANCE" || 
            evt$event_type == "LATERAL_SCAN")
            has_recon = T;
        else if (evt$event_type == "CREDENTIAL_ACCESS" || 
                 evt$event_type == "PTH_DETECTED")
            has_exploit = T;
        else if (evt$event_type == "LATERAL_MOVEMENT" || 
                 evt$event_type == "PSEXEC")
            has_lateral = T;
        else if (evt$event_type == "C2_BEACON")
            has_c2 = T;
        else if (evt$event_type == "EXFILTRATION")
            has_exfil = T;
    }

    if (has_recon && has_exploit && has_lateral && has_c2) {
        return "APT_FULL_CHAIN";
    } else if (has_recon && has_exploit) {
        return "TARGETED_ATTACK";
    } else if (has_lateral && has_c2) {
        return "POST_EXPLOITATION";
    } else if (has_recon) {
        return "RECONNAISSANCE";
    }
    
    return "SUSPICIOUS_ACTIVITY";
}

event Lateral_Scan_Detected(scanner: addr, target_count: count) {
    log_attack_event(scanner, 0.0.0.0, "RECONNAISSANCE", 
                    fmt("扫描%d台主机", target_count), 15);
}

event PTH_Attack_Detected(c: connection) {
    log_attack_event(c$id$orig_h, c$id$resp_h, "CREDENTIAL_ACCESS",
                    "Pass-the-Hash攻击", 30);
}

event PSExec_Detected(c: connection) {
    log_attack_event(c$id$orig_h, c$id$resp_h, "LATERAL_MOVEMENT",
                    "PSExec远程执行", 25);
}

event WMI_Execution_Detected(c: connection) {
    log_attack_event(c$id$orig_h, c$id$resp_h, "LATERAL_MOVEMENT",
                    "WMI远程执行", 25);
}

event C2_Beacon_Behavior(c: connection) {
    log_attack_event(c$id$orig_h, c$id$resp_h, "C2_BEACON",
                    "规律性信标通信", 20);
    
    if (c?$ssl && c$ssl?$server_name) {
        if (c$id$orig_h in active_campaigns) {
            add active_campaigns[c$id$orig_h]$c2_domains[c$ssl$server_name];
        }
    }
}

event Data_Exfiltration_Detected(c: connection, bytes: count) {
    log_attack_event(c$id$orig_h, c$id$resp_h, "EXFILTRATION",
                    fmt("数据渗出%d MB", bytes/1048576), 35);
}

event connection_state_remove(c: connection) {
    if (c$id$orig_h in active_campaigns) {
        local pattern = analyze_attack_pattern(c$id$orig_h);
        
        if (pattern == "APT_FULL_CHAIN") {
            NOTICE([$note=Multi_Stage_Attack,
                    $conn=c,
                    $msg=fmt("检测到完整APT攻击链: %s", c$id$orig_h),
                    $sub="侦察 -> 利用 -> 横向 -> C2 -> 渗出",
                    $identifier=cat(c$id$orig_h)]);
        }
    }
}

event zeek_init() {
    print "攻击链关联分析模块已加载";
}
