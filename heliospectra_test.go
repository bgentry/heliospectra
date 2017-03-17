package heliospectra

import (
	"encoding/xml"
	"net"
	"reflect"
	"testing"
)

func TestUnmarshalDeviceInfo(t *testing.T) {
	const payload = `<HelioDevice>
<MACAddress>64:1A:10:10:10:10</MACAddress><DHCP>true</DHCP><IPAddress>192.168.1.8</IPAddress><NetMask>255.255.255.0</NetMask><Gateway>192.168.1.1</Gateway><DNS1>192.168.1.1</DNS1><DNS2>0.0.0.0</DNS2><FwVersion>R2.2.25</FwVersion><SerialNr>fcaaaaaaaaaa</SerialNr></HelioDevice>`
	var di DeviceInfo
	if err := xml.Unmarshal([]byte(payload), &di); err != nil {
		t.Fatal(err)
	}
	if expectedMAC := "64:1A:10:10:10:10"; expectedMAC != di.MAC {
		t.Errorf("expected MAC % #x, got % #x", expectedMAC, di.MAC)
	}
	if !di.DHCP {
		t.Errorf("expected DHCP=true")
	}
	if expectedIP := net.IPv4(192, 168, 1, 8); !reflect.DeepEqual(expectedIP, di.IPAddr) {
		t.Errorf("expected IPAddr=%s, got %s", expectedIP, di.IPAddr)
	}
	if expectedMask := "255.255.255.0"; expectedMask != di.NetMask {
		t.Errorf("expected NetMask=%s, got %s", expectedMask, di.NetMask)
	}
	if expectedGw := net.IPv4(192, 168, 1, 1); !reflect.DeepEqual(expectedGw, di.Gateway) {
		t.Errorf("expected Gateway=%s, got %s", expectedGw, di.Gateway)
	}
	if expectedDNS1 := net.IPv4(192, 168, 1, 1); !reflect.DeepEqual(expectedDNS1, di.DNS1) {
		t.Errorf("expected DNS1=%s, got %s", expectedDNS1, di.DNS1)
	}
	if expectedDNS2 := net.IPv4(0, 0, 0, 0); !reflect.DeepEqual(expectedDNS2, di.DNS2) {
		t.Errorf("expected DNS2=%s, got %s", expectedDNS2, di.DNS2)
	}
	if expectedFW := "R2.2.25"; expectedFW != di.FwVersion {
		t.Errorf("expected FwVersion=%s, got %s", expectedFW, di.FwVersion)
	}
	if expectedSerial := "fcaaaaaaaaaa"; expectedSerial != di.SerialNum {
		t.Errorf("expected SerialNum=%s, got %s", expectedSerial, di.SerialNum)
	}
}
