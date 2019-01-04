/*
 * @Description:
 * @Author: EagleXiang
 * @Github: https://github.com/eaglexiang
 * @Date: 2019-01-03 18:06:14
 * @LastEditors: EagleXiang
 * @LastEditTime: 2019-01-04 18:36:59
 */

package eagletunnel

import (
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	"../eaglelib/src"
)

// LoginStatus 用户的登录记录
type LoginStatus struct {
	log   *eaglelib.Cache
	count int
	Cap   int
	lock  sync.Mutex
}

// CreateLoginStatus 创建LoginStatus，ttl：登录记录的生存时间
func CreateLoginStatus(cap int) *LoginStatus {
	log := eaglelib.CreateCache(time.Minute * time.Duration(3)) // 3min
	ls := LoginStatus{log: log, Cap: cap}
	return &ls
}

// Login 登录
func (ls *LoginStatus) Login(ip string) error {
	if ls.Cap == 0 {
		// 不需要登记
		return nil
	}
	if ls.log.Exsit(ip) {
		// 已登录
		ls.log.Update(ip, "")
		return nil
	}
	ls.lock.Lock()
	defer ls.lock.Unlock()
	if ls.log.Exsit(ip) {
		// 已登录
		ls.log.Update(ip, "")
		return nil
	}
	err := ls.newLogin(ip)
	if err != nil {
		return errors.New("LoginStatus.Login -> " + err.Error())
	}
	return nil
}

func (ls *LoginStatus) newLogin(ip string) error {
	if ls.count == ls.Cap {
		errStr := "LoginStatus.newLogin -> too much login reqs " +
			ls.Count()
		return errors.New(errStr)
	}
	ls.count++
	ls.log.Add(ip)
	fmt.Println("New Login: " + ip)
	return nil
}

func parseLoginCount(arg string) (int, error) {
	switch arg {
	case "private", "PRIVATE":
		return PrivateUser, nil
	case "share", "shared", "SHARED":
		return SharedUser, nil
	default:
		value, err := strconv.ParseInt(arg, 10, 32)
		return int(value), err
	}
}

// Count 以比例的格式输出当前登录地占用状况，如 2/5
func (ls *LoginStatus) Count() string {
	value := strconv.FormatInt(int64(ls.count), 10) + "/" +
		strconv.FormatInt(int64(ls.Cap), 10)
	return value
}
