@load ./main

module LateralMovement;

export {
    const auth_fail_threshold = 5 &redef;
    const auth_interval = 5min &redef;
    const pth_window = 1hr &redef;
}

global ntlm_hashes: table[string] of set[addr] &create_expire=pth_window;
global smb_auth_fails: table[addr, addr] of count &create_expire=10min &default=0;
global rdp_auth_fails: table[addr, addr] of count &create_expire=10min &default=0;
global ssh_auth_fails: table[addr, addr] of count &create_expire=10min &default=0;

event zeek_init()
{
    local r1: SumStats::Reducer = [$stream="smb.auth.fail",
                                    $apply=set(SumStats::SUM)];
    
    local r2: SumStats::Reducer = [$stream="rdp.auth.fail",
                                    $apply=set(SumStats::SUM)];

    local r3: SumStats::Reducer = [$stream="ssh.auth.fail",
                                    $apply=set(SumStats::SUM)];

    SumStats::create([$name="detect-auth-bruteforce",
                      $epoch=auth_interval,
                      $reducers=set(r1, r2, r3),
                      $threshold_val(key: SumStats::Key, result: SumStats::Result) =
                      {
                          local smb_fails = result["smb.auth.fail"]$sum;
                          local rdp_fails = result["rdp.auth.fail"]$sum;
                          local ssh_fails = result["ssh.auth.fail"]$sum;
                          return smb_fails + rdp_fails + ssh_fails;
                      },
                      $threshold=auth_fail_threshold,
                      $threshold_crossed(key: SumStats::Key, result: SumStats::Result) =
                      {
                          local attacker = key$host;
                          local smb_fails = result["smb.auth.fail"]$sum;
                          local rdp_fails = result["rdp.auth.fail"]$sum;
                          local ssh_fails = result["ssh.auth.fail"]$sum;
                          local total = smb_fails + rdp_fails + ssh_fails;

                          local protocol = "";
                          local note_type = Lateral_Auth_Anomaly;
                          
                          if (smb_fails >= auth_fail_threshold) {
                              protocol = "SMB";
                              note_type = SMB_Bruteforce_Detected;
                          } else if (rdp_fails >= auth_fail_threshold) {
                              protocol = "RDP";
                              note_type = RDP_Bruteforce_Detected;
                          } else if (ssh_fails >= auth_fail_threshold) {
                              protocol = "SSH";
                          }

                          NOTICE([$note=note_type,
                                  $src=attacker,
                                  $msg=fmt("检测到%s暴力破解: %s 失败认证 %d 次",
                                          protocol, attacker, total),
                                  $sub=fmt("SMB:%d RDP:%d SSH:%d", smb_fails, rdp_fails, ssh_fails),
                                  $identifier=cat(attacker, protocol)]);

                          local info: Info = [
                              $ts=network_time(),
                              $orig_h=attacker,
                              $resp_h=0.0.0.0,
                              $attack_type=fmt("%s_BRUTEFORCE", protocol),
                              $severity="CRITICAL",
                              $description=fmt("%s暴力破解攻击", protocol),
                              $evidence=fmt("失败次数:%d", total)
                          ];
                          Log::write(LateralMovement::LOG, info);
                      }]);
}

event ntlm_authenticate(c: connection, request: NTLM::Authenticate)
{
    if (!is_lateral_movement(c))
        return;

    if (!request?$response)
        return;

    local hash_value = request$response;
    local client_ip = c$id$orig_h;

    if (hash_value !in ntlm_hashes)
        ntlm_hashes[hash_value] = set();

    add ntlm_hashes[hash_value][client_ip];

    if (|ntlm_hashes[hash_value]| >= 3) {
        local hosts = "";
        for (host in ntlm_hashes[hash_value]) {
            hosts = fmt("%s%s ", hosts, host);
        }

        NOTICE([$note=PTH_Attack_Detected,
                $conn=c,
                $msg=fmt("检测到Pass-the-Hash攻击: 相同NTLM Hash在%d台主机上使用",
                        |ntlm_hashes[hash_value]|),
                $sub=fmt("涉及主机: %s", hosts),
                $identifier=cat(hash_value)]);

        local info: Info = [
            $ts=network_time(),
            $uid=c$uid,
            $orig_h=c$id$orig_h,
            $orig_p=c$id$orig_p,
            $resp_h=c$id$resp_h,
            $resp_p=c$id$resp_p,
            $attack_type="PASS_THE_HASH",
            $severity="CRITICAL",
            $description="Pass-the-Hash攻击",
            $evidence=fmt("Hash重用于%d台主机", |ntlm_hashes[hash_value]|)
        ];
        Log::write(LateralMovement::LOG, info);
    }
}

event smb2_tree_connect_response(c: connection, hdr: SMB2::Header, response: SMB2::TreeConnectResponse)
{
    if (!is_lateral_movement(c))
        return;

    if (hdr$status != 0) {
        smb_auth_fails[c$id$orig_h, c$id$resp_h] += 1;
        
        SumStats::observe("smb.auth.fail",
                         [$host=c$id$orig_h],
                         [$num=1]);
    }
}

event rdp_connect_request(c: connection, cookie: string)
{
    if (!is_lateral_movement(c))
        return;

    SumStats::observe("rdp.auth.fail",
                     [$host=c$id$orig_h],
                     [$num=0]);
}

event rdp_client_security_data(c: connection, data: RDP::ClientSecurityData)
{
    if (!is_lateral_movement(c))
        return;
}

event ssh_auth_failed(c: connection, authenticated_with: string)
{
    if (!is_lateral_movement(c))
        return;

    ssh_auth_fails[c$id$orig_h, c$id$resp_h] += 1;
    
    SumStats::observe("ssh.auth.fail",
                     [$host=c$id$orig_h],
                     [$num=1]);
}

event kerberos_as_request(c: connection, msg: Kerberos::AS_Request)
{
    if (!is_lateral_movement(c))
        return;
}

event kerberos_tgs_request(c: connection, msg: Kerberos::TGS_Request)
{
    if (!is_lateral_movement(c))
        return;

    if (msg?$ticket && msg$ticket?$realm) {
        local current_hour = double_to_count(network_time()) % 86400 / 3600;
        
        if (current_hour < 6 || current_hour > 22) {
            NOTICE([$note=Lateral_Auth_Anomaly,
                    $conn=c,
                    $msg=fmt("异常时间Kerberos认证: %s -> %s",
                            c$id$orig_h, c$id$resp_h),
                    $sub=fmt("时间: %d点", current_hour),
                    $identifier=cat(c$id$orig_h, c$id$resp_h)]);
        }
    }
}
