package banner

import (
	"context"
	"crypto/tls"
	"io"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"mdnsmap/internal/model"
)

var titleRe = regexp.MustCompile(`(?is)<title[^>]*>(.*?)</title>`)

func ParseTXT(txt []string) (map[string]string, []string) {
	m := make(map[string]string, len(txt))
	order := make([]string, 0, len(txt))
	for _, item := range txt {
		if item == "" {
			continue
		}
		parts := strings.SplitN(item, "=", 2)
		key := strings.TrimSpace(parts[0])
		val := ""
		if len(parts) == 2 {
			val = strings.TrimSpace(parts[1])
		}
		if key == "" {
			continue
		}
		if _, ok := m[key]; !ok {
			order = append(order, key)
		}
		m[key] = val
	}
	return m, order
}

func TxtLines(rec model.ServiceRecord) []string {
	if len(rec.TXT) == 0 {
		return nil
	}
	keys := append([]string{}, rec.TXTOrder...)
	if len(keys) == 0 {
		for k := range rec.TXT {
			keys = append(keys, k)
		}
		sort.Strings(keys)
	}
	pairs := make([]string, 0, len(keys))
	for _, k := range keys {
		pairs = append(pairs, k+"="+rec.TXT[k])
	}
	return []string{strings.Join(pairs, ",")}
}

func EnrichHTTP(ctx context.Context, rec *model.ServiceRecord, timeout time.Duration) {
	if rec.Port <= 0 || len(rec.IPv4) == 0 && len(rec.IPv6) == 0 {
		return
	}
	if rec.Banner == nil {
		rec.Banner = map[string]string{}
	}
	ip := ""
	if len(rec.IPv4) > 0 {
		ip = rec.IPv4[0]
	} else {
		ip = "[" + rec.IPv6[0] + "]"
	}
	scheme := "http"
	if rec.ServiceShort == "https" {
		scheme = "https"
	}
	url := scheme + "://" + ip + ":" + strconv.Itoa(rec.Port) + "/"
	client := &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //nolint:gosec
		},
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	rec.Banner["path"] = "/"
	rec.Banner["status"] = resp.Status
	if v := resp.Header.Get("Server"); v != "" {
		rec.Banner["server"] = v
	}
	if v := resp.Header.Get("Content-Type"); v != "" {
		rec.Banner["contentType"] = v
	}
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 8192))
	if m := titleRe.FindStringSubmatch(string(body)); len(m) == 2 {
		title := strings.TrimSpace(m[1])
		if title != "" {
			rec.Banner["title"] = title
		}
	}
}
