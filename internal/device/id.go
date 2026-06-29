package device

import (
	"errors"
	"net"
	"strings"
)

func GetMACAddress() (string, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}

	for _, iface := range interfaces {
		if iface.Flags&net.FlagLoopback != 0 || iface.Flags&net.FlagUp == 0 {
			continue
		}
		if len(iface.HardwareAddr) == 0 {
			continue
		}
		mac := iface.HardwareAddr.String()
		return strings.ReplaceAll(mac, ":", "-"), nil
	}

	return "", errors.New("no valid network interface found")
}
