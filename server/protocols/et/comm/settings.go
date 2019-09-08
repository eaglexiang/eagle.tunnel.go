/*
 * @Author: EagleXiang
 * @LastEditors: EagleXiang
 * @Email: eagle.xiang@outlook.com
 * @Github: https://github.com/eaglexiang
 * @Date: 2019-09-08 11:36:08
 * @LastEditTime: 2019-09-08 12:03:03
 */

package comm

import "github.com/eaglexiang/go/settings"

const etSettingClass = "et"

var localSettings *settings.Settings

func init() {
	// 保持conf向前兼容
	settings.Bind("relayer", getClassKey("relay"))
	settings.Bind("relay", getClassKey("relay"))
	settings.Bind("location", getClassKey("location"))
	settings.Bind("ip-type", getClassKey("ip-type"))

	// 设定默认值
	settings.SetDefault("et.relay", "127.0.0.1")
	settings.SetDefault("et.location", "1;CN;CHN;China")
	settings.SetDefault("et.ip-type", "46")

	localSettings = settings.GetChild(etSettingClass)
}

func getClassKey(key string) (newKey string) {
	newKey = settings.JoinClassNameAndKey(etSettingClass, key)
	return
}

// GetSetting 从全局Settings中获取配置项
func GetSetting(key string) (value string) {
	value = localSettings.Get(key)
	return
}
