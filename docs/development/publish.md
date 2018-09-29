# 编译和打包

项目自带编译和打包脚本，可以按下述规则进行调用：

```shell
./build.sh <os> <arch>
# 编译程序
./publish.sh <os> <arch>
# 编译程序并对其进行打包
```

示例：

```shell
./build.sh linux amd64
# 编译为amd64指令集的linux程序
./publish.sh windows 386
# 编译为386指令集的Windows程序，并将其进行打包
```

os参数指代编译目标平台的操作系统，arch参数指代编译目标平台的指令集。受支持目标平台参见[go/src/go/build/syslist.go](https://github.com/golang/go/blob/master/src/go/build/syslist.go)，这里也有一份组合总结：[A list of GOOS/GOARCH supported by go out of the box](https://gist.github.com/asukakenji/f15ba7e588ac42795f421b48b8aede63#a-list-of-goosgoarch-supported-by-go-out-of-the-box)