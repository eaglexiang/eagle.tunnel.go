/*
 * @Author: EagleXiang
 * @Github: https://github.com/eaglexiang
 * @Date: 2019-01-13 06:34:08
 * @LastEditors: EagleXiang
 * @LastEditTime: 2019-03-17 16:34:08
 */

package core

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	myet "github.com/eaglexiang/eagle.tunnel.go/src/core/protocols/et"
	httpproxy "github.com/eaglexiang/eagle.tunnel.go/src/core/protocols/httpproxy"
	socks5 "github.com/eaglexiang/eagle.tunnel.go/src/core/protocols/socks5"
	logger "github.com/eaglexiang/eagle.tunnel.go/src/logger"
	mycipher "github.com/eaglexiang/go-cipher"
	settings "github.com/eaglexiang/go-settings"
	myuser "github.com/eaglexiang/go-user"
)

// LocalUser 本地用户
var LocalUser *myuser.User

// Users 所有授权用户
var Users map[string]*myuser.User

// Service ET服务
// 必须使用CreateService方法进行构造
type Service struct {
	sync.Mutex
	listener    net.Listener
	stopRunning chan interface{}
	reqs        chan net.Conn
	relay       Relay
	MaxClients  int              // 最大客户数量
	clients     chan interface{} // 当前客户，用来统计当前客户数量
}

func createCipher() mycipher.Cipher {
	cipherType := mycipher.ParseCipherType(settings.Get("cipher"))
	switch cipherType {
	case mycipher.SimpleCipherType:
		c := mycipher.SimpleCipher{}
		c.SetKey(settings.Get("data-key"))
		return &c
	default:
		logger.Error("invalid cipher: ", settings.Get("cipher"))
		return nil
	}
}

// CreateETArg 构建ET.Arg
func CreateETArg() *myet.Arg {
	users := myet.UsersArg{
		LocalUser:  LocalUser,
		ValidUsers: Users,
	}
	connArg := myet.ConnArg{
		RemoteIPE: settings.Get("relay"),
		Timeout:   Timeout,
		Head:      settings.Get("head"),
	}
	smartArg := myet.SmartArg{
		ProxyStatus:   ProxyStatus,
		LocalLocation: settings.Get("location"),
	}

	return &myet.Arg{
		ConnArg:  connArg,
		SmartArg: smartArg,
		UsersArg: users,
		IPType:   settings.Get("ip-type"),
	}
}

func setHandlersAndSender(service *Service) {
	et := myet.CreateET(CreateETArg())

	// 添加后端协议Handler
	if settings.Get("et") == "on" {
		service.relay.AddHandler(et)
	}
	if settings.Get("http") == "on" {
		service.relay.AddHandler(&httpproxy.HTTPProxy{})
	}
	if settings.Get("socks") == "on" {
		service.relay.AddHandler(&socks5.Socks5{})
	}
	for name, h := range AllHandlers {
		if !settings.Exsit(name) {
			continue
		}
		if settings.Get(name) == "on" {
			service.relay.AddHandler(h)
		}
	}

	// 设置后端协议Sender
	service.relay.SetSender(et)
	if DefaultSender != nil {
		service.relay.SetSender(DefaultSender)
	}
}

func setMaxClients(service *Service) {
	maxclients, _ := strconv.ParseInt(settings.Get("maxclients"), 10, 64)
	if maxclients > 0 {
		service.clients = make(chan interface{}, maxclients)
	}
}

// CreateService 构造Service
func CreateService() *Service {
	mycipher.DefaultCipher = createCipher
	service := &Service{
		reqs:  make(chan net.Conn),
		relay: Relay{},
	}
	setHandlersAndSender(service)
	setMaxClients(service)
	return service
}

// Start 启动ET服务
func (s *Service) Start() (err error) {
	s.Lock()
	defer s.Unlock()
	if s.stopRunning != nil {
		logger.Error("fail to start the already started service")
		return errors.New("service is already started")
	}

	// disable tls check for ET-LOCATION
	http.DefaultTransport.(*http.Transport).TLSClientConfig =
		&tls.Config{InsecureSkipVerify: true}

	ipe := settings.Get("listen")
	s.listener, err = net.Listen("tcp", ipe)
	if err != nil {
		logger.Error(err)
		return err
	}
	fmt.Println("start to listen: ", ipe)
	s.reqs = make(chan net.Conn)
	go s.listen()
	go s.handle()

	s.stopRunning = make(chan interface{})
	return nil
}

func (s *Service) listen() {
	for s.stopRunning != nil {
		req, err := s.listener.Accept()
		if err != nil {
			fmt.Println(err.Error())
			s.Close()
			return
		}
		s.reqs <- req
	}
}

func (s *Service) handle() {
	for {
		select {
		case req, ok := <-s.reqs:
			if !ok {
				return
			}
			if s.clients != nil {
				s.clients <- new(interface{})
			}
			if Timeout != 0 {
				req.SetReadDeadline(time.Now().Add(Timeout))
			}
			go s._Handle(req)
		case <-s.stopRunning:
			break
		}
	}
}

func (s *Service) _Handle(req net.Conn) {
	s.relay.Handle(req)
	if s.clients != nil {
		<-s.clients
	}
}

// Close 关闭服务
func (s *Service) Close() {
	s.Lock()
	defer s.Unlock()
	if s.stopRunning == nil {
		return
	}
	close(s.stopRunning)
	s.stopRunning = nil
	s.listener.Close()
	close(s.reqs)
}
