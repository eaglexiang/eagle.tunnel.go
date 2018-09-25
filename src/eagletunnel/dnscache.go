package eagletunnel

import (
	"time"
)

// DNSCache DNS缓存，包含域名，以及对应的IP，及更新时间
type DNSCache struct {
	domain  string
	ip      string
	updated time.Time
}

// CreateDNSCache 创建DNSCache的方法
func CreateDNSCache(domain, ip string) *DNSCache {
	cache := DNSCache{domain: domain, ip: ip, updated: time.Now()}
	return &cache
}

// OverTTL 检查DNS是否过期
func (cache *DNSCache) OverTTL() bool {
	duration := time.Since(cache.updated)
	if duration > 2*time.Hour {
		return false
	}
	return true
}
