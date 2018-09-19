#!/bin/bash

echo "begin to copy hosts files..."

mkdir -p ${HostsDesPath}
cp -f ${HostsSrcPath}/* ${HostsDesPath}

echo "done"