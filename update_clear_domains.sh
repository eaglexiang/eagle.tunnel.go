#!/bin/bash
###
# @Author: EagleXiang
# @LastEditors: EagleXiang
# @Email: eagle.xiang@outlook.com
# @Github: https://github.com/eaglexiang
# @Date: 2019-08-28 21:20:55
# @LastEditTime: 2019-08-28 21:41:06
###

echo "============================================================"
echo "UPDATE CLEAR DOMAINS"
echo "============================================================"
echo "PROXY LIST"
echo "============================================================"
wget "https://raw.githubusercontent.com/remmina/proxy-list/master/proxylist.txt" -O ./config/proxylists/proxylist.txt
echo "============================================================"
echo "DIRECT LIST"
echo "============================================================"
wget "https://raw.githubusercontent.com/remmina/proxy-list/master/directlist.txt" -O ./config/directlists/directlist.txt
echo "============================================================"
echo "DONE"
echo "============================================================"
