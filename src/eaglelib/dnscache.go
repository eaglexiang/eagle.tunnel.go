package eaglelib

import (
	"time"
)

// DNSCache DNS缓存，包含域名，以及对应的IP，及更新时间
type DNSCache struct {
	Domain  string
	IP      string
	Updated time.Time
}

// CreateDNSCache 创建DNSCache的方法
func CreateDNSCache(domain, ip string) *DNSCache {
	cache := DNSCache{Domain: domain, IP: ip, Updated: time.Now()}
	return &cache
}

// OverTTL 检查DNS是否过期
func (cache *DNSCache) OverTTL() bool {
	duration := time.Since(cache.Updated)
	if duration > 2*time.Hour {
		return false
	}
	return true
}
