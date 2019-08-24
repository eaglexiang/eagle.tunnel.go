#!/bin/bash
###
# @Author: EagleXiang
# @LastEditors: EagleXiang
# @Email: eagle.xiang@outlook.com
# @Github: https://github.com/eaglexiang
# @Date: 2019-08-24 11:40:54
# @LastEditTime: 2019-08-24 11:43:59
###

# user check
if [ $UID -ne 0 ]; then
    echo "this script requires superuser privileges."
    exit 1
fi

echo "libs removing..."
rm -f $root/bin/et
rm -rf $root/etc/eagle-tunnel.d
rm -rf $root/usr/lib/eagle-tunnel

echo "systemd units removing..."
rm -f $root/usr/lib/systemd/system/eagle-tunnel-client.service
rm -f $root/usr/lib/systemd/system/eagle-tunnel-server.service
systemctl daemon-reload

echo "done"
