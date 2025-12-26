@load ./main

module LateralMovement;

export {
    const scan_threshold = 20 &redef;
    const scan_interval = 5min &redef;
    const port_scan_threshold = 10 &redef;
}

global scanners: table[addr] of set[addr] &create_expire=10min;
global port_scanners: table[addr] of set[port] &create_expire=10min;

event zeek_init()
{
    local r1: SumStats::Reducer = [$stream="lateral.scan.hosts",
                                    $apply=set(SumStats::UNIQUE)];
    
    local r2: SumStats::Reducer = [$stream="lateral.scan.ports",
                                    $apply=set(SumStats::UNIQUE)];

    SumStats::create([$name="detect-lateral-scanner",
                      $epoch=scan_interval,
                      $reducers=set(r1, r2),
                      $threshold_val(key: SumStats::Key, result: SumStats::Result) =
                      {
                          return result["lateral.scan.hosts"]$num;
                      },
                      $threshold=scan_threshold,
                      $threshold_crossed(key: SumStats::Key, result: SumStats::Result) =
                      {
                          local scanner = key$host;
                          local target_count = result["lateral.scan.hosts"]$num;
                          local port_count = result["lateral.scan.ports"]$num;
                          
                          NOTICE([$note=Lateral_Scan_Detected,
                                  $src=scanner,
                                  $msg=fmt("检测到横向扫描: %s 扫描了 %d 个内网主机, %d 个端口",
                                          scanner, target_count, port_count),
                                  $sub=fmt("目标数量: %d, 端口数量: %d", target_count, port_count),
                                  $identifier=cat(scanner)]);

                          local info: Info = [
                              $ts=network_time(),
                              $orig_h=scanner,
                              $resp_h=0.0.0.0,
                              $attack_type="LATERAL_SCAN",
                              $severity="HIGH",
                              $description=fmt("横向扫描攻击: 扫描%d个主机", target_count),
                              $evidence=fmt("目标数:%d 端口数:%d", target_count, port_count)
                          ];
                          Log::write(LateralMovement::LOG, info);
                      }]);
}

event connection_state_remove(c: connection)
{
    if (!is_lateral_movement(c))
        return;

    if (c$id$resp_p !in sensitive_ports)
        return;

    local orig = c$id$orig_h;
    local resp = c$id$resp_h;
    local resp_port = c$id$resp_p;

    SumStats::observe("lateral.scan.hosts",
                     [$host=orig],
                     [$str=cat(resp)]);

    SumStats::observe("lateral.scan.ports",
                     [$host=orig],
                     [$str=cat(resp_port)]);

    if (orig !in scanners)
        scanners[orig] = set();
    add scanners[orig][resp];

    if (orig !in port_scanners)
        port_scanners[orig] = set();
    add port_scanners[orig][resp_port];
}

event new_connection(c: connection)
{
    if (!is_lateral_movement(c))
        return;

    if (c$id$resp_p in sensitive_ports) {
        if (c$id$orig_h in scanners && |scanners[c$id$orig_h]| >= 5) {
            local targets = "";
            for (target in scanners[c$id$orig_h]) {
                targets = fmt("%s%s ", targets, target);
            }
        }
    }
}

event connection_attempt(c: connection)
{
    if (!is_lateral_movement(c))
        return;

    if (c$id$resp_p == 445/tcp || c$id$resp_p == 3389/tcp) {
        if (c$id$orig_h in scanners && |scanners[c$id$orig_h]| >= scan_threshold/2) {
        }
    }
}
