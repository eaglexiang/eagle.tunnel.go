package comm

import (
	"github.com/eaglexiang/go-tunnel"
)

// Connect2Remote 连接到远端
var Connect2Remote func(tunnel *tunnel.Tunnel) error

// SendQueryReq 发送查询请求
// 查询类请求的发射过程都是类似的
// 连接 - 发送请求 - 得到反馈 - 关闭连接
// 区别仅仅在请求命令的内容
var SendQueryReq func(req string) (string, error)
