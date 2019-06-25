/*
 * @Author: EagleXiang
 * @Github: https://github.com/eaglexiang
 * @Date: 2018-12-27 08:37:36
 * @LastEditors: EagleXiang
 * @LastEditTime: 2019-06-20 21:42:44
 */

package config

import (
	settings "github.com/eaglexiang/go-settings"
)

func init() {
	// 绑定不同的参数名
	settings.Bind("relayer", "relay")
	settings.Bind("channel", "data-key")
	// 设定参数默认值
	settings.SetDefault("timeout", "30")
	settings.SetDefault("location", "1;CN;CHN;China")
	settings.SetDefault("ip-type", "46")
	settings.SetDefault("data-key", "34")
	settings.SetDefault("head", "eagle_tunnel")
	settings.SetDefault("proxy-status", "smart")
	settings.SetDefault("user", "null:null")
	settings.SetDefault("user-check", "off")
	settings.SetDefault("listen", "0.0.0.0")
	settings.SetDefault("relay", "127.0.0.1")
	settings.SetDefault("http", "off")
	settings.SetDefault("socks", "off")
	settings.SetDefault("et", "off")
	settings.SetDefault("debug", "warning")
	settings.SetDefault("cipher", "simple")
	settings.SetDefault("maxclients", "0")
	settings.SetDefault("buffer.size", "1000")
	settings.SetDefault("dynamic-ipe", "off")
}
