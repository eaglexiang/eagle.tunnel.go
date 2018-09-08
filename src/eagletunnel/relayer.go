package eagletunnel

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
)

const protocolVersion string = "1.1"

var LocalAddr string
var LocalPort string

type Relayer struct {
	listener net.Listener
}

func (relayer *Relayer) Start() {
	var err error

	// disable tls check for ip-inside cache
	http.DefaultTransport.(*http.Transport).TLSClientConfig =
		&tls.Config{InsecureSkipVerify: true}

	ipe := LocalAddr + ":" + LocalPort
	relayer.listener, err = net.Listen("tcp", ipe)
	if err != nil {
		fmt.Println("error: failed to listen! ", err)
	} else {
		fmt.Println("start to listen: ", ipe)
		relayer.listen()
	}
}

func (relayer *Relayer) listen() {
	for {
		conn, err := relayer.listener.Accept()
		if err != nil {
			fmt.Println("error: failed to accept! ", err)
			break
		} else {
			go relayer.handleClient(conn)
		}
	}
	fmt.Println("quit")
}

func (relayer *Relayer) handleClient(conn net.Conn) {
	var buffer = make([]byte, 1024)
	count, err := conn.Read(buffer)
	if err != nil {
		return
	}
	request := Request{requestMsg: buffer[:count]}
	tunnel := Tunnel{left: &conn, encryptKey: EncryptKey}
	var handler Handler
	switch request.getType() {
	case EAGLE_TUNNEL:
		if EnableET {
			handler = new(EagleTunnel)
		}
	case HTTP_PROXY:
		if EnableHTTP {
			handler = new(HttpProxy)
		}
	case SOCKS:
		if EnableSOCKS5 {
			handler = new(Socks5)
		}
	default:
		handler = nil
	}
	if handler != nil {
		result := handler.handle(request, &tunnel)
		if result {
			tunnel.flow()
		} else {
			tunnel.close()
		}
	} else {
		tunnel.close()
	}
}
