package mdns

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"time"

	"mdnsmap/internal/banner"
	"mdnsmap/internal/model"
)

type Options struct {
	IfaceName string
	Wait      time.Duration
	Verbose   bool
}

type commandRunner interface {
	Run(context.Context, string, ...string) (string, error)
}

type execRunner struct{}

func (e execRunner) Run(ctx context.Context, name string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func Discover(ctx context.Context, opts Options) (model.DiscoveryResult, error) {
	if opts.Wait <= 0 {
		opts.Wait = 3 * time.Second
	}
	return discoverWithRunner(ctx, opts, execRunner{})
}

func discoverWithRunner(ctx context.Context, opts Options, r commandRunner) (model.DiscoveryResult, error) {
	browseCtx, cancel := context.WithTimeout(ctx, opts.Wait)
	defer cancel()
	args := []string{"--all", "--resolve", "--parsable", "--terminate"}
	if opts.IfaceName != "" {
		args = append(args, "--interface", opts.IfaceName)
	}
	out, err := r.Run(browseCtx, "avahi-browse", args...)
	if err != nil {
		return model.DiscoveryResult{}, fmt.Errorf("avahi-browse failed: %w", err)
	}
	res := parseAvahiOutput(out)
	if len(res.PTRAnswers) == 0 && len(res.Services) == 0 {
		return model.DiscoveryResult{}, errors.New("no mDNS/DNS-SD records discovered")
	}
	res.Wait = opts.Wait
	return res, nil
}

func parseAvahiOutput(raw string) model.DiscoveryResult {
	ptrSet := map[string]struct{}{}
	services := make([]model.ServiceRecord, 0)
	s := bufio.NewScanner(strings.NewReader(raw))
	for s.Scan() {
		line := s.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}
		parts := strings.Split(line, ";")
		if len(parts) < 10 {
			continue
		}
		if parts[0] != "=" && parts[0] != "+" {
			continue
		}
		serviceType := normalizeServiceType(parts[4])
		ptrSet[serviceType] = struct{}{}
		port, _ := strconv.Atoi(parts[8])
		txtRaw := parseAvahiTXT(parts[9])
		txt, order := banner.ParseTXT(txtRaw)
		rec := model.ServiceRecord{
			InstanceName: unescape(parts[3]),
			ServiceType:  serviceType,
			ServiceShort: shortName(serviceType),
			HostName:     strings.TrimSuffix(parts[6], "."),
			Port:         port,
			TXT:          txt,
			TXTOrder:     order,
			RawTXT:       txtRaw,
		}
		ip := strings.TrimSpace(parts[7])
		if strings.Contains(ip, ":") {
			rec.IPv6 = []string{ip}
		} else if ip != "" {
			rec.IPv4 = []string{ip}
		}
		services = append(services, rec)
	}
	ptr := make([]string, 0, len(ptrSet))
	for p := range ptrSet {
		ptr = append(ptr, p)
	}
	sort.Strings(ptr)
	return model.DiscoveryResult{Services: dedupe(services), PTRAnswers: ptr}
}

func parseAvahiTXT(raw string) []string {
	raw = strings.TrimSpace(raw)
	raw = strings.Trim(raw, "\"")
	if raw == "" {
		return nil
	}
	items := strings.Split(raw, "\" \"")
	out := make([]string, 0, len(items))
	for _, item := range items {
		out = append(out, unescape(strings.Trim(item, "\"")))
	}
	return out
}

func unescape(s string) string {
	s = strings.ReplaceAll(s, "\\\\", "\\")
	s = strings.ReplaceAll(s, "\\032", " ")
	return s
}

func dedupe(in []model.ServiceRecord) []model.ServiceRecord {
	seen := map[string]struct{}{}
	out := make([]model.ServiceRecord, 0, len(in))
	for _, r := range in {
		k := fmt.Sprintf("%s|%s|%s|%d|%v|%v", r.InstanceName, r.ServiceType, r.HostName, r.Port, r.IPv4, r.IPv6)
		if _, ok := seen[k]; ok {
			continue
		}
		seen[k] = struct{}{}
		out = append(out, r)
	}
	return out
}

func normalizeServiceType(s string) string {
	s = strings.TrimSpace(strings.TrimSuffix(s, "."))
	if s == "" {
		return s
	}
	if strings.HasSuffix(s, ".local") {
		return s
	}
	return s + ".local"
}

func shortName(t string) string {
	t = strings.TrimSuffix(t, ".local")
	parts := strings.Split(t, ".")
	if len(parts) == 0 {
		return t
	}
	return strings.TrimPrefix(parts[0], "_")
}
