package heliospectra

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"net"
	"time"
)

const (
	// TCPPort is the TCP Port that Heliospectra LED fixtures listen on.
	TCPPort = 50630
	// UDPPort is the UDP Port that Heliospectra LED fixtures listen on.
	UDPPort = 50632
)

type commandID uint8

const (
	// commandIDQuery is the command used to query for available devices.
	commandIDQuery commandID = 0
	// commandIDUnmute includes the device in a selective device query.
	commandIDUnmute commandID = 1
	// commandIDQueryUnmuted is the same as a Query, but only unmuted devices
	// shall send a response.
	commandIDQueryUnmuted commandID = 2
	// commandIDMute excludes the lamp from a selective device query. Only affects
	// QUERY_UNMUTED (commandIDQueryUnmuted) queries.
	commandIDMute commandID = 3
	// commandIDSet sets the device network configuration.
	commandIDSet commandID = 4
	// commandIDRestart restarts a device.
	commandIDRestart commandID = 5
	// commandIDInfoReply is sent by each lamp receiving a QUERY package. Each
	// will send (possibly multiple) INFO_REPLY packagescontaining basic device
	// information.
	commandIDInfoReply commandID = 6
	// commandIDSetCommand sets the device light intensities.
	commandIDSetCommand commandID = 7
	// commandIDSetCommand sets the device light intensities.
	// commandIDSendAddMasterToSlave is a broadcast sent from masters to announce
	// to all lamps. Sent on startup, when lamp is set as master, and periodically
	// every 90s.
	commandIDSendAddMasterToSlave commandID = 8
	// commandIDSendSetWavelengthsRelativePower is a light intensity message
	// broadcast from masters to all lamps.	Sent on startup, when lamp is set as
	// master and periodically every 60s.
	commandIDSendSetWavelengthsRelativePower commandID = 9
)

// DeviceInfo is the information about a device returned during a scan.
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

var broadcastIPV4 = net.IPv4(255, 255, 255, 255)

// ScanUDP performs a UDP device scan. The scan ends when the ctx is closed or
// after 4 seconds.
func ScanUDP(ctx context.Context) ([]DeviceInfo, error) {
	ctx, cancel := context.WithTimeout(ctx, 4*time.Second)
	defer cancel()

	socket, err := net.DialUDP("udp4", nil, &net.UDPAddr{
		IP:   broadcastIPV4,
		Port: UDPPort,
	})
	if err != nil {
		return nil, err
	}
	defer socket.Close()

	recvSocket, err := net.ListenUDP("udp4", &net.UDPAddr{
		IP:   net.IPv4(0, 0, 0, 0),
		Port: UDPPort,
	})
	if err != nil {
		return nil, err
	}
	defer recvSocket.Close()

	ch := make(chan DeviceInfo)
	go udpScanReceive(ctx, recvSocket, ch)

	payload, err := makeUDPPayloadShort(commandIDQuery)
	if err != nil {
		return nil, err
	}
	if _, err = socket.Write(payload); err != nil {
		return nil, err
	}

	resultSerials := make(map[string]bool)
	results := make([]DeviceInfo, 0, 64)
	for {
		select {
		case di := <-ch:
			if !resultSerials[di.SerialNum] {
				resultSerials[di.SerialNum] = true
				results = append(results, di)
			}
		case <-ctx.Done():
			return results, nil
		}
	}
}

func udpScanReceive(ctx context.Context, conn *net.UDPConn, ch chan<- DeviceInfo) {
	data := make([]byte, 4096)
	for {
		read, remoteAddr, err := conn.ReadFromUDP(data)
		if err != nil {
			return
		}
		if remoteAddr.Port != UDPPort {
			continue
		}
		if read < 17 {
			continue // invalid, scan results must be > 17 chars
		}
		cmdID := commandID(data[12])
		if cmdID != commandIDInfoReply {
			continue // we only care about scan responses
		}
		xmldata := data[16:read]
		di := DeviceInfo{}
		if err := xml.Unmarshal(xmldata, &di); err != nil {
			fmt.Printf("error unmarshaling scan response: %#v\n", err)
			continue
		}
		select {
		case ch <- di:
		case <-ctx.Done():
			return
		}
	}
}

// makeUDPPayloadShort makes a UDP command payload using default values.
func makeUDPPayloadShort(cmd commandID) ([]byte, error) {
	hwAddr, err := net.ParseMAC("FF:FF:FF:FF:FF:FF")
	if err != nil {
		return nil, err
	}
	return makeUDPPayload(cmd, hwAddr, nil)
}

// makeUDPPayload makes a UDP command payload.
func makeUDPPayload(cmd commandID, mac net.HardwareAddr, data []byte) ([]byte, error) {
	var buf bytes.Buffer
	if _, err := buf.Write([]byte("ABC321")); err != nil {
		return nil, err
	}
	if _, err := buf.Write(mac); err != nil {
		return nil, err
	}
	cmdHex, err := cmdAsHex(cmd)
	if err != nil {
		return nil, err
	}
	if _, err := buf.Write(cmdHex); err != nil {
		return nil, err
	}
	if err := buf.WriteByte(0x00); err != nil {
		return nil, err
	}

	if data != nil {
		if err := buf.WriteByte(uint8(len(data) % 256)); err != nil {
			return nil, err
		}
		if err := buf.WriteByte(uint8(len(data) / 256)); err != nil {
			return nil, err
		}
		if _, err := buf.Write(data); err != nil {
			return nil, err
		}
	} else {
		if _, err := buf.Write([]byte{0x00, 0x00}); err != nil {
			return nil, err
		}
	}

	return buf.Bytes(), nil
}

func cmdAsHex(cmd commandID) ([]byte, error) {
	dst := make([]byte, 1)
	_, err := hex.Decode(dst, []byte(fmt.Sprintf("%02d", cmd)))
	return dst, err
}
