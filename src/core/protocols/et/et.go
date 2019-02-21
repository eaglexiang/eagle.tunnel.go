/*
 * @Author: EagleXiang
 * @Github: https://github.com/eaglexiang
 * @Date: 2018-12-27 08:24:57
 * @LastEditors: EagleXiang
 * @LastEditTime: 2019-02-22 00:44:43
 */

package et

import (
	"errors"
	"net"
	"reflect"
	"strings"

	"github.com/eaglexiang/go-bytebuffer"

	mycipher "github.com/eaglexiang/go-cipher"
	mynet "github.com/eaglexiang/go-net"
	mytunnel "github.com/eaglexiang/go-tunnel"
	myuser "github.com/eaglexiang/go-user"
	version "github.com/eaglexiang/go-version"
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
	et := ET{arg: arg}
	dns := DNS{arg: arg}
	dns6 := DNS6{arg: arg}
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
	if et.subSenders == nil {
		et.subSenders = make(map[int]Sender)
	}
	et.subSenders[sender.Type()] = sender
}

// Match 判断请求消息是否匹配该业务
func (et *ET) Match(firstMsg []byte) bool {
	firstMsgStr := string(firstMsg)
	return firstMsgStr == et.arg.Head
}

// Handle 处理ET请求
func (et *ET) Handle(e *mynet.Arg) error {
	args := strings.Split(string(e.Msg), " ")
	err := et.checkHeaderOfReq(args, e.Tunnel) // 检查协议头
	if err != nil {
		return errors.New("ET.Handle -> " + err.Error())
	}
	createCipher(e.Tunnel)            // 创建cipher
	err = et.checkUserOfReq(e.Tunnel) // 检查请求用户
	if err != nil {
		return errors.New("ET.Handle -> " + err.Error())
	}
	// 选择子协议handler
	subReq, err := e.Tunnel.ReadLeftStr()
	if err != nil {
		return errors.New("ET.Handle -> " + err.Error())
	}
	handler := getHandler(subReq, et.subHandlers)
	if handler == nil {
		return errors.New("ET.Handle -> invalid req: " +
			subReq)
	}
	// 进入子协议业务
	err = handler.Handle(subReq, e.Tunnel)
	if err != nil {
		return errors.New("ET.Handle -> " +
			err.Error())
	}
	// 只有TCP子协议需要继续运行
	if reflect.TypeOf(handler) != reflect.TypeOf(TCP{}) {
		return errors.New("no need to continue")
	}
	return nil
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
	// 外部Send请求只允许TCP请求
	sender, ok := et.subSenders[EtTCP]
	if !ok {
		return errors.New("ET.Send -> no tcp sender")
	}
	// 进入子协议业务
	newE, err := parseNetArg(e)
	if err != nil {
		return errors.New("ET.Send -> " +
			err.Error())
	}
	err = sender.Send(et, newE)
	if err != nil {
		return errors.New("ET.Send -> " +
			err.Error())
	}
	return nil
}

// Name Sender的名字
func (et *ET) Name() string {
	return "ET"
}

// connect2Relayer 连接到下一个Relayer，完成版本校验和用户校验两个步骤
func (et *ET) connect2Relayer(tunnel *mytunnel.Tunnel) error {
	conn, err := net.DialTimeout("tcp", et.arg.RemoteET, et.arg.Timeout)
	if err != nil {
		return errors.New("connect2Relayer -> " + err.Error())
	}
	tunnel.Right = &conn
	err = et.checkVersionOfRelayer(tunnel)
	if err != nil {
		return errors.New("connect2Relayer -> " + err.Error())
	}
	c := mycipher.DefaultCipher()
	if c == nil {
		return errors.New("ET.connect2Relayer -> cipher is nil")
	}
	tunnel.EncryptRight = c.Encrypt
	tunnel.DecryptRight = c.Decrypt
	err = et.checkUserOfLocal(tunnel)
	if err != nil {
		return errors.New("connect2Relayer -> " + err.Error())
	}
	return nil
}

func (et *ET) checkVersionOfRelayer(tunnel *mytunnel.Tunnel) error {
	req := et.arg.Head
	_, err := tunnel.WriteRight([]byte(req))
	if err != nil {
		return errors.New("checkVersionOfRelayer -> " + err.Error())
	}
	buffer := bytebuffer.GetKBBuffer()
	defer bytebuffer.PutKBBuffer(buffer)
	buffer.Length, err = tunnel.ReadRight(buffer.Buf())
	if err != nil {
		return errors.New("checkVersionOfRelayer -> " + err.Error())
	}
	reply := buffer.String()
	if reply != "valid valid valid" {
		return errors.New("checkVersionOfRelayer -> " + reply)
	}
	return nil
}

func (et *ET) checkHeaderOfReq(
	headers []string,
	tunnel *mytunnel.Tunnel) error {
	if len(headers) < 1 {
		return errors.New("checkHeaderOfReq -> nil req")
	}
	if headers[0] != et.arg.Head {
		return errors.New("checkHeaderOfReq -> wrong head: " + headers[0])
	}
	reply := "valid valid valid"
	count, _ := tunnel.WriteLeft([]byte(reply))
	if count != 17 {
		return errors.New("checkHeaderOfReq -> fail to reply")
	}
	return nil
}

func (et *ET) checkUserOfLocal(tunnel *mytunnel.Tunnel) (err error) {
	if et.arg.Users.LocalUser.ID() == "null" {
		return nil // no need to check
	}
	user := et.arg.Users.LocalUser.ToString()
	_, err = tunnel.WriteRight([]byte(user))
	if err != nil {
		return errors.New("checkUserOfLocal -> " + err.Error())
	}
	buffer := bytebuffer.GetKBBuffer()
	defer bytebuffer.PutKBBuffer(buffer)
	buffer.Length, err = tunnel.ReadRight(buffer.Buf())
	if err != nil {
		return errors.New("checkUserOfLocal -> " + err.Error())
	}
	reply := buffer.String()
	if reply != "valid" {
		return errors.New("checkUserOfLocal -> invalid reply: " + reply)
	}
	tunnel.SpeedLimiter = et.arg.Users.LocalUser.SpeedLimiter()
	return nil
}

func (et *ET) checkUserOfReq(tunnel *mytunnel.Tunnel) (err error) {
	if et.arg.Users.ValidUsers == nil {
		return nil
	}
	// 接收用户信息
	userStr, err := tunnel.ReadLeftStr()
	if err != nil {
		return errors.New("ET.checkUserOfReq -> " +
			err.Error())
	}
	// 获取用户IP
	addr := (*tunnel.Left).RemoteAddr()
	ip := strings.Split(addr.String(), ":")[0]
	user2Check, err := myuser.ParseReqUser(userStr, ip)
	if err != nil {
		tunnel.WriteLeft([]byte(err.Error()))
		return errors.New("checkUserOfReq -> " + err.Error())
	}
	if user2Check.ID == "null" {
		return errors.New("checkUserOfReq -> " +
			"username shouldn't be 'null'")
	}
	validUser, ok := et.arg.Users.ValidUsers[user2Check.ID]
	if !ok {
		// 找不到该用户
		reply := "incorrent username or password"
		tunnel.WriteLeft([]byte(reply))
		return errors.New("checkUserOfReq -> username not found: " +
			user2Check.ID)
	}
	err = validUser.CheckAuth(user2Check)
	if err != nil {
		reply := err.Error()
		tunnel.WriteLeft([]byte(reply))
		return errors.New("checkUserOfReq -> " + err.Error())
	}
	reply := "valid"
	tunnel.WriteLeft([]byte(reply))
	tunnel.SpeedLimiter = validUser.SpeedLimiter()
	return nil
}

// 查询类请求的发射过程都是类似的
// 连接 - 发送请求 - 得到反馈
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
