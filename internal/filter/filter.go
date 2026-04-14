package filter

import (
	"fmt"
	"net"
	"strings"

	"mdnsmap/internal/model"
	"mdnsmap/internal/ports"
)

type CIDRMatcher struct {
	nets []*net.IPNet
}

func NewCIDRMatcher(inputs []string) (CIDRMatcher, error) {
	if len(inputs) == 0 {
		return CIDRMatcher{}, fmt.Errorf("at least one --cidr is required")
	}
	var nets []*net.IPNet
	for _, in := range inputs {
		for _, piece := range strings.Split(in, ",") {
			piece = strings.TrimSpace(piece)
			if piece == "" {
				continue
			}
			_, ipnet, err := net.ParseCIDR(piece)
			if err != nil {
				return CIDRMatcher{}, fmt.Errorf("invalid cidr %q: %w", piece, err)
			}
			nets = append(nets, ipnet)
		}
	}
	if len(nets) == 0 {
		return CIDRMatcher{}, fmt.Errorf("at least one --cidr is required")
	}
	return CIDRMatcher{nets: nets}, nil
}

func (m CIDRMatcher) ContainsAny(ips []string) bool {
	for _, raw := range ips {
		ip := net.ParseIP(raw)
		if ip == nil {
			continue
		}
		for _, n := range m.nets {
			if n.Contains(ip) {
				return true
			}
		}
	}
	return false
}

func Apply(records []model.ServiceRecord, c CIDRMatcher, p ports.Set) []model.ServiceRecord {
	out := make([]model.ServiceRecord, 0, len(records))
	for _, r := range records {
		ips := append(append([]string{}, r.IPv4...), r.IPv6...)
		if !c.ContainsAny(ips) {
			continue
		}
		if r.Port <= 0 {
			out = append(out, r)
			continue
		}
		if p.Contains(r.Port) {
			out = append(out, r)
		}
	}
	return out
}
