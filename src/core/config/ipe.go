package config

import (
	"errors"
	"net"
	"strings"

	mynet "github.com/eaglexiang/go-net"
)

const (
	ipeSplitSig = ","
	defaultPort = "8080"
)

// ipPorts 每个ipPorts由一个IP与多个Port组成，ports的长度至少为1
type ipPorts struct {
	ip    string
	ports []string
}

func (ip *ipPorts) setDefaultPort(port string) {
	if len(ip.ports) == 0 || ip.ports[0] == "" {
		ip.ports = []string{port}
	}
}

func (ip *ipPorts) addPort(port string) {
	for _, p := range ip.ports {
		if p == port {
			return
		}
	}

	ip.ports = append(ip.ports, port)
}

func (ip ipPorts) toString() (result string) {
	for _, port := range ip.ports {
		ipe := ip.ip + ":" + port

		if result != "" {
			result += ipeSplitSig
		}
		result += ipe
	}
	return
}

func parseIPPortsSlice(src string) []*ipPorts {
	// map[ipPorts.ip] ipPorts
	ipPortsMap := make(map[string]*ipPorts)

	ipes := strings.Split(src, ipeSplitSig)
	for _, ipe := range ipes {
		ipeArgs := strings.Split(ipe, ipeSplitSig)
		ip := ipeArgs[0]
		port := ipeArgs[1]

		if ipports, ok := ipPortsMap[ip]; ok {
			ipports.addPort(port)
		} else {
			ipPortsMap[ip] = &ipPorts{
				ip:    ip,
				ports: []string{port},
			}
		}
	}

	// 设置默认端口
	for _, ip := range ipPortsMap {
		ip.setDefaultPort(defaultPort)
	}

	ipPortsSlice := []*ipPorts{}
	for _, ip := range ipPortsMap {
		ipPortsSlice = append(ipPortsSlice, ip)
	}

	return ipPortsSlice
}

func getIPPort(ipe string) (ip, port string, err error) {
	if strings.HasPrefix(ipe, "[") {
		ip, port, err = getIPPortFromIPv6IPE(ipe)
	} else {
		ip, port, err = getIPPortFromIPv4IPE(ipe)
	}

	return
}

func getIPPortFromIPv4IPE(ipe string) (ip, port string, err error) {
	ipeSlice := strings.Split(ipe, ":")

	ip = ipeSlice[0]

	if len(ipeSlice) > 1 {
		port = ipeSlice[1]
	}

	return
}

func getIPPortFromIPv6IPE(ipe string) (ip, port string, err error) {
	ipeSlice := strings.Split(ipe, ":")

	if len(ipeSlice) == 6 {
		if !strings.HasSuffix(ipe, "]") {
			err = errors.New("ipv6 need ]")
		} else {
			ip = ipe
		}
	} else if len(ipeSlice) == 7 {
		ip = strings.Join(ipeSlice[:6], ":")
		port = ipeSlice[6]
	} else {
		err = errors.New("invalid ipe")
	}

	return
}

// finishPort 补全端口号
func finishIPEPort(ipe string) string {
	switch mynet.TypeOfAddr(ipe) {
	case mynet.IPv4Addr:
		if ip := net.ParseIP(ipe); ip != nil {
			// 不包含端口号
			ipe += ":8080"
		}
	case mynet.IPv6Addr:
		if strings.HasSuffix(ipe, "]") {
			// 不包含端口号
			ipe += ":8080"
		}
	}
	return ipe
}

// finishIPEs ipes的示例：192.168.0.1:8080,192.168.0.1:8081
func finishIPEs(ipes string) (newIPEs string) {
	_ipes := strings.Split(ipes, ",")
	for _, ipe := range _ipes {
		newIPEs += "," + finishIPEPort(ipe)
	}
	newIPEs = strings.TrimPrefix(newIPEs, ",") // 去掉头部多余的,符号
	return
}
