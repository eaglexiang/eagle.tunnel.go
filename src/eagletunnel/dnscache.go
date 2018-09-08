package eagletunnel

import (
	"time"
)

type DnsCache struct {
	domain  string
	ip      string
	created time.Time
}

func CreateDnsCache(domain, ip string) *DnsCache {
	cache := DnsCache{domain: domain, ip: ip, created: time.Now()}
	return &cache
}

func (cache *DnsCache) Check() bool {
	duration := time.Since(cache.created)
	if duration > 2*time.Hour {
		return false
	}
	return true
}
