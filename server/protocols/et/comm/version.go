/*
 * @Author: EagleXiang
 * @LastEditors: EagleXiang
 * @Email: eagle.xiang@outlook.com
 * @Github: https://github.com/eaglexiang
 * @Date: 2019-08-28 19:47:35
 * @LastEditTime: 2019-08-28 19:48:34
 */

package comm

import "github.com/eaglexiang/go/version"

// ProtocolVersion 作为Sender使用的协议版本号
var ProtocolVersion, _ = version.CreateVersion("1.6")

// ProtocolCompatibleVersion 作为Handler可兼容的最低协议版本号
var ProtocolCompatibleVersion, _ = version.CreateVersion("1.3")
