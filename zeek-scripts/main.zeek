@load base/frameworks/notice
@load base/frameworks/sumstats
@load base/protocols/conn
@load base/protocols/smb
@load base/protocols/rdp
@load base/protocols/ssh
@load base/protocols/dce-rpc
@load base/protocols/ntlm
@load base/protocols/krb

@load ./lateral-scan
@load ./lateral-auth
@load ./lateral-exec
@load ./deep-inspection
@load ./encrypted-traffic
@load ./zeroday-detection
@load ./attack-chain
@load ./kafka-output

module LateralMovement;

export {
    redef enum Notice::Type += {
        Lateral_Scan_Detected,
        Lateral_Auth_Anomaly,
        Lateral_Exec_Detected,
        PTH_Attack_Detected,
        SMB_Bruteforce_Detected,
        RDP_Bruteforce_Detected,
        Kerberos_Golden_Ticket,
        WMI_Execution_Detected,
        PSExec_Detected
    };

    global sensitive_ports: set[port] = {
        135/tcp,  # RPC
        139/tcp,  # NetBIOS
        445/tcp,  # SMB
        3389/tcp, # RDP
        22/tcp,   # SSH
        5985/tcp, # WinRM HTTP
        5986/tcp  # WinRM HTTPS
    };

    global internal_networks: set[subnet] = {
        10.0.0.0/8,
        172.16.0.0/12,
        192.168.0.0/16
    };

    global log_lateral: event(rec: LateralMovement::Info);
}

type Info: record {
    ts: time &log;
    uid: string &log &optional;
    orig_h: addr &log;
    orig_p: port &log &optional;
    resp_h: addr &log;
    resp_p: port &log &optional;
    attack_type: string &log;
    severity: string &log;
    description: string &log;
    evidence: string &log &optional;
};

redef record connection += {
    lateral: Info &optional;
};

event zeek_init()
{
    Log::create_stream(LateralMovement::LOG, [$columns=Info, $ev=log_lateral, $path="lateral_movement"]);
    print "Zeek横向移动检测模块已加载";
}

function is_internal(ip: addr): bool
{
    for (net in internal_networks) {
        if (ip in net)
            return T;
    }
    return F;
}

function is_lateral_movement(c: connection): bool
{
    return is_internal(c$id$orig_h) && is_internal(c$id$resp_h);
}