#!/bin/bash

echo "begin to copy service files..."

mkdir -p ${BinPath}/services
cp ${ServiceSrcPath}/* ${BinPath}/services

echo "done"