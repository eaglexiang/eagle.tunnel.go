#!/bin/bash
###
# @Author: EagleXiang
# @LastEditors: EagleXiang
# @Email: eagle.xiang@outlook.com
# @Github: https://github.com/eaglexiang
# @Date: 2019-08-24 12:52:28
# @LastEditTime: 2019-08-24 13:20:48
###

defaultOS=linux
defaultArch=amd64

# $1:os
if [ $1 ]; then
    os=$1
else
    echo "using default OS $defaultOS"
    os=$defaultOS
fi

if [ $os = "mac" ]; then
    echo "macOS equals darwin kernel"
    os="darwin"
fi

# $2:arch
if [ $2 ]; then
    arch=$2
else
    echo "using default Arch $defaultArch"
    arch=$defaultArch
fi

# suffix
if [ $os = "windows" ]; then
    suffix="exe"
else
    suffix=$os
fi

echo "suffix for operating system $os is $suffix"

# build
echo "=============================================="
echo -e "build:\tEagle Tunnel for $os on $arch ..."

CGO_ENABLED=0 GOOS=$os GOARCH=$arch go build -o ./et.go.$suffix ./main.go
echo -e "got:\tet.go.$suffix"
echo "=============================================="

echo "done"
