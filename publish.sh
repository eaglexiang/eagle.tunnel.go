#!/bin/bash

echo "start to zip publish"

./build.sh $*

tmpPath="./Eagle_Tunnel_Go"

mkdir -p ${tmpPath}

echo "copy publish..."
cp -r ./publish ${tmpPath}
cp ./install.sh ${tmpPath}
cp ./uninstall.sh ${tmpPath}

echo "compressing..."
zip -r et.go.zip ${tmpPath}

echo "clear tmp files"
rm -rf ${tmpPath}
./build.sh clean

echo "done"