# Eagle Tunnel Go

利用Golang重新实现了ET的代理协议，从而拥有以下新特性（打钩代表已实现）：

- [x] 一键部署
- [x] 高并发能力
- [x] 低资源占用
- [x] 更简单的配置过程
- [x] 全指令集平台支持
- [ ] 用户友好的Web UI

以下传统特性也是受支持的：

- [x] 自建ET协议
- [x] HTTP(S) 协议
- [x] SOCKS5 协议
- [x] 账户控制
- [x] 账户限速
- [x] 固定资源缓存加速
- [x] 智能直连
- [x] hosts
- [x] 共享代理到内网
- [x] 多操作系统平台支持
- [x] Systemd Unit 服务
- [ ] 负载均衡

## 下载

[最新发布](https://github.com/eaglexiang/eagle.tunnel.go/releases)

## 简单服务示例

### 服务端

进入程序目录，运行以下程序：

Linux

```shell
./et.go.linux ./config/server.conf
```

Windows

```powershell
.\et.go.exe .\config\server.conf
```

Mac

```shell
./et.go.mac ./config/server.conf
```

### 客户端

进入程序目录，修改配置文件`config/client.conf`

配置文件示例：

> listen = 0.0.0.0  
> relayer = # 服务端的IP  
> http = on  
> socks = on

将`relayer`修改为服务端IP，然后运行

Linux

```shell
./et.go.linux
```

Windows

```powershell
.\et.go.exe
```

Mac

```shell
./et.go.mac
```

**注意** ：服务端必须打开TCP-8080端口的防火墙，系统或浏览器的代理地址设置为`127.0.0.1:8080`即可使用。