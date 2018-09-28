# Linux用户 使用指南

ET分为服务端和客户端两部分，服务端通过ET协议提供隧道，而客户端提供对HTTP代理和SOCKS代理的支持，并将其流量转换为ET协议流量。因此部署ET会分为服务端部署和客户端部署两部分。

## 服务端

### 下载

可从[Release](https://github.com/eaglexiang/eagle.tunnel.go/releases)下载你需要的软件包。

对于Linux_x64用户，也可尝试执行下面的指令

```shell
curl https://raw.githubusercontent.com/eaglexiang/eagle.tunnel.go/master/docs/latestReleases/linux64.txt | xargs wget
```

### 安装

```shell
# 具体的文件名和目录名根据你实际下载的文件进行替换
tar -zxvf ./et.go.linux.amd64.tar.gz # 解压
cd ./Eagle_Tunnel_Go # 进入程序目录
sudo ./install.sh # 安装
```

### 启动

最简单的启动方式是执行指令`et server`，不过如果你使用的是支持Systemd服务的Linux发行版，建议执行以下指令启动并开机启动后台服务：

```shell
sudo systemctl enable eagle-tunnel-server
sudo systemctl start eagle-tunnel-server
```

完成后可通过以下指令检查其执行状态：

```shell
sudo systemctl status eagle-tunnel-server
```

## 客户端

### 同样是下载

可从[Release](https://github.com/eaglexiang/eagle.tunnel.go/releases)下载你需要的软件包。

对于Linux_x64用户，也可尝试执行下面的指令

```shell
curl https://raw.githubusercontent.com/eaglexiang/eagle.tunnel.go/master/docs/latestReleases/linux64.txt | xargs wget
```

### 同样是安装

```shell
# 具体的文件名和目录名根据你实际下载的文件进行替换
tar -zxvf ./et.go.linux.amd64.tar.gz # 解压
cd ./Eagle_Tunnel_Go # 进入解压出的程序目录
sudo ./install.sh # 安装
```

### 查看服务端的IP

我们现在需要查询服务端的IP，如果是VPS，通常可以在网页端的控制面板里找到它。如果不能，则可在服务端所在系统中执行以下两条指令的其中之一：

```shell
ifconfig
ip addr show
```

这两条命令都会打印一长串输出，而我们需要的是其中明叫`inet`的部分。举个例子，我这里有一份假设的输出实例：

```shell
lo: flags=73<UP,LOOPBACK,RUNNING>  mtu 65536
        inet 127.0.0.1  netmask 255.0.0.0
        inet6 ::1  prefixlen 128  scopeid 0x10<host>
        loop  txqueuelen 1000  (Local Loopback)
        RX packets 6664426  bytes 4371007885 (4.0 GiB)
        RX errors 0  dropped 0  overruns 0  frame 0
        TX packets 6664426  bytes 4371007885 (4.0 GiB)
        TX errors 0  dropped 0 overruns 0  carrier 0  collisions 0

wlp8s0: flags=4163<UP,BROADCAST,RUNNING,MULTICAST>  mtu 1500
        inet 41.22.248.101  netmask 255.255.254.0  broadcast 41.22.248.255
        inet6 fe80::e461:5f71:76ff:3007  prefixlen 64  scopeid 0x20<link>
        ether 84:a6:c8:3f:aa:98  txqueuelen 1000  (Ethernet)
        RX packets 5561119  bytes 6401865599 (5.9 GiB)
        RX errors 0  dropped 60836  overruns 0  frame 0
        TX packets 3495412  bytes 1297836856 (1.2 GiB)
        TX errors 0  dropped 0 overruns 0  carrier 0  collisions 0
```

在其中我们能拿到两个`inet`：`127.0.0.1`和`41.22.248.101`，而我们需要的则是`41.22.248.101`。简单的判断方法为，除开`127.0.0.1`的另一个就是了。

### 配置客户端

由于客户端并不知道你的服务端运行在哪里，所以你至少要为客户端配置一个参数（`relayer`)告诉它这一点。

```shell
# 使用任何你喜欢或习惯的文本编辑器打开/etc/eagle-tunnel.d/client.conf文件
sudo nano /etc/eagle-tunnel.d/client.conf
```

一个典型的客户端配置模板为：

> listen=0.0.0.0  
> relayer=  
> http=on  
> socks=on

我们需要将服务端的IP填在`relayer`中，假设其IP为上文的`41.22.248.101`，则配置过后的文件应该为：

> listen=0.0.0.0  
> relayer=41.22.248.101  
> http=on  
> socks=on

### 稍有不同的启动

> 注意和上文服务端的指令是稍有不同的

最简单的启动方式是执行指令`et client`，不过如果你使用的是支持Systemd服务的Linux发行版，建议执行以下指令启动并开机启动后台服务：

```shell
sudo systemctl enable eagle-tunnel-client
sudo systemctl start eagle-tunnel-client
```

完成后可通过以下指令检查其执行状态：

```shell
sudo systemctl status eagle-tunnel-client
```

## 验证

配置完成后也许有验证服务是否正在运行的需求，可在客户端处执行以下指令：

```shell
et ask local ping
```

如果服务正常，应该会返回类似以下结果：

> ping to x.x.x.x:8080 time=99ms

`time`指的是你的客户端到服务端的时延

## 更多参数

ET支持很多参数，帮助你实现更高级的体验，请参照[ET 配置](/docs/config.md)一文。