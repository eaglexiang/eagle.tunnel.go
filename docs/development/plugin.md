<!--
 * @Author: EagleXiang
 * @LastEditors: EagleXiang
 * @Email: eagle.xiang@outlook.com
 * @Github: https://github.com/eaglexiang
 * @Date: 2019-02-07 01:11:26
 * @LastEditTime: 2019-02-07 01:32:57
 -->
# 插件接口

ET提供了简易的插件接口，凡是符合ET插件格式的动态链接库，在被放入`mod-dir`参数指定的文件夹后，可被程序自动加载。使ET支持更多它本不支持的协议。

由于Golang的插件模式暂时只支持Linux，因此ET的插件接口也只对Linux生效。

## 工作流程

ET整体的工作流程可分为五步：

1. 获取握手消息，判断协议类型，分发业务
2. 业务处理者提取目标地址（IP:Port），注册可能需要的委托
3. 由业务发送者连接该目标地址
4. 完成由业务处理者在第2步注册的委托行为
5. 开始数据的透明流动

## 插件的格式

每个插件必须实现`业务处理者`这个接口，其接口定义在[handler.go
](https://github.com/eaglexiang/eagle.tunnel.go/blob/master/src/core/core/handler.go)中：

```go
package etcore

import mynet "github.com/eaglexiang/go-net"

// Handler 请求处理者
type Handler interface {
	Handle(e *mynet.Arg) error  // 处理业务
	Match(firstMsg []byte) bool // 判断业务请求是否符合该handler
	Name() string               // Handler的名字
}

// AllHandlers 注册handler的标准位置
var AllHandlers = make(map[string]Handler)
```

需要注意的是：

- `Handle`方法中需要完成对目标地址（IP:Port）的提取，将结果存放于`e`中。
- 需要在`init`中定义新协议的实例，并将其添加进AllHandlers中。
- 委托行为请以闭包的方式存放在`e`中

参数`e`的类型定义在[arg.go](https://github.com/eaglexiang/go-net/blob/master/arg.go)中：

```go
package net

import tunnel "github.com/eaglexiang/go-tunnel"

// Arg 网络业务会用到的参数集
type Arg struct {
	Msg       []byte             // 消息
	Host      string             // 主机地址:端口
	Tunnel    *tunnel.Tunnel     // 数据隧道
	TheType   int                // 业务类型
	Delegates []func() bool      // 委托队列
}
```