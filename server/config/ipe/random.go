/*
 * @Author: EagleXiang
 * @LastEditors: EagleXiang
 * @Email: eagle.xiang@outlook.com
 * @Github: https://github.com/eaglexiang
 * @Date: 2019-08-28 19:50:17
 * @LastEditTime: 2019-08-28 19:50:21
 */

package ipe

import (
	"fmt"
	"math/rand"
)

func randPort() string {
	newPort := rand.Intn(1000)
	newPort += 1000
	return fmt.Sprint(newPort)
}

func randPorts(n int) (ports []string) {
	for i := 0; i < n; i++ {
		ports = append(ports, randPort())
	}
	return
}

func (ip *IPPorts) RandPorts(n int) {
	ports := randPorts(n)
	for _, port := range ports {
		ip.addPort(port)
	}
}
