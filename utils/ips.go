package utils

import (
	"net"
	"os"
)

var(
	internalIp string
)

func GetInternal() string {
	if internalIp!=""{
		return internalIp
	}

	addrs, err := net.InterfaceAddrs()
	if err != nil {
		os.Stderr.WriteString("Oops:" + err.Error())
		os.Exit(1)
	}
	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				internalIp=ipnet.IP.To4().String()
				return internalIp
			}
		}
	}
	return ""
}
