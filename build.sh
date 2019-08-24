#!/bin/bash
###
# @Author: EagleXiang
# @LastEditors: EagleXiang
# @Email: eagle.xiang@outlook.com
# @Github: https://github.com/eaglexiang
# @Date: 2019-08-24 12:52:28
# @LastEditTime: 2019-08-24 22:29:35
###

if [ "$1" == "clean" ]; then
    rm ./et.go.*
    exit 0
fi

echo "============================================================"
echo "PREPARE"
echo "============================================================"

defaultOS=linux
defaultArch=amd64

# $1:os
if [ "$1" ]; then
    os=$1
else
    os=$defaultOS
fi
echo -e "OS:\t$os"

if [ $os == "mac" ]; then
    os="darwin"
fi

# $2:arch
if [ "$2" ]; then
    arch=$2
else
    arch=$defaultArch
fi
echo -e "ARCH:\t$arch"

# suffix
if [ $os == "windows" ]; then
    suffix="exe"
else
    suffix=$os
fi

echo -e "SUFFIX:\t$suffix"

# build
file="et.go.$suffix"
echo "============================================================"
echo -e "BUILD:\tEagle Tunnel for $os on $arch ..."
echo -e "BIN:\t$file"
echo "============================================================"

CGO_ENABLED=0 GOOS=$os GOARCH=$arch go build -o ./$file ./main.go

if [ -f $file ]; then
    echo "COMPILE DONE"
    echo "============================================================"
else
    echo "COMPILE ERROR"
    echo "============================================================"
    exit 1
fi
