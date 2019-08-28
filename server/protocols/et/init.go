/*
 * @Author: EagleXiang
 * @LastEditors: EagleXiang
 * @Email: eagle.xiang@outlook.com
 * @Github: https://github.com/eaglexiang
 * @Date: 2019-08-28 20:46:40
 * @LastEditTime: 2019-08-28 20:48:19
 */

package et

import "github.com/eaglexiang/go/settings"

func init() {
	settings.Bind("relayer", "relay")

	settings.SetDefault("relay", "127.0.0.1")
	settings.SetDefault("location", "1;CN;CHN;China")
	settings.SetDefault("ip-type", "46")
}
