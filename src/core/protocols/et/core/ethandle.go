/*
 * @Author: EagleXiang
 * @LastEditors: EagleXiang
 * @Email: eagle.xiang@outlook.com
 * @Github: https://github.com/eaglexiang
 * @Date: 2019-03-19 20:08:35
 * @LastEditTime: 2019-06-14 22:38:37
 */

package et

import (
	"errors"
	"strings"

	"github.com/eaglexiang/eagle.tunnel.go/src/core/protocols/et/comm"
	mycipher "github.com/eaglexiang/go-cipher"
	"github.com/eaglexiang/go-logger"
	mynet "github.com/eaglexiang/go-net"
	"github.com/eaglexiang/go-tunnel"
	myuser "github.com/eaglexiang/go-user"
	"go.uber.org/ratelimit"
)

// Handle 处理ET请求
func (et *ET) Handle(e *mynet.Arg) (err error) {
	tunnel := e.Tunnel
	err = et.checkHeaderOfReq(string(e.Msg), tunnel)
	if err != nil {
		return err
	}
	createCipher(tunnel)
	err = et.checkUserOfReq(tunnel)
	if err != nil {
		return err
	}
	// 选择子协议handler
	subReq, h, err := et.subShake(tunnel)
	if err != nil {
		return err
	}
	// 进入子协议业务
	err = h.Handle(subReq, tunnel)
	if err != nil {
		return err
	}
	// 只有TCP子协议需要继续运行
	if h.Type() != comm.TCP {
		return errors.New("no need to continue")
	}
	return nil
}

func (et *ET) subShake(tunnel *tunnel.Tunnel) (subReq string,
	h comm.Handler, err error) {
	subReq, err = tunnel.ReadLeftStr()
	if err != nil {
		return "", nil, err
	}
	h, err = comm.GetHandler(strings.Split(subReq, " ")[0])
	if err != nil {
		logger.Warning(err)
	}
	return subReq, h, err
}

func createCipher(t *tunnel.Tunnel) {
	c := mycipher.DefaultCipher()
	if c == nil {
		panic("ET.Handle -> cipher is nil")
	}
	t.Update(tunnel.WithLeftCipher(c))
}

func (et *ET) checkHeaderOfReq(
	header string,
	tunnel *tunnel.Tunnel) error {
	headers := strings.Split(header, " ")
	switch {
	case len(headers) < 1:
		return errors.New("checkHeaderOfReq -> nil req")
	case headers[0] != comm.ETArg.Head:
		logger.Warning("invalid header of req: ", headers[0])
		return errors.New("checkHeaderOfReq -> wrong head")
	default:
		reply := "valid valid valid"
		_, err := tunnel.WriteLeft([]byte(reply))
		return err
	}
}

func (et *ET) checkUserOfReq(t *tunnel.Tunnel) (err error) {
	if comm.ETArg.ValidUsers == nil {
		// 未启用用户校验
		return
	}
	var user2Check *myuser.ReqUser
	if user2Check, err = findReqUser(t); err != nil {
		logger.Warning(err)
		return
	}
	if sl, err := et._checkUserOfReq(user2Check); err == nil {
		t.Update(tunnel.WithSpeedLimiter(sl))
		_, err = t.WriteLeft([]byte("valid"))
	}
	return
}

func findReqUser(t *tunnel.Tunnel) (*myuser.ReqUser, error) {
	userStr, err := t.ReadLeftStr()
	if err != nil {
		return nil, err
	}
	user2Check, err := parseReqUser(
		userStr,
		mynet.GetIPOfConnRemote(t.Left()))
	return user2Check, err
}

func (et *ET) _checkUserOfReq(user2Check *myuser.ReqUser) (limiter ratelimit.Limiter, err error) {
	validUser, ok := comm.ETArg.ValidUsers[user2Check.ID]
	if !ok {
		logger.Warning("user not found: ", user2Check.ID)
		return nil, errors.New("user not found")
	}
	if err = validUser.CheckAuth(user2Check); err == nil {
		limiter = validUser.SpeedLimiter()
	} else {
		logger.Warning(err)
	}
	return
}

func parseReqUser(userStr, ip string) (*myuser.ReqUser, error) {
	user2Check, err := myuser.ParseReqUser(userStr, ip)
	if err != nil {
		logger.Warning("invalid user text: ", userStr)
		return nil, err
	}
	if user2Check.ID == "null" {
		return nil, errors.New("invalid null req user")
	}
	return user2Check, nil
}
