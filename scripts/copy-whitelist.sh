#! /bin/bash

echo "copy whitelist files..."

mkdir -p ${ConfDesPath}
cp ${WhiteListSrcPath}/list.txt ${ConfDesPath}/whitelist_domain.txt

echo "done"