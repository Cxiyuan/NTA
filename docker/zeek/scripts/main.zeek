@load base/frameworks/notice

module NTA;

export {
    redef enum Notice::Type += {
        Lateral_Movement_Detected,
        Suspicious_DNS_Query,
        Data_Exfiltration_Attempt,
        Abnormal_Connection_Pattern,
    };
}

# Load all NTA detection modules
@load nta/lateral-scan
@load nta/lateral-auth
@load nta/lateral-exec
@load nta/encrypted-traffic
@load nta/attack-chain
@load nta/deep-inspection
@load nta/zeroday-detection

event zeek_init()
{
    print "NTA detection modules loaded";
}

event zeek_done()
{
    print "NTA analysis complete";
}
