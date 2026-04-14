package mdns

import (
	"context"
	"errors"
	"testing"
	"time"
)

type fakeRunner struct {
	out string
	err error
}

func (f fakeRunner) Run(context.Context, string, ...string) (string, error) { return f.out, f.err }

func TestParseAvahiOutput(t *testing.T) {
	raw := `=;eth0;IPv4;slw-nas;_qdiscover._tcp;local;slw-nas.local;192.168.1.2;5000;"accessType=https" "model=TS-464C"
=;eth0;IPv4;slw-nas(AFP);_device-info._tcp;local;slw-nas.local;192.168.1.2;0;"model=Xserve"`
	res := parseAvahiOutput(raw)
	if len(res.Services) != 2 {
		t.Fatalf("expected 2 services, got %d", len(res.Services))
	}
	if res.Services[0].TXT["accessType"] != "https" {
		t.Fatalf("unexpected txt: %#v", res.Services[0].TXT)
	}
	if len(res.PTRAnswers) != 2 {
		t.Fatalf("expected ptr answers")
	}
}

func TestDiscoverWithRunnerError(t *testing.T) {
	_, err := discoverWithRunner(context.Background(), Options{Wait: time.Second}, fakeRunner{err: errors.New("boom")})
	if err == nil {
		t.Fatal("expected error")
	}
}
