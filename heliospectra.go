package heliospectra

import "net"

const (
	// TCPPort is the TCP Port that Heliospectra LED fixtures listen on.
	TCPPort = 50630
	// UDPPort is the UDP Port that Heliospectra LED fixtures listen on.
	UDPPort = 50632
)

type DeviceInfo struct {
	MAC       string `xml:"MACAddress"`
	DHCP      bool
	IPAddr    net.IP `xml:"IPAddress"`
	NetMask   string
	Gateway   net.IP
	DNS1      net.IP
	DNS2      net.IP
	FwVersion string
	SerialNum string `xml:"SerialNr"`
}
