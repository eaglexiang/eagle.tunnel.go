#! /bin/bash

echo "copy whitelist files..."

mkdir -p ${ConfDesPath}
cp ${ClearListSrcPath}/proxylist.txt ${ConfDesPath}/proxylist_domain.txt
cp ${ClearListSrcPath}/directlist.txt ${ConfDesPath}/directlist_domain.txt


echo "done"