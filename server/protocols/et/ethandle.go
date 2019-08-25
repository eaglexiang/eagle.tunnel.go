/*
 * @Author: EagleXiang
 * @LastEditors: EagleXiang
 * @Email: eagle.xiang@outlook.com
 * @Github: https://github.com/eaglexiang
 * @Date: 2019-03-19 20:08:35
 * @LastEditTime: 2019-08-25 13:13:50
 */

package et

import (
	"errors"
	"strings"

	speedlimitconn "github.com/eaglexiang/go-speedlimit-conn"

	"github.com/eaglexiang/eagle.tunnel.go/server/protocols/et/comm"
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
	// cipher
	conn := comm.NewCipherConn(tunnel.Left())
	tunnel.SetLeft(conn)
	// check auth
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

func (et *ET) checkHeaderOfReq(
	header string,
	tunnel *tunnel.Tunnel) error {
	headers := strings.Split(header, " ")
	switch {
	case len(headers) < 1:
		return errors.New("checkHeaderOfReq -> nil req")
	case headers[0] != comm.DefaultArg.Head:
		logger.Warning("invalid header of req: ", headers[0])
		return errors.New("checkHeaderOfReq -> wrong head")
	default:
		reply := "valid valid valid"
		_, err := tunnel.WriteLeft([]byte(reply))
		return err
	}
}

func (et *ET) checkUserOfReq(t *tunnel.Tunnel) (err error) {
	if comm.DefaultArg.ValidUsers == nil {
		// 未启用用户校验
		return
	}

	var user2Check *myuser.ReqUser
	if user2Check, err = findReqUser(t); err != nil {
		logger.Warning(err)
		return
	}
	sl, err := et._checkUserOfReq(user2Check)
	if err != nil {
		return
	}

	// set speed limiter
	left := speedlimitconn.New(t.Left(), sl)
	t.SetLeft(left)
	right := speedlimitconn.New(t.Right(), sl)
	t.SetRight(right)

	_, err = t.WriteLeft([]byte("valid"))
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
	validUser, ok := comm.DefaultArg.ValidUsers[user2Check.ID]
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
