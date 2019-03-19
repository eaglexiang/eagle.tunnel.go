/*
 * @Author: EagleXiang
 * @Github: https://github.com/eaglexiang
 * @Date: 2018-12-27 08:24:57
 * @LastEditors: EagleXiang
 * @LastEditTime: 2019-03-17 20:27:12
 */

package et

import (
	"errors"
	"net"
	"reflect"
	"strings"

	logger "github.com/eaglexiang/eagle.tunnel.go/src/logger"
	bytebuffer "github.com/eaglexiang/go-bytebuffer"
	mycipher "github.com/eaglexiang/go-cipher"
	mynet "github.com/eaglexiang/go-net"
	mytunnel "github.com/eaglexiang/go-tunnel"
	myuser "github.com/eaglexiang/go-user"
	version "github.com/eaglexiang/go-version"
	"go.uber.org/ratelimit"
)

// ProtocolVersion 作为Sender使用的协议版本号
var ProtocolVersion, _ = version.CreateVersion("1.5")

// ProtocolCompatibleVersion 作为Handler可兼容的最低协议版本号
var ProtocolCompatibleVersion, _ = version.CreateVersion("1.3")

// ET ET代理协议的实现
// 必须使用CreateET来构造该结构
type ET struct {
	arg         *Arg
	subHandlers []Handler
	subSenders  map[int]Sender
}

// CreateET 构造ET
func CreateET(arg *Arg) *ET {
	et := ET{
		arg:        arg,
		subSenders: make(map[int]Sender),
	}
	dns := DNS{arg: arg, dnsResolver: mynet.ResolvIPv4}
	dns6 := DNS{arg: arg, dnsResolver: mynet.ResolvIPv6}
	tcp := TCP{
		arg:  arg,
		dns:  dns,
		dns6: dns6,
	}
	location := Location{arg: arg}
	check := Check{arg: arg}

	// 添加子协议的handler
	et.AddSubHandler(tcp)
	et.AddSubHandler(dns)
	et.AddSubHandler(dns6)
	et.AddSubHandler(location)
	et.AddSubHandler(check)

	// 添加子协议的sender
	et.AddSubSender(tcp)
	et.AddSubSender(dns)
	et.AddSubSender(dns6)
	et.AddSubSender(location)
	return &et
}

// AddSubHandler 添加ET子协议handler
func (et *ET) AddSubHandler(handler Handler) {
	et.subHandlers = append(et.subHandlers, handler)
}

// AddSubSender 添加子协议Sender
func (et *ET) AddSubSender(sender Sender) {
	et.subSenders[sender.Type()] = sender
}

// Match 判断请求消息是否匹配该业务
func (et *ET) Match(firstMsg []byte) bool {
	firstMsgStr := string(firstMsg)
	return firstMsgStr == et.arg.Head
}

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

func getHandler(subReq string, subHandlers []Handler) Handler {
	var handler Handler
	for _, h := range subHandlers {
		if h.Match(subReq) {
			handler = h
			break
		}
	}
	return handler
}

// Send 发送ET请求
func (et *ET) Send(e *mynet.Arg) error {
	// 选择Sender
	newE, err := parseNetArg(e)
	if err != nil {
		return err
	}
	sender, ok := et.subSenders[newE.TheType]
	if !ok {
		logger.Error("no tcp sender")
		return errors.New("no tcp sender")
	}
	// 进入子协议业务
	err = sender.Send(et, newE)
	if err != nil {
		return err
	}
	return nil
}

// Name Sender的名字
func (et *ET) Name() string {
	return "ET"
}

// connect2Relayer 连接到下一个Relayer，完成版本校验和用户校验两个步骤
func (et *ET) connect2Relayer(tunnel *mytunnel.Tunnel) error {
	conn, err := net.DialTimeout("tcp", et.arg.RemoteIPE, et.arg.Timeout)
	if err != nil {
		logger.Warning(err)
		return err
	}
	tunnel.Right = conn
	err = et.checkVersionOfRelayer(tunnel)
	if err != nil {
		return err
	}
	c := mycipher.DefaultCipher()
	if c == nil {
		panic("cipher is nil")
	}
	tunnel.EncryptRight = c.Encrypt
	tunnel.DecryptRight = c.Decrypt
	return et.checkUserOfLocal(tunnel)
}

func (et *ET) checkVersionOfRelayer(tunnel *mytunnel.Tunnel) error {
	req := et.arg.Head
	_, err := tunnel.WriteRight([]byte(req))
	if err != nil {
		return err
	}
	reply, err := tunnel.ReadRightStr()
	if reply != "valid valid valid" {
		logger.Warning("invalid reply for et version check: ",
			reply)
		return errors.New("invalid reply")
	}
	return nil
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

func (et *ET) checkUserOfLocal(tunnel *mytunnel.Tunnel) (err error) {
	if et.arg.LocalUser.ID() == "null" {
		return nil // no need to check
	}
	user := et.arg.LocalUser.ToString()
	_, err = tunnel.WriteRight([]byte(user))
	if err != nil {
		return err
	}
	reply, _ := tunnel.ReadRightStr()
	if reply != "valid" {
		logger.Error("invalid reply for check local user: ", reply)
		return errors.New("invalid reply")
	}
	tunnel.SpeedLimiter = et.arg.LocalUser.SpeedLimiter()
	return nil
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

// 查询类请求的发射过程都是类似的
// 连接 - 发送请求 - 得到反馈 - 关闭连接
// 区别仅仅在请求命令的内容
func sendQueryReq(et *ET, req string) string {
	tunnel := mytunnel.GetTunnel()
	defer mytunnel.PutTunnel(tunnel)
	err := et.connect2Relayer(tunnel)
	if err != nil {
		return "sendNormalEtCheckReq-> " + err.Error()
	}

	// 发送请求
	_, err = tunnel.WriteRight([]byte(req))
	if err != nil {
		return "sendNormalEtCheckReq-> " + err.Error()
	}

	// 接受回复
	buffer := bytebuffer.GetKBBuffer()
	defer bytebuffer.PutKBBuffer(buffer)
	buffer.Length, err = tunnel.ReadRight(buffer.Buf())
	if err != nil {
		return "sendNormalEtCheckReq-> " + err.Error()
	}
	reply := buffer.String()
	return reply
}
