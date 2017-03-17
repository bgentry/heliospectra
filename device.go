package heliospectra

import (
	"context"
	"encoding/xml"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strconv"
)

// Device is a Heliospectra LED device.
type Device struct {
	addr   net.IP
	client *http.Client
}

// NewDevice creates a new device from an IP address. If client is nil, the
// http.DefaultClient is used.
func NewDevice(addr net.IP, client *http.Client) *Device {
	if client == nil {
		client = http.DefaultClient
	}
	return &Device{addr: addr, client: client}
}

// Diagnostic executes a diagnostic request against the Device.
func (d *Device) Diagnostic(ctx context.Context) (*Diagnostic, error) {
	u := url.URL{
		Host:   d.addr.String() + ":" + strconv.Itoa(TCPPort),
		Scheme: "http",
		Path:   "diag.xml",
	}
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}
	req.WithContext(ctx)
	res, err := d.client.Do(req)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("unexpected status code %d", res.StatusCode)
	}
	diag := &Diagnostic{}
	if err = xml.NewDecoder(res.Body).Decode(diag); err != nil {
		return nil, err
	}
	return diag, nil
}

// Diagnostic is the result of a diagnostic request against a Device.
type Diagnostic struct {
	Model          string `xml:"model"`
	CPUFW          string `xml:"cpuFW"`
	DriverFW       string `xml:"driverFW"`
	EthernetMAC    string `xml:"ethernetMAC"`
	WlanMAC        string `xml:"wlanMAC"`
	Wavelengths    string `xml:"wavelengths"`
	Clock          string `xml:"clock"`
	OnSchedule     string `xml:"onSchedule"`
	MasterOrSlave  string `xml:"masterOrSlave"`
	SystemStatus   string `xml:"systemStatus"`
	Runtime        string `xml:"runtime"`
	LatestChange   string `xml:"latestChange"`
	ChangedBy      string `xml:"changedBy"`
	ChangeIP       string `xml:"changeIP"`
	ChangeType     string `xml:"changeType"`
	Temps          string `xml:"temps"`
	Intensities    string `xml:"intensities"`
	UseNTP         uint   `xml:"useNTP"`
	NetworkType    string `xml:"networkType"`
	NetworkIP      net.IP `xml:"networkIP"`
	NetworkSubnet  net.IP `xml:"networkSubnet"`
	NetworkGateway net.IP `xml:"networkGateway"`
	NetworkDNS1    net.IP `xml:"networkDNS1"`
	NetworkDNS2    net.IP `xml:"networkDNS2"`
	AllowedTemp    string `xml:"allowedTemp"`
	Hs             string `xml:"hs"`
	Title          string `xml:"title"`
	WLANIP         net.IP `xml:"wlanIP"`
	EthernetIP     net.IP `xml:"ethernetIP"`
	NTPOffset      string `xml:"ntpOffset"`
	Masters        string `xml:"masters"`
	Dialog         string `xml:"dialog"`
	PoweredLink    string `xml:"poweredLink"`
	PoweredText    string `xml:"poweredText"`
	NTPPoolType    string `xml:"ntpPoolType"`
	NTPPoolCustom  string `xml:"ntpPoolCustom"`
	Favicon        string `xml:"favicon"`
	TempUnit       string `xml:"tempUnit"`
	LockData       string `xml:"lockData"`
	Shortcuts      string `xml:"shortcuts"`
	NTPData        string `xml:"ntpData"`
	MulticastIP    string `xml:"multicastIP"`
	Tags           string `xml:"tags"`
}
