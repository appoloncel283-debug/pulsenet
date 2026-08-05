package core

import "testing"

func TestParseWindowsGateway(t *testing.T) {
	output := `IPv4 Route Table
===========================================================================
Active Routes:
Network Destination        Netmask          Gateway       Interface  Metric
          0.0.0.0          0.0.0.0      192.168.1.1    192.168.1.20     25
`
	if got := parseWindowsGateway(output); got != "192.168.1.1" {
		t.Fatalf("gateway = %q", got)
	}
}

func TestParseIPRouteGateway(t *testing.T) {
	output := "default via 10.0.0.1 dev wlan0 proto dhcp src 10.0.0.5 metric 600\n"
	if got := parseIPRouteGateway(output); got != "10.0.0.1" {
		t.Fatalf("gateway = %q", got)
	}
}

func TestParseUnixRouteGateway(t *testing.T) {
	output := "   route to: default\ndestination: default\n       mask: default\n    gateway: 192.168.50.1\n"
	if got := parseUnixRouteGateway(output); got != "192.168.50.1" {
		t.Fatalf("gateway = %q", got)
	}
}
