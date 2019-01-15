/*
 * @Description:
 * @Author: EagleXiang
 * @Github: https://github.com/eaglexiang
 * @Date: 2019-01-13 06:34:08
 * @LastEditors: EagleXiang
 * @LastEditTime: 2019-01-13 07:26:25
 */

package eagletunnel

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/http"
	"sync"
)

// Service ET服务
type Service struct {
	sync.Mutex
	listener net.Listener
	running  bool
	reqs     chan net.Conn
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

	ipe := LocalAddr + ":" + LocalPort
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
		relayer := Relayer{}
		go relayer.Handle(req)
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
