package eagletunnel

import (
	"time"
)

type DNSCache struct {
	domain  string
	ip      string
	created time.Time
}

func CreateDNSCache(domain, ip string) *DNSCache {
	cache := DNSCache{domain: domain, ip: ip, created: time.Now()}
	return &cache
}

func (cache *DNSCache) Check() bool {
	duration := time.Since(cache.created)
	if duration > 2*time.Hour {
		return false
	}
	return true
}
