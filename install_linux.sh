#!/bin/bash

###
# @Author: EagleXiang
# @LastEditors: EagleXiang
# @Email: eagle.xiang@outlook.com
# @Github: https://github.com/eaglexiang
# @Date: 2019-08-24 11:16:00
# @LastEditTime: 2019-08-24 22:19:48
###

# $1 为文件名
copy_config() {
    if [ -f "$1" ]; then
        echo "find $1, new file will be named $1_new"
        cp "./config/$1" "$root/etc/eagle-tunnel.d/$1_new"
    else
        cp "./config/$1" "$root/etc/eagle-tunnel.d/$1"
    fi
}

# main

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
\cp -rf ./hosts "$root/etc/eagle-tunnel.d/"
# proxylists
proxylists_dir="$root/etc/eagle-tunnel.d/proxylists"
mkdir -p "$proxylists_dir"
cp ./clearDomains/proxylist.txt "$proxylists_dir/"
# directlists
directlists_dir="$root/etc/eagle-tunnel.d/directlists"
mkdir -p "$directlists_dir"
cp ./clearDomains/directlist.txt "$directlists_dir/"

# bin
echo "bin installing..."
mkdir -p "$root/bin"
ln -sf "$root/usr/lib/eagle-tunnel/et.go.linux" "$root/bin/et"

# systemd
echo "systemd units installing..."
mkdir -p "$root/usr/lib/systemd/system"
\cp -f ./nix/systemd/* "$root/usr/lib/systemd/system"
if [ X"$root" == "X" ];then # root not specific
    systemctl daemon-reload
fi

echo "done"
