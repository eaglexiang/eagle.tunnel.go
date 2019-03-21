package cmd

import "github.com/eaglexiang/eagle.tunnel.go/src/core/protocols/et/comm"

func sendQuery(s comm.Sender, req string) (string, error) {
	req = s.Name() + " " + req
	return comm.SendQueryReq(req)
}
