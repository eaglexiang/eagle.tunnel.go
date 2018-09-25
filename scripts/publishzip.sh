#!/bin/bash

echo "start to zip publish"

source ./scripts/env_build.sh

if [ -d ${PublishPath}];then
    rm -rf PublishPath
fi

./build.sh $*

tmpPath="./Eagle_Tunnel_Go"

mkdir -p ${tmpPath}

echo "copy publish..."
cp -r ${PublishPath} ${tmpPath}
cp ./install.sh ${tmpPath}
cp ./uninstall.sh ${tmpPath}
echo "copy done"

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
echo "Bin File: ${bin}"
mkdir -p ${binPath}
zip -r ${binPath}/${bin} ${tmpPath}
echo "compress done"

echo "clear tmp files"
rm -rf ${tmpPath}
./build.sh clean
echo "tmp clear done"

echo "zip publish done"