/*
 * @Author: EagleXiang
 * @Github: https://github.com/eaglexiang
 * @Date: 2019-01-13 06:34:08
 * @LastEditors: EagleXiang
 * @LastEditTime: 2019-08-06 21:14:25
 */

package server

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/eaglexiang/eagle.tunnel.go/src/core/config"
	"github.com/eaglexiang/go-settings"

	et "github.com/eaglexiang/eagle.tunnel.go/src/core/protocols/et/core"
	httpproxy "github.com/eaglexiang/eagle.tunnel.go/src/core/protocols/httpproxy"
	socks5 "github.com/eaglexiang/eagle.tunnel.go/src/core/protocols/socks5"
	mycipher "github.com/eaglexiang/go-cipher"
	"github.com/eaglexiang/go-counter"
	logger "github.com/eaglexiang/go-logger"
)

// Service ET服务
// 必须使用CreateService方法进行构造
type Service struct {
	sync.Mutex
	listeners   []net.Listener
	stopRunning chan interface{}
	reqs        chan net.Conn
	relay       Relay
	counter     counter.Counter // 当前请求的数量
	maxCount    int64           // 当前请求的最大数量
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

func setHandlersAndSender(service *Service) {
	relayIPE := config.RelayIPE()
	et := et.NewET(config.CreateETArg(relayIPE))

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
	maxclients, err := strconv.ParseInt(settings.Get("maxclients"), 10, 64)
	if err != nil {
		panic(err)
	}
	service.maxCount = maxclients
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

	s.disableTLS()
	s.start2ListenIPEs()

	s.reqs = make(chan net.Conn)

	go s.listen()
	go s.handleReqs()

	s.stopRunning = make(chan interface{})
	return
}

// disableTLS disable tls check for ET-LOCATION
func (s *Service) disableTLS() {
	http.DefaultTransport.(*http.Transport).TLSClientConfig =
		&tls.Config{InsecureSkipVerify: true}
}

// start2ListenIPEs 开始所有IPE的监听
func (s *Service) start2ListenIPEs() {
	for _, ipPorts := range config.ListenIPEs {
		ipes := ipPorts.ToStrings()
		for _, ipe := range ipes {
			s.listenIPE(ipe)
		}
	}
}

func (s *Service) listen() {
	for _, listener := range s.listeners {
		go s.startListener(listener)
	}
}

func (s *Service) startListener(listener net.Listener) {
	defer fmt.Println("quit listener: ", listener.Addr())

	for s.stopRunning != nil {
		req, err := listener.Accept()
		if err != nil {
			fmt.Println(err.Error())
			s.Close()
			return
		}
		s.reqs <- req
	}
}

// full 当前请求数已满
func (s *Service) full() bool {
	if s.maxCount == 0 {
		return false
	}
	if s.counter.Value > s.maxCount {
		return true
	}
	return false
}

// clientsUp 当前请求数+1
func (s *Service) clientsUp() {
	if s.maxCount == 0 {
		return
	}
	s.counter.Up()
	logger.Info("clients now: ", s.counter.Value)
}

// clientsDown 当前请求数-1
func (s *Service) clientsDown() {
	if s.maxCount == 0 {
		return
	}
	s.counter.Down()
	logger.Info("clients now: ", s.counter.Value)
}

func (s *Service) handleReqs() {
	for {
		req, ok, err := s.recvReq()
		if err != nil {
			break
		}
		if ok {
			go s.handleReq(req)
		}
	}
}

func (s *Service) recvReq() (req net.Conn, ok bool, err error) {
	if s.full() {
		logger.Warning("over working... ===> ", s.counter.Value)
		time.Sleep(time.Millisecond * 100)
		return
	}

	select {
	case req, ok = <-s.reqs:
		if !ok {
			err = errors.New("no new req")
		}
		return
	case <-s.stopRunning:
		err = errors.New("stopped")
		return
	}
}

func (s *Service) handleReq(req net.Conn) {
	s.clientsUp()

	if config.Timeout != 0 {
		req.SetReadDeadline(time.Now().Add(config.Timeout))
	}
	s.relay.Handle(req)

	s.clientsDown()
	return
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
	for _, lis := range s.listeners {
		lis.Close()
	}
	close(s.reqs)
}

func (s *Service) listenIPE(ipe string) {
	if listener, err := net.Listen("tcp", ipe); err != nil {
		fmt.Println(err)
	} else {
		s.listeners = append(s.listeners, listener)
		fmt.Println("start to listen: ", ipe)
	}
}
