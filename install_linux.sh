#!/bin/bash

###
# @Author: EagleXiang
# @LastEditors: EagleXiang
# @Email: eagle.xiang@outlook.com
# @Github: https://github.com/eaglexiang
# @Date: 2019-08-24 11:16:00
# @LastEditTime: 2019-08-28 21:33:55
###

# $1 为文件名
copy_config() {
    if [ -f "$root/etc/eagle-tunnel.d/$1" ]; then
        echo "find $1, new file will be named $1_new"
        cp "./config/$1" "$root/etc/eagle-tunnel.d/$1_new"
    else
        cp "./config/$1" "$root/etc/eagle-tunnel.d/$1"
    fi
}

# $1 为文件夹名
copy_config_dir() {
    \cp -rf "./config/$1" "$root/etc/eagle-tunnel.d/"
}

# main

if [ ! -f  "./et.go.linux" ];then
    echo "no et.go.linux"
    exit 1
fi

if [ "$1" == "test" ] && [ "$2" == "clean" ]
then
    rm -rf "$(pwd)/test"
    exit 0
elif [ "$1" == "test" ]
then
    root="$(pwd)/test"
fi


# user check
if [ X"$root" == "X" ];then # root not specific
    if [ $UID -ne 0 ]; then
        echo "this script requires superuser privileges."
        exit 1
    fi
fi

# copy files

# lib
echo "lib installing..."
mkdir -p "$root/usr/eagle-tunnel"
\cp -f ./et.go.linux "$root/usr/eagle-tunnel/"

# etc
echo "etc installing..."
mkdir -p "$root/etc/eagle-tunnel.d"
copy_config client.conf
copy_config server.conf
copy_config users.list

copy_config_dir hosts
copy_config_dir proxylists
copy_config_dir directlists

# bin
echo "bin installing..."
mkdir -p "$root/bin"
ln -sf "$root/usr/lib/eagle-tunnel/et.go.linux" "$root/bin/et"

# systemd
echo "systemd units installing..."
mkdir -p "$root/usr/lib/systemd/system"
\cp -f ./config/nix/systemd/* "$root/usr/lib/systemd/system"
if [ X"$root" == "X" ];then # root not specific
    systemctl daemon-reload
fi

echo "done"
