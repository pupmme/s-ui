package util

var InboundTypeWithLink = []string{
	"vmess", "vless", "trojan", "shadowsocks", "hysteria2", "wirenorm",
}

func LinkGenerator(config interface{}, inbound interface{}, hostname string) []string {
	return []string{}
}

func SlicesContains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func FillOutJson(inbound interface{}, hostname string) error {
	return nil
}
