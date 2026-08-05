package core

import (
	"context"
	"fmt"
	"net"
	"sort"
	"strings"
	"sync"
	"time"
)

type resolverSpec struct {
	name string
	addr string
}

func ProbeDNS(ctx context.Context, host string, timeout time.Duration) []DNSResult {
	if ip := net.ParseIP(host); ip != nil {
		return []DNSResult{{Resolver: "IP address", Addresses: []string{ip.String()}}}
	}
	specs := []resolverSpec{
		{name: "System resolver"},
		{name: "Cloudflare 1.1.1.1", addr: "1.1.1.1:53"},
		{name: "Google 8.8.8.8", addr: "8.8.8.8:53"},
	}
	results := make([]DNSResult, len(specs))
	var wg sync.WaitGroup
	for i, spec := range specs {
		wg.Add(1)
		go func(index int, resolver resolverSpec) {
			defer wg.Done()
			results[index] = lookupWithResolver(ctx, host, timeout, resolver)
		}(i, spec)
	}
	wg.Wait()
	return results
}

func lookupWithResolver(parent context.Context, host string, timeout time.Duration, spec resolverSpec) DNSResult {
	started := time.Now()
	ctx, cancel := context.WithTimeout(parent, timeout)
	defer cancel()
	resolver := net.DefaultResolver
	if spec.addr != "" {
		dialer := &net.Dialer{Timeout: timeout}
		resolver = &net.Resolver{PreferGo: true, Dial: func(ctx context.Context, network, _ string) (net.Conn, error) {
			return dialer.DialContext(ctx, network, spec.addr)
		}}
	}
	addresses, err := resolver.LookupHost(ctx, host)
	result := DNSResult{Resolver: spec.name, DurationMS: time.Since(started).Milliseconds()}
	if err != nil {
		result.Error = err.Error()
		return result
	}
	result.Addresses = uniqueSorted(addresses)
	return result
}

func LookupDNSRecords(parent context.Context, input string, timeout time.Duration) DNSRecords {
	started := time.Now()
	records := DNSRecords{Host: input}
	ctx, cancel := context.WithTimeout(parent, timeout)
	defer cancel()

	if ip := net.ParseIP(input); ip != nil {
		ptr, err := net.DefaultResolver.LookupAddr(ctx, input)
		if err != nil {
			records.Errors = append(records.Errors, "PTR: "+err.Error())
		} else {
			records.PTR = uniqueSorted(ptr)
		}
		records.DurationMS = time.Since(started).Milliseconds()
		return records
	}

	ips, err := net.DefaultResolver.LookupIP(ctx, "ip", input)
	if err != nil {
		records.Errors = append(records.Errors, "A/AAAA: "+err.Error())
	} else {
		for _, ip := range ips {
			if ip.To4() != nil {
				records.A = append(records.A, ip.String())
			} else {
				records.AAAA = append(records.AAAA, ip.String())
			}
		}
		records.A = uniqueSorted(records.A)
		records.AAAA = uniqueSorted(records.AAAA)
	}
	if cname, err := net.DefaultResolver.LookupCNAME(ctx, input); err == nil {
		cname = strings.TrimSuffix(cname, ".")
		if !strings.EqualFold(cname, strings.TrimSuffix(input, ".")) {
			records.CNAME = cname
		}
	}
	if values, err := net.DefaultResolver.LookupMX(ctx, input); err == nil {
		for _, value := range values {
			records.MX = append(records.MX, fmt.Sprintf("%d %s", value.Pref, strings.TrimSuffix(value.Host, ".")))
		}
		sort.Strings(records.MX)
	}
	if values, err := net.DefaultResolver.LookupNS(ctx, input); err == nil {
		for _, value := range values {
			records.NS = append(records.NS, strings.TrimSuffix(value.Host, "."))
		}
		records.NS = uniqueSorted(records.NS)
	}
	if values, err := net.DefaultResolver.LookupTXT(ctx, input); err == nil {
		records.TXT = uniqueSorted(values)
	}
	records.DurationMS = time.Since(started).Milliseconds()
	return records
}

func uniqueSorted(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	out := make([]string, 0, len(values))
	for _, value := range values {
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	sort.Strings(out)
	return out
}
