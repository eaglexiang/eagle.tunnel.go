#!/bin/bash

###
# @Author: EagleXiang
# @LastEditors: EagleXiang
# @Email: eagle.xiang@outlook.com
# @Github: https://github.com/eaglexiang
# @Date: 2019-08-24 11:16:00
# @LastEditTime: 2019-08-24 11:43:25
###

# $1 为文件名
copy_config() {
    if [ -f $1 ]; then
        echo "find $1, new file will be named $1_new"
        cp ./config/$1 $root/etc/eagle-tunnel.d/$1_new
    else
        cp ./config/$1 $root/etc/eagle-tunnel.d/$1
    fi
}

# main

# user check
if [ $UID -ne 0 ]; then
    echo "this script requires superuser privileges."
    exit 1
fi

# copy files

# lib
echo "lib installing..."
mkdir -p $root/usr/eagle-tunnel
\cp -f ./et $root/usr/eagle-tunnel/

# etc
echo "etc installing..."
mkdir -p $root/etc/eagle-tunnel.d
\cp -rf ./clearDomains $root/etc/eagle-tunnel.d/
copy_config client.conf
copy_config server.conf
copy_config users.list
\cp -rf ./hosts/* $root/etc/eagle-tunnel.d/

# bin
echo "bin installing..."
ln -s /usr/lib/eagle-tunnel/et $root/bin/et

# systemd
echo "systemd units installing..."
\cp -f ./nix/systemd/* $root/usr/lib/systemd/system
systemctl daemon-reload

echo "done"
