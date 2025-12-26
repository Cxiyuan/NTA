package threatintel

import (
	"strings"

	"github.com/Cxiyuan/NTA/pkg/models"
)

var (
	threatTypeMap = map[string]string{
		"botnet_cc":         "僵尸网络",
		"payload_delivery":  "恶意分发",
		"c2":                "僵尸网络",
		"command_control":   "僵尸网络",
	}

	malwareKeywords = map[string]string{
		"rat":           "远控木马",
		"remote access": "远控木马",
		"meterpreter":   "渗透工具",
		"stealer":       "窃密木马",
		"stealc":        "窃密木马",
		"lumma":         "窃密木马",
		"redline":       "窃密木马",
		"miner":         "挖矿木马",
		"cryptominer":   "挖矿木马",
		"xmrig":         "挖矿木马",
		"ransomware":    "勒索软件",
		"locker":        "勒索软件",
		"trojan":        "木马病毒",
		"backdoor":      "后门木马",
		"phishing":      "钓鱼攻击",
		"clearfake":     "钓鱼攻击",
		"fakeupdates":   "钓鱼攻击",
		"apt":           "APT组织",
		"muddywater":    "APT组织",
		"lazarus":       "APT组织",
		"shadowpad":     "间谍木马",
		"spy":           "间谍木马",
		"espionage":     "间谍木马",
		"gh0st":         "远控木马",
		"ghost":         "远控木马",
		"cobalt":        "渗透工具",
		"metasploit":    "渗透工具",
		"sectop":        "远控木马",
		"hook":          "窃密木马",
		"adaptix":       "远控木马",
	}

	severityLabels = map[string]string{
		"critical": "高危威胁",
		"high":     "高危威胁",
		"medium":   "恶意地址",
		"low":      "可疑地址",
	}
)

func GetThreatLabel(intel *models.ThreatIntel) string {
	desc := strings.ToLower(intel.Description)
	source := strings.ToLower(intel.Source)
	tags := strings.ToLower(intel.Tags)

	combined := desc + " " + tags + " " + source

	for keyword, label := range threatTypeMap {
		if strings.Contains(combined, keyword) {
			return label
		}
	}

	for keyword, label := range malwareKeywords {
		if strings.Contains(combined, keyword) {
			return label
		}
	}

	if label, ok := severityLabels[intel.Severity]; ok {
		return label
	}

	return "恶意地址"
}
