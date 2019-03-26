#! /bin/bash

echo "copy whitelist files..."

ProxyDomainsDesDir=${ConfDesPath}/proxylists
DirectDomainsDesDir=${ConfDesPath}/directlists

mkdir -p ${ProxyDomainsDesDir}
mkdir -p ${DirectDomainsDesDir}

cp ${ClearListSrcPath}/proxylist.txt ${ProxyDomainsDesDir}
cp ${ClearListSrcPath}/directlist.txt ${DirectDomainsDesDir}


echo "done"