@load ./main

module LateralMovement;

export {
    const psexec_pipes: set[string] = {
        "\\pipe\\svcctl",
        "\\pipe\\atsvc",
        "\\pipe\\winreg",
        "\\ADMIN$",
        "\\C$",
        "\\IPC$"
    } &redef;

    const wmi_endpoints: set[string] = {
        "IWbemServices",
        "IWbemLevel1Login",
        "ISystemActivator",
        "IRemUnknown2"
    } &redef;
}

global smb_write_admin: table[addr, addr] of count &create_expire=5min &default=0;
global dce_rpc_calls: table[addr, addr] of set[string] &create_expire=10min;

event smb2_write_request(c: connection, hdr: SMB2::Header, file_id: count, offset: count, write_len: count)
{
    if (!is_lateral_movement(c))
        return;
}

event smb1_message(c: connection, hdr: SMB1::Header, is_orig: bool)
{
    if (!is_lateral_movement(c))
        return;

    if (!is_orig)
        return;
}

event file_over_new_connection(f: fa_file, c: connection, is_orig: bool)
{
    if (!is_lateral_movement(c))
        return;

    if (c$id$resp_p != 445/tcp)
        return;

    if (f?$source && /ADMIN\$/ in f$source || /C\$/ in f$source) {
        smb_write_admin[c$id$orig_h, c$id$resp_h] += 1;

        if (smb_write_admin[c$id$orig_h, c$id$resp_h] >= 2) {
            NOTICE([$note=PSExec_Detected,
                    $conn=c,
                    $msg=fmt("检测到PSExec执行: %s -> %s",
                            c$id$orig_h, c$id$resp_h),
                    $sub=fmt("写入管理共享: %s", f$source),
                    $identifier=cat(c$id$orig_h, c$id$resp_h)]);

            local info: Info = [
                $ts=network_time(),
                $uid=c$uid,
                $orig_h=c$id$orig_h,
                $orig_p=c$id$orig_p,
                $resp_h=c$id$resp_h,
                $resp_p=c$id$resp_p,
                $attack_type="PSEXEC",
                $severity="CRITICAL",
                $description="PSExec远程执行",
                $evidence=fmt("写入: %s", f$source)
            ];
            Log::write(LateralMovement::LOG, info);
        }
    }
}

event dce_rpc_request(c: connection, fid: count, opnum: count, stub_len: count)
{
    if (!is_lateral_movement(c))
        return;
}

event dce_rpc_bind(c: connection, uuid: string)
{
    if (!is_lateral_movement(c))
        return;

    local is_wmi = F;
    local endpoint_name = "";

    if (uuid == "00000000-0000-0000-c000-000000000046") {
        endpoint_name = "IUnknown";
    }
    else if (uuid == "000001a0-0000-0000-c000-000000000046") {
        endpoint_name = "ISystemActivator";
        is_wmi = T;
    }
    else if (uuid == "9556dc99-828c-11cf-a37e-00aa003240c7") {
        endpoint_name = "IWbemServices";
        is_wmi = T;
    }
    else if (uuid == "f309ad18-d86a-11d0-a075-00c04fb68820") {
        endpoint_name = "IWbemLevel1Login";
        is_wmi = T;
    }
    else if (uuid == "423ec01e-2e35-11d2-b604-00104b703efd") {
        endpoint_name = "IWbemContext";
        is_wmi = T;
    }

    if (is_wmi) {
        local key = c$id$orig_h;
        local target = c$id$resp_h;

        if ([key, target] !in dce_rpc_calls)
            dce_rpc_calls[key, target] = set();

        add dce_rpc_calls[key, target][endpoint_name];

        if (|dce_rpc_calls[key, target]| >= 2) {
            local endpoints = "";
            for (ep in dce_rpc_calls[key, target]) {
                endpoints = fmt("%s%s ", endpoints, ep);
            }

            NOTICE([$note=WMI_Execution_Detected,
                    $conn=c,
                    $msg=fmt("检测到WMI远程执行: %s -> %s",
                            c$id$orig_h, c$id$resp_h),
                    $sub=fmt("调用接口: %s", endpoints),
                    $identifier=cat(c$id$orig_h, c$id$resp_h, uuid)]);

            local info: Info = [
                $ts=network_time(),
                $uid=c$uid,
                $orig_h=c$id$orig_h,
                $orig_p=c$id$orig_p,
                $resp_h=c$id$resp_h,
                $resp_p=c$id$resp_p,
                $attack_type="WMI_EXECUTION",
                $severity="CRITICAL",
                $description="WMI远程执行",
                $evidence=fmt("接口: %s", endpoints)
            ];
            Log::write(LateralMovement::LOG, info);
        }
    }
}

event smb2_tree_connect_request(c: connection, hdr: SMB2::Header, path: string)
{
    if (!is_lateral_movement(c))
        return;

    for (pipe in psexec_pipes) {
        if (pipe in path) {
            local current_hour = double_to_count(network_time()) % 86400 / 3600;
            
            if (current_hour < 6 || current_hour > 22) {
                NOTICE([$note=Lateral_Exec_Detected,
                        $conn=c,
                        $msg=fmt("异常时间管道访问: %s -> %s",
                                c$id$orig_h, c$id$resp_h),
                        $sub=fmt("管道: %s, 时间: %d点", path, current_hour),
                        $identifier=cat(c$id$orig_h, c$id$resp_h, path)]);
            }
            break;
        }
    }
}

event connection_state_remove(c: connection)
{
    if (!is_lateral_movement(c))
        return;

    if (c$id$resp_p == 5985/tcp || c$id$resp_p == 5986/tcp) {
        if (c?$service && "http" in c$service) {
            NOTICE([$note=Lateral_Exec_Detected,
                    $conn=c,
                    $msg=fmt("检测到WinRM连接: %s -> %s:%d",
                            c$id$orig_h, c$id$resp_h, c$id$resp_p),
                    $sub="可能的PowerShell远程执行",
                    $identifier=cat(c$id$orig_h, c$id$resp_h)]);

            local info: Info = [
                $ts=network_time(),
                $uid=c$uid,
                $orig_h=c$id$orig_h,
                $orig_p=c$id$orig_p,
                $resp_h=c$id$resp_h,
                $resp_p=c$id$resp_p,
                $attack_type="WINRM",
                $severity="HIGH",
                $description="WinRM远程执行",
                $evidence=fmt("端口: %d", c$id$resp_p)
            ];
            Log::write(LateralMovement::LOG, info);
        }
    }
}
