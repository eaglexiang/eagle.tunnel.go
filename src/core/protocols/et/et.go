/*
 * @Author: EagleXiang
 * @Github: https://github.com/eaglexiang
 * @Date: 2018-12-27 08:24:57
 * @LastEditors: EagleXiang
 * @LastEditTime: 2019-02-21 18:18:32
 */

package et

import (
	"errors"
	"net"
	"reflect"
	"strings"
	"time"

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
	proxyStatus   int
	head          string
	remoteET      string
	localLocation string
	localUser     *myuser.User
	validUsers    map[string]*myuser.User
	subHandlers   []Handler
	subSenders    map[int]Sender
	timeout       time.Duration
}

// CreateET 构造ET
func CreateET(
	proxyStatus int,
	ipType string,
	head string,
	remoteET string,
	localLocation string,
	localUser *myuser.User,
	validUsers map[string]*myuser.User,
	timeout time.Duration,
) *ET {
	et := ET{
		proxyStatus:   proxyStatus,
		head:          head,
		remoteET:      remoteET,
		localLocation: localLocation,
		localUser:     localUser,
		validUsers:    validUsers,
		timeout:       timeout,
	}

	dns := DNS{ProxyStatus: et.proxyStatus}
	dns6 := DNS6{ProxyStatus: et.proxyStatus}
	tcp := createTCP(
		et.proxyStatus,
		localUser.SpeedLimiter(),
		timeout,
		ipType,
		dns,
		dns6,
	)
	location := createLocation(et.localLocation)
	check := Check{validUsers: validUsers}

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
	return firstMsgStr == et.head
}

// Handle 处理ET请求
func (et *ET) Handle(e *mynet.Arg) error {
	// 验证流程
	args := strings.Split(string(e.Msg), " ")
	err := et.checkHeaderOfReq(args, e.Tunnel) // 检查握手头
	if err != nil {
		return errors.New("ET.Handle -> " + err.Error())
	}
	c := mycipher.DefaultCipher() // 创建Cipher
	if c == nil {
		return errors.New("ET.Handle -> cipher is nil")
	}
	e.Tunnel.EncryptLeft = c.Encrypt
	e.Tunnel.DecryptLeft = c.Decrypt

	err = et.checkUserOfReq(e.Tunnel) // 检查请求用户
	if err != nil {
		return errors.New("ET.Handle -> " + err.Error())
	}
	// 选择子协议handler
	buffer := bytebuffer.GetKBBuffer()
	defer bytebuffer.PutKBBuffer(buffer)
	buffer.Length, err = e.Tunnel.ReadLeft(buffer.Buf())
	if err != nil {
		return errors.New("ET.Handle -> get sub-header: " +
			err.Error())
	}
	subReq := buffer.String()
	var handler Handler
	for _, h := range et.subHandlers {
		if h.Match(subReq) {
			handler = h
			break
		}
	}
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
	conn, err := net.DialTimeout("tcp", et.remoteET, et.timeout)
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
	req := et.head
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
	if headers[0] != et.head {
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
	if et.localUser.ID() == "null" {
		return nil // no need to check
	}
	user := et.localUser.ToString()
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
	tunnel.SpeedLimiter = et.localUser.SpeedLimiter()
	return nil
}

func (et *ET) checkUserOfReq(tunnel *mytunnel.Tunnel) (err error) {
	if et.validUsers == nil {
		return nil
	}
	// 接收用户信息
	buffer := bytebuffer.GetKBBuffer()
	defer bytebuffer.PutKBBuffer(buffer)
	buffer.Length, err = tunnel.ReadLeft(buffer.Buf())
	if err != nil {
		return errors.New("checkUserOfReq -> " + err.Error())
	}
	userStr := buffer.String()
	addr := (*tunnel.Left).RemoteAddr()
	ip := strings.Split(addr.String(), ":")[0]
	user2Check, err := myuser.ParseReqUser(userStr, ip)
	if err != nil {
		tunnel.WriteLeft([]byte(err.Error()))
		return errors.New("checkUserOfReq -> " + err.Error())
	}
	if user2Check.ID == "null" {
		reply := "username shouldn't be 'null'"
		tunnel.WriteLeft([]byte(reply))
		return errors.New("checkUserOfReq -> " + reply)
	}
	validUser, ok := et.validUsers[user2Check.ID]
	if !ok {
		// 找不到该用户
		reply := "incorrent username or password"
		tunnel.WriteLeft([]byte(reply))
		return errors.New("checkUserOfReq -> username not found: " + user2Check.ID)
	}
	err = validUser.CheckAuth(user2Check)
	if err != nil {
		reply := err.Error()
		tunnel.WriteLeft([]byte(reply))
		return errors.New("checkUserOfReq -> " + err.Error())
	}
	reply := "valid"
	count, _ := tunnel.WriteLeft([]byte(reply))
	if count != 5 {
		// 发送响应失败
		return errors.New("checkUserOfReq -> wrong reply")
	}
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
