package core

import (
	"reflect"
	"testing"
)

func TestParsePorts(t *testing.T) {
	ports, err := ParsePorts("443,80,8000-8002,443", 20)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"80", "443", "8000", "8001", "8002"}
	if !reflect.DeepEqual(ports, want) {
		t.Fatalf("ports = %#v, want %#v", ports, want)
	}
}

func TestParsePortsLimit(t *testing.T) {
	if _, err := ParsePorts("1-200", 128); err == nil {
		t.Fatal("expected a limit error")
	}
}
