#!/bin/bash

echo "start to zip publish"

./build.sh $*

tmpPath="./Eagle_Tunnel_Go"
binPath="./zip"

mkdir -p ${tmpPath}

echo "copy publish..."
cp -r ./publish ${tmpPath}
cp ./install.sh ${tmpPath}
cp ./uninstall.sh ${tmpPath}

echo "compressing..."
if [ $# -gt 0 ]; then
    os=$1
else
    os="linux"
fi

if [ $# -gt 1 ];then
    arch=$2
else
    arch="amd64"
fi

bin="et.go.${os}.${arch}.zip"
zip -r ${binPath}/${bin} ${tmpPath}

echo "clear tmp files"
rm -rf ${tmpPath}
./build.sh clean

echo "done"