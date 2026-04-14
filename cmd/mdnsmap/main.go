package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"mdnsmap/internal/banner"
	"mdnsmap/internal/filter"
	"mdnsmap/internal/mdns"
	"mdnsmap/internal/model"
	"mdnsmap/internal/output"
	"mdnsmap/internal/ports"
)

type listFlag []string

func (l *listFlag) String() string { return strings.Join(*l, ",") }
func (l *listFlag) Set(v string) error {
	*l = append(*l, v)
	return nil
}

func main() {
	var cidrs listFlag
	var portsSpec, iface, outPath string
	var wait, timeout time.Duration
	var workers int
	var asJSON, verbose bool

	flag.Var(&cidrs, "cidr", "repeatable and/or comma-separated CIDR filter")
	flag.StringVar(&portsSpec, "ports", "1-65535", "single ports/ranges: 80,445,5000-5005")
	flag.StringVar(&iface, "iface", "", "optional network interface name")
	flag.DurationVar(&wait, "wait", 3*time.Second, "mDNS collection window")
	flag.DurationVar(&timeout, "timeout", 1500*time.Millisecond, "HTTP probe timeout")
	flag.IntVar(&workers, "workers", 4, "worker count for active HTTP enrichment")
	flag.BoolVar(&asJSON, "json", false, "emit JSON output")
	flag.StringVar(&outPath, "output", "", "write output file path")
	flag.BoolVar(&verbose, "verbose", false, "verbose logging")
	flag.Parse()

	matcher, err := filter.NewCIDRMatcher(cidrs)
	if err != nil {
		exitErr(err)
	}
	ps, err := ports.Parse(portsSpec)
	if err != nil {
		exitErr(err)
	}

	ctx := context.Background()
	result, err := mdns.Discover(ctx, mdns.Options{IfaceName: iface, Wait: wait, Verbose: verbose})
	if err != nil {
		exitErr(err)
	}

	result.Services = filter.Apply(result.Services, matcher, ps)
	enrich(result.Services, timeout, workers)

	var out *os.File
	if outPath != "" {
		out, err = os.Create(outPath)
		if err != nil {
			exitErr(err)
		}
		defer out.Close()
	} else {
		out = os.Stdout
	}
	if asJSON {
		err = output.WriteJSON(out, result)
	} else {
		err = output.WriteText(out, result)
	}
	if err != nil {
		exitErr(err)
	}
}

func enrich(services []model.ServiceRecord, timeout time.Duration, workers int) {
	if workers <= 0 {
		workers = 1
	}
	ch := make(chan *model.ServiceRecord)
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for s := range ch {
				if s.ServiceShort == "http" || s.ServiceShort == "https" {
					ctx, cancel := context.WithTimeout(context.Background(), timeout)
					banner.EnrichHTTP(ctx, s, timeout)
					cancel()
				}
			}
		}()
	}
	for i := range services {
		ch <- &services[i]
	}
	close(ch)
	wg.Wait()
}

func exitErr(err error) {
	fmt.Fprintln(os.Stderr, "error:", err)
	os.Exit(1)
}
