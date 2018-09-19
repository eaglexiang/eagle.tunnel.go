#!/bin/bash

echo "begin to copy service files..."

mkdir -p ${ServiceDesPath}
cp -f ${ServiceSrcPath}/* ${ServiceDesPath}

echo "done"