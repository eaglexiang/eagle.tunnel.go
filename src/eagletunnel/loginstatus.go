/*
 * @Description:
 * @Author: EagleXiang
 * @Github: https://github.com/eaglexiang
 * @Date: 2019-01-03 18:06:14
 * @LastEditors: EagleXiang
 * @LastEditTime: 2019-01-03 18:06:48
 */

package eagletunnel

import (
	"time"
)

// LoginStatus 用户的登录记录
type LoginStatus struct {
	ip       string
	lastTime time.Time
	ttl      int // min
}

func (ls *LoginStatus) isDead() bool {
	if ls.ttl == 0 {
		return false
	}
	now := time.Now()
	duration := now.Sub(ls.lastTime)
	if duration > time.Duration(ls.ttl)*time.Minute {
		return true
	}
	ls.lastTime = now
	return false
}
