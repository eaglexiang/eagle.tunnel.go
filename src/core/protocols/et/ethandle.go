package et

import (
	"errors"
	"reflect"
	"strings"

	"github.com/eaglexiang/eagle.tunnel.go/src/logger"
	mycipher "github.com/eaglexiang/go-cipher"
	mynet "github.com/eaglexiang/go-net"
	mytunnel "github.com/eaglexiang/go-tunnel"
	myuser "github.com/eaglexiang/go-user"
	"go.uber.org/ratelimit"
)

// Handle 处理ET请求
func (et *ET) Handle(e *mynet.Arg) (err error) {
	err = et.checkHeaderOfReq(string(e.Msg), e.Tunnel)
	if err != nil {
		return err
	}
	createCipher(e.Tunnel)
	err = et.checkUserOfReq(e.Tunnel)
	if err != nil {
		return err
	}
	// 选择子协议handler
	subReq, handler, err := et.subShake(e.Tunnel)
	if err != nil {
		return err
	}
	// 进入子协议业务
	err = handler.Handle(subReq, e.Tunnel)
	if err != nil {
		return err
	}
	// 只有TCP子协议需要继续运行
	if reflect.TypeOf(handler) != reflect.TypeOf(TCP{}) {
		return errors.New("no need to continue")
	}
	return nil
}

func (et *ET) subShake(tunnel *mytunnel.Tunnel) (subReq string,
	handler Handler, err error) {
	subReq, err = tunnel.ReadLeftStr()
	if err != nil {
		return "", nil, err
	}
	handler = getHandler(subReq, et.subHandlers)
	if handler == nil {
		logger.Warning("invalid req: ", subReq)
	}
	return subReq, handler, nil
}

func createCipher(tunnel *mytunnel.Tunnel) {
	c := mycipher.DefaultCipher()
	if c == nil {
		panic("ET.Handle -> cipher is nil")
	}
	tunnel.EncryptLeft = c.Encrypt
	tunnel.DecryptLeft = c.Decrypt
}

func (et *ET) checkHeaderOfReq(
	header string,
	tunnel *mytunnel.Tunnel) error {
	headers := strings.Split(header, " ")
	switch {
	case len(headers) < 1:
		return errors.New("checkHeaderOfReq -> nil req")
	case headers[0] != et.arg.Head:
		logger.Warning("invalid header of req: ", headers[0])
		return errors.New("checkHeaderOfReq -> wrong head")
	default:
		reply := "valid valid valid"
		_, err := tunnel.WriteLeft([]byte(reply))
		return err
	}
}

func (et *ET) checkUserOfReq(tunnel *mytunnel.Tunnel) (err error) {
	if et.arg.ValidUsers == nil {
		// 未启用用户校验
		return nil
	}
	var user2Check *myuser.ReqUser
	if user2Check, err = findReqUser(tunnel); err != nil {
		logger.Warning(err)
		return err
	}
	if tunnel.SpeedLimiter, err = et._checkUserOfReq(user2Check); err == nil {
		_, err = tunnel.WriteLeft([]byte("valid"))
	}
	return err
}

func findReqUser(tunnel *mytunnel.Tunnel) (*myuser.ReqUser, error) {
	userStr, err := tunnel.ReadLeftStr()
	if err != nil {
		return nil, err
	}
	user2Check, err := parseReqUser(
		userStr,
		mynet.GetIPOfConnRemote(tunnel.Left))
	return user2Check, err
}

func (et *ET) _checkUserOfReq(user2Check *myuser.ReqUser) (limiter *ratelimit.Limiter, err error) {
	validUser, ok := et.arg.ValidUsers[user2Check.ID]
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
		return nil, err
	}
	if user2Check.ID == "null" {
		return nil, errors.New("invalid user")
	}
	return user2Check, nil
}
