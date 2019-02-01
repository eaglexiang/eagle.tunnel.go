/*
 * @Description:
 * @Author: EagleXiang
 * @Github: https://github.com/eaglexiang
 * @Date: 2019-01-13 06:34:08
 * @LastEditors: EagleXiang
 * @LastEditTime: 2019-02-01 12:32:28
 */

package etcore

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	mycipher "github.com/eaglexiang/go-cipher"
	"github.com/eaglexiang/go-simplecipher"
	myuser "github.com/eaglexiang/go-user"

	myet "./et"
)

// LocalUser 本地用户
var LocalUser *myuser.User

// Users 所有授权用户
var Users *map[string]*myuser.User

// Debug 开启Debug
var Debug bool

// Service ET服务
// 必须使用CreateService方法进行构造
type Service struct {
	sync.Mutex
	listener net.Listener
	running  bool
	reqs     chan net.Conn
	relayer  *Relayer
}

// CreateService 构造Service
func CreateService() *Service {
	mycipher.DefaultCipher = func() mycipher.Cipher {
		cipherType := mycipher.ParseCipherType(ConfigKeyValues["cipher"])
		switch cipherType {
		case mycipher.SimpleCipherType:
			c := simplecipher.SimpleCipher{}
			c.SetKey(ConfigKeyValues["data-key"])
			return &c
		default:
			return nil
		}
	}

	service := Service{
		reqs:    make(chan net.Conn),
		relayer: CreateRelayer(Debug),
	}

	et := myet.CreateET(
		ProxyStatus,
		ConfigKeyValues["ip-type"],
		ConfigKeyValues["head"],
		ConfigKeyValues["relayer"],
		ConfigKeyValues["location"],
		LocalUser,
		Users,
		time.Second*time.Duration(Timeout),
	)

	// 添加后端协议Handler
	if ConfigKeyValues["et"] == "on" {
		service.relayer.AddHandler(et)
	}
	if ConfigKeyValues["http"] == "on" {
		service.relayer.AddHandler(&HTTPProxy{})
	}
	if ConfigKeyValues["socks"] == "on" {
		service.relayer.AddHandler(&Socks5{})
	}
	for name, h := range AllHandlers {
		v, ok := ConfigKeyValues[name]
		if !ok {
			continue
		}
		if v == "on" {
			service.relayer.AddHandler(h)
		}
	}

	// 设置后端协议Sender
	service.relayer.SetSender(et)
	if DefaultSender != nil {
		service.relayer.SetSender(DefaultSender)
	}
	return &service
}

// Start 启动ET服务
func (s *Service) Start() (err error) {
	s.Lock()
	defer s.Unlock()
	if s.running {
		return errors.New("Service.Start -> the service is already started")
	}

	// disable tls check for ip-inside cache
	http.DefaultTransport.(*http.Transport).TLSClientConfig =
		&tls.Config{InsecureSkipVerify: true}

	ipe := ConfigKeyValues["listen"]
	s.listener, err = net.Listen("tcp", ipe)
	if err != nil {
		return errors.New("Service.Start -> " + err.Error())
	}
	fmt.Println("start to listen: ", ipe)
	s.reqs = make(chan net.Conn, 10)
	go s.listen()
	go s.handle()

	s.running = true
	return nil
}

func (s *Service) listen() {
	for s.running {
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
	for s.running {
		req, ok := <-s.reqs
		if !ok {
			return
		}
		go s._Handle(req)
	}
}

func (s *Service) _Handle(req net.Conn) {
	err := s.relayer.Handle(req)
	if err != nil {
		fmt.Println(err.Error())
	}
}

// Close 关闭服务
func (s *Service) Close() {
	s.Lock()
	defer s.Unlock()
	if s.running == false {
		return
	}
	s.running = false
	s.listener.Close()
	close(s.reqs)
}