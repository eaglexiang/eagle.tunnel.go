#! /bin/bash

echo "copy whitelist files..."

mkdir -p ${BinPath}/config
cp ${WhiteListSrcPath}/list.txt ${BinPath}/config/whitelist_domain.txt

echo "done"