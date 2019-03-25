#! /bin/bash

echo "copy whitelist files..."

mkdir -p ${ConfDesPath}

cp ${ClearListSrcPath}/proxylist.txt ${ConfDesPath}
cp ${ClearListSrcPath}/directlist.txt ${ConfDesPath}


echo "done"