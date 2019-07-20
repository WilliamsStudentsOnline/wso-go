package lib

import "net"

var schoolSubnet = &net.IPNet{
	IP: net.ParseIP("137.165.0.0"),
	Mask: net.CIDRMask(16, 32),
}

var localSubnet = &net.IPNet{
	IP: net.ParseIP("192.168.0.0"),
	Mask: net.CIDRMask(24, 32),
}

func OnCampusIP(ipString string) bool {
	ip := net.ParseIP(ipString)

	if ip == nil {
		return false
	}

	return schoolSubnet.Contains(ip) || localSubnet.Contains(ip)
}
