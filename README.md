<!--
 * @Author: EagleXiang
 * @LastEditors: EagleXiang
 * @Email: eagle.xiang@outlook.com
 * @Github: https://github.com/eaglexiang
 * @Date: 2019-01-29 15:00:05
 * @LastEditTime: 2019-02-21 17:39:02
 -->
# Eagle Tunnel Go

![build](https://travis-ci.org/eaglexiang/eagle.tunnel.go.svg?branch=master) ![platforms](https://img.shields.io/badge/platform-Linux|Windows|macOS-lightgrey.svg) ![](https://img.shields.io/badge/language-go-blue.svg) ![license](https://img.shields.io/badge/license-MIT-green.svg)

## 什么是ET

一个轻量且简单易用的代理工具

## 它有什么优势？

### 稳定

ET采用自建协议——而不是对传统协议的重新实现，因此针对传统协议的干扰对它是不生效的。

ET采用短连接——每次请求开启独占的TCP连接，请求结束即关闭连接。常见的针对长连接的干扰也是不生效的。

ET支持自定义协议头——每个用户都可以有独立的协议特征，这使它具有一定抗嗅探能力。

### 智能

智能模式支持国内国外智能切换，无需频繁开关代理。

### 简单

配置过程能简则简，最少一分钟搭建服务，无需过多学习。

## 简易服务搭建

服务端

```shell
et --listen 0.0.0.0 --et on
```

客户端

```shell
et --listen 0.0.0.0 --http on --relayer [服务端IP]
```

此时本地HTTP代理服务地址为`127.0.0.1:8080`。

> 如果服务端开启了防火墙服务，则必须开启`8080/tcp`端口

## 下载

[最新发布](https://github.com/eaglexiang/eagle.tunnel.go/releases/latest)

## 详细

参照[指南](./docs/guide.md)一文。