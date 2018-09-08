# Eagle Tunnel Go

利用Golang重新实现了ET的代理协议，从而拥有以下新特性（打钩代表已实现）：

- [x] 单文件部署
- [x] 高并发能力
- [x] 低内存占用
- [x] 更简单的配置过程
- [x] 全指令集平台支持
- [ ] 用户友好的Web UI

以下传统特性也是受支持的：

- [x] 自建ET协议
- [x] HTTP(S) 协议
- [x] SOCKS5 协议
- [x] 账户控制
- [x] 两级缓存加速
- [x] 智能直连
- [x] 共享代理到内网
- [x] Linux/Windows/Mac 三平台支持

可配合额外的工具或脚本实现：

- [ ] systemd服务

## 下载

[最新发布](https://github.com/eaglexiang/eagle.tunnel.go/releases)

## 简单服务示例

### 服务端

在可执行程序所在目录建立配置文件`eagle-tunnel.conf`，然后启动程序即可。

配置文件示例：

> et = on

### 客户端

同样在可执行程序所在目录建立配置文件`eagle-tunnel.conf`，然后启动程序即可。

配置文件示例：

> relayer = *.*.*.* # 服务端的IP  
> http = on  
> socks = on

**注意** ：服务端必须打开TCP-8080端口的防火墙，系统或浏览器的代理地址设置为`客户端IP:8080`即可。