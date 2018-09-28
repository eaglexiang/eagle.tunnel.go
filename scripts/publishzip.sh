#!/bin/bash

echo "start to zip publish"

source ./scripts/env_build.sh

./build.sh $*

tmpPath="./Eagle_Tunnel_Go"

mkdir -p ${tmpPath}

echo "copy publish..."
cp -r ${PublishPath} ${tmpPath}
cp ./install.sh ${tmpPath}
cp ./uninstall.sh ${tmpPath}
mkdir -p ${tmpPath}/scripts
cp ${ScriptPath}/env_install.sh ${tmpPath}/scripts
cp ${ScriptPath}/after-install.sh ${tmpPath}/scripts
cp ${ScriptPath}/after-uninstall.sh ${tmpPath}/scripts
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

if [ "${os}x" == "windowsx" ]; then
    ZIP="zip -r"
    bin="et.go.${os}.${arch}.zip"
else
    ZIP="tar -zcvf"
    bin="et.go.${os}.${arch}.tar.gz"
fi
echo "Bin File: ${bin}"
mkdir -p ${binPath}
${ZIP} ${binPath}/${bin} ${tmpPath}
echo "compress done"

echo "clear tmp files"
rm -rf ${tmpPath}
./build.sh clean
echo "tmp clear done"

echo "zip publish done"