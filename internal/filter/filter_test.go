package filter

import (
	"testing"

	"mdnsmap/internal/model"
	"mdnsmap/internal/ports"
)

func TestCIDRMatcherAndApply(t *testing.T) {
	m, err := NewCIDRMatcher([]string{"192.168.1.0/24,fe80::/64"})
	if err != nil {
		t.Fatal(err)
	}
	ps, _ := ports.Parse("80,5000-5001")
	in := []model.ServiceRecord{
		{InstanceName: "a", Port: 80, IPv4: []string{"192.168.1.22"}},
		{InstanceName: "b", Port: 445, IPv4: []string{"192.168.1.23"}},
		{InstanceName: "c", Port: 0, IPv6: []string{"fe80::1"}},
		{InstanceName: "d", Port: 80, IPv4: []string{"10.0.0.1"}},
	}
	got := Apply(in, m, ps)
	if len(got) != 2 {
		t.Fatalf("expected 2, got %d", len(got))
	}
}
