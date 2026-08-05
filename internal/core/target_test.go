package core

import "testing"

func TestParseTargetURL(t *testing.T) {
	target, err := ParseTarget("https://example.com:8443/status")
	if err != nil {
		t.Fatal(err)
	}
	if target.Host != "example.com" || target.ExplicitPort != "8443" || !target.ExplicitURL {
		t.Fatalf("unexpected target: %#v", target)
	}
}

func TestParseTargetIPv6(t *testing.T) {
	target, err := ParseTarget("2001:db8::1")
	if err != nil {
		t.Fatal(err)
	}
	if target.Host != "2001:db8::1" {
		t.Fatalf("host = %q", target.Host)
	}
}

func TestParseTargetRejectsPathWithoutScheme(t *testing.T) {
	if _, err := ParseTarget("example.com/path"); err == nil {
		t.Fatal("expected an error")
	}
}
