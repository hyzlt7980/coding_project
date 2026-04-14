package output

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"

	"mdnsmap/internal/banner"
	"mdnsmap/internal/model"
)

func SortServices(in []model.ServiceRecord) {
	sort.Slice(in, func(i, j int) bool {
		if in[i].Port != in[j].Port {
			return in[i].Port < in[j].Port
		}
		if in[i].ServiceType != in[j].ServiceType {
			return in[i].ServiceType < in[j].ServiceType
		}
		return in[i].InstanceName < in[j].InstanceName
	})
}

func WriteText(w io.Writer, res model.DiscoveryResult) error {
	services := append([]model.ServiceRecord{}, res.Services...)
	SortServices(services)
	if _, err := fmt.Fprintln(w, "services:"); err != nil {
		return err
	}
	for _, s := range services {
		header := fmt.Sprintf("%d/tcp %s:", s.Port, s.ServiceShort)
		if s.Port <= 0 {
			header = fmt.Sprintf("%s:", s.ServiceShort)
		}
		if _, err := fmt.Fprintln(w, header); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(w, "Name=%s\n", s.InstanceName); err != nil {
			return err
		}
		if len(s.IPv4) > 0 {
			fmt.Fprintf(w, "IPv4=%s\n", strings.Join(s.IPv4, ","))
		}
		if len(s.IPv6) > 0 {
			fmt.Fprintf(w, "IPv6=%s\n", strings.Join(s.IPv6, ","))
		}
		if s.HostName != "" {
			fmt.Fprintf(w, "Hostname=%s\n", s.HostName)
		}
		if s.TTL > 0 {
			fmt.Fprintf(w, "TTL=%d\n", s.TTL)
		}
		for _, line := range banner.TxtLines(s) {
			fmt.Fprintln(w, line)
		}
		for _, k := range sortedKeys(s.Banner) {
			fmt.Fprintf(w, "%s=%s\n", k, s.Banner[k])
		}
		fmt.Fprintln(w)
	}
	if _, err := fmt.Fprintln(w, "answers:"); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w, "PTR:"); err != nil {
		return err
	}
	ptr := append([]string{}, res.PTRAnswers...)
	sort.Strings(ptr)
	for _, p := range ptr {
		if _, err := fmt.Fprintln(w, p); err != nil {
			return err
		}
	}
	return nil
}

func WriteJSON(w io.Writer, res model.DiscoveryResult) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(res)
}

func sortedKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
