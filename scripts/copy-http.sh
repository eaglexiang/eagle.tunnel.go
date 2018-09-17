#! /bin/bash

echo "copy http files..."

echo "from ${HttpSrcPath} to ${BinPath}"

mkdir -p ${BinPath}/http
cp -r ${HttpSrcPath}/* ${BinPath}/http

echo "done"