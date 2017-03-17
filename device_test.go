package heliospectra

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestNewDevice(t *testing.T) {
	client := &http.Client{}
	d := NewDevice(net.IPv4(1, 2, 3, 4), client)
	if !reflect.DeepEqual(d.addr, net.IPv4(1, 2, 3, 4)) {
		t.Errorf("expected IP 1.2.3.4, got %s", d.addr)
	}
	if !reflect.DeepEqual(d.client, client) {
		t.Errorf("expected passed-in client to be used")
	}
	d = NewDevice(net.IPv4(1, 2, 3, 4), nil)
	if !reflect.DeepEqual(d.client, http.DefaultClient) {
		t.Errorf("expected DefaultClient to be used")
	}
}

const diagResponse = `
<diagnostic>
	<model>L4</model>
	<cpuFW>R2.2.25</cpuFW>
	<driverFW>N/A</driverFW>
	<ethernetMAC>64:1a:00:00:00:00</ethernetMAC>
	<wlanMAC></wlanMAC>
	<wavelengths>0:450nm:10.2W,1:660nm:5.2W,2:735nm:10.0W,3:5700K:6.0W,</wavelengths>
	<clock>2017:03:17:02:48:41</clock>
	<onSchedule>Not running</onSchedule>
	<masterOrSlave>Independent</masterOrSlave>
	<systemStatus>OK</systemStatus>
	<runtime>0d 02h 10m 08s</runtime>
	<latestChange>2017-03-17	02:06:25</latestChange>
	<changedBy>Web</changedBy>
	<changeIP>192.168.1.3</changeIP>
	<changeType>Light setting</changeType>
	<temps>0:26.8C,</temps>
	<intensities>0:0,1:0,2:0,3:0,</intensities>
	<useNTP>1</useNTP>
	<networkType>dynamic</networkType>
	<networkIP>192.168.1.8</networkIP>
	<networkSubnet>255.255.255.0</networkSubnet>
	<networkGateway>192.168.1.1</networkGateway>
	<networkDNS1>192.168.1.1</networkDNS1>
	<networkDNS2>0.0.0.0</networkDNS2>
	<allowedTemp>15.0 60.0:59.0 140.0</allowedTemp>
	<hs>51</hs>
	<title>L4</title>
	<wlanIP></wlanIP>
	<ethernetIP>192.168.1.8</ethernetIP>
	<ntpOffset>00:00:00</ntpOffset>
	<masters> </masters>
	<dialog> </dialog>
	<poweredLink>http://www.heliospectra.com</poweredLink>
	<poweredText>Powered by Heliospectra</poweredText>
	<ntpPoolType>default</ntpPoolType>
	<ntpPoolCustom>pool.ntp.org</ntpPoolCustom>
	<favicon>/favi.ico</favicon>
	<tempUnit>C</tempUnit>
	<lockData>off:Enter your message here:heliospectra</lockData>
	<shortcuts> </shortcuts>
	<ntpData>on, pool.ntp.org, 00:00:00</ntpData>
	<multicastIP>239.153.155.131</multicastIP>
	<tags>0|^|name|^||~|</tags>
</diagnostic>`

func TestDevice_Diagnostic(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	statusToReturn := 200
	bodyToReturn := diagResponse

	diagHandler := func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/diag.xml" {
			t.Errorf("expected URL /diag.xml, got %s", r.URL.Path)
		}
		w.WriteHeader(statusToReturn)
		if _, err := w.Write([]byte(bodyToReturn)); err != nil {
			t.Fatal(err)
		}
	}
	server := httptest.NewServer(http.HandlerFunc(diagHandler))
	defer server.Close()

	testIP := net.IPv4(192, 168, 1, 8)

	client := &http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				expectedAddr := testIP.String() + ":80"
				if addr != expectedAddr {
					t.Errorf("expected request to be sent to %s, was sent to %s", expectedAddr, addr)
				}
				// override the DialContext func to only dial to our test server:
				return (&net.Dialer{
					Timeout:   1 * time.Second,
					KeepAlive: 1 * time.Second,
					DualStack: false,
				}).DialContext(ctx, network, strings.TrimPrefix(server.URL, "http://"))
			},
		},
	}

	device := NewDevice(testIP, client)
	diag, err := device.Diagnostic(ctx)
	if err != nil {
		t.Fatal(err)
	}

	if expModel := "L4"; expModel != diag.Model {
		t.Errorf("expected model=%q, got %q", expModel, diag.Model)
	}
	expWavelengths := WavelengthList{
		{Number: 0, Wavelength: "450nm", Power: "10.2W"},
		{Number: 1, Wavelength: "660nm", Power: "5.2W"},
		{Number: 2, Wavelength: "735nm", Power: "10.0W"},
		{Number: 3, Wavelength: "5700K", Power: "6.0W"},
	}
	if !reflect.DeepEqual(expWavelengths, diag.Wavelengths) {
		t.Errorf("expected wavelengths=%#v\n\tgot %#v", expWavelengths, diag.Wavelengths)
		t.Logf("DIAG: %+v\n", diag)
	}

	statusToReturn = 400
	_, err = device.Diagnostic(ctx)
	if err == nil {
		t.Errorf("expected an error on status 400, got none")
	}

	statusToReturn = 200
	bodyToReturn = `{"some":"json"}`
	if _, err = device.Diagnostic(ctx); err == nil {
		t.Errorf("expected an error on a non-XML body, got none")
	}
}

func TestDevice_SetIntensities(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	statusToReturn := 200
	expectedQuery := url.Values{"int": []string{"1:2:3:4"}}

	diagHandler := func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if err := r.ParseForm(); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(expectedQuery, r.Form) {
			t.Errorf("expected query %#v, got %#v", expectedQuery, r.Form)
		}
		if r.URL.Path != "/intensity.cgi" {
			t.Errorf("expected URL /intensity.cgi, got %s", r.URL.Path)
		}
		w.WriteHeader(statusToReturn)
	}
	server := httptest.NewServer(http.HandlerFunc(diagHandler))
	defer server.Close()

	testIP := net.IPv4(192, 168, 1, 8)

	client := &http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				expectedAddr := testIP.String() + ":80"
				if addr != expectedAddr {
					t.Errorf("expected request to be sent to %s, was sent to %s", expectedAddr, addr)
				}
				// override the DialContext func to only dial to our test server:
				return (&net.Dialer{
					Timeout:   1 * time.Second,
					KeepAlive: 1 * time.Second,
					DualStack: false,
				}).DialContext(ctx, network, strings.TrimPrefix(server.URL, "http://"))
			},
		},
	}

	device := NewDevice(testIP, client)
	if err := device.SetIntensities(ctx, 1, 2, 3, 4); err != nil {
		t.Fatal(err)
	}

	statusToReturn = 400
	if err := device.SetIntensities(ctx, 1, 2, 3, 4); err == nil {
		t.Errorf("expected an error on status 400, got none")
	}
}

const statusResponse = `<r>
<a>2017:03:17:19:07:56</a>
<b>Not running</b>
<c>OK</c>
<d>0d 02h 39m 37s</d>
<e>2017-03-17	18:58:34</e>
<f>Web</f>
<g>192.168.1.3</g>
<h>Light setting</h>
<i>0:26.0C,</i>
<j>0:0,1:0,2:0,3:0,</j>
<k> </k>
<l> </l>
<m>Independent</m>
<n>C:on</n>
<o>off:Enter your message here:heliospectra</o>
<p> </p>
<q>on, pool.ntp.org, 00:00:00</q>
<s>on</s>
<r></r>
<t>0.0A,0.0W</t>
</r>`

func TestDevice_Status(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	statusToReturn := 200
	bodyToReturn := statusResponse

	statusHandler := func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/status.xml" {
			t.Errorf("expected URL /status.xml, got %s", r.URL.Path)
		}
		w.WriteHeader(statusToReturn)
		if _, err := w.Write([]byte(bodyToReturn)); err != nil {
			t.Fatal(err)
		}
	}
	server := httptest.NewServer(http.HandlerFunc(statusHandler))
	defer server.Close()

	testIP := net.IPv4(192, 168, 1, 8)

	client := &http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				expectedAddr := testIP.String() + ":80"
				if addr != expectedAddr {
					t.Errorf("expected request to be sent to %s, was sent to %s", expectedAddr, addr)
				}
				// override the DialContext func to only dial to our test server:
				return (&net.Dialer{
					Timeout:   1 * time.Second,
					KeepAlive: 1 * time.Second,
					DualStack: false,
				}).DialContext(ctx, network, strings.TrimPrefix(server.URL, "http://"))
			},
		},
	}

	device := NewDevice(testIP, client)
	status, err := device.Status(ctx)
	if err != nil {
		t.Fatal(err)
	}

	expected := &Status{
		InternalTime: "2017:03:17:19:07:56",
		OnSchedule:   "Not running",
		Status:       "OK",
		Uptime:       "0d 02h 39m 37s",
		LastChangeAt: "2017-03-17	18:58:34",
		LastChangeInterface: "Web",
		LastChangeBy:        net.IPv4(192, 168, 1, 3),
		LastChangeType:      "Light setting",
		Temp:                "0:26.0C,",
		Intensities:         "0:0,1:0,2:0,3:0,",
		Masters:             " ",
		Reserved:            " ",
		ControlMode:         "Independent",
		NTPTimeSettings:     "on, pool.ntp.org, 00:00:00",
	}

	if !reflect.DeepEqual(expected, status) {
		t.Errorf("expected status=%#v\n\ngot status=%#v", expected, status)
	}

	statusToReturn = 400
	_, err = device.Status(ctx)
	if err == nil {
		t.Errorf("expected an error on status 400, got none")
	}

	statusToReturn = 200
	bodyToReturn = `{"some":"json"}`
	if _, err = device.Status(ctx); err == nil {
		t.Errorf("expected an error on a non-XML body, got none")
	}
}
