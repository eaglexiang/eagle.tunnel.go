#! /bin/bash

echo "copy http files..."

mkdir -p ${HttpDesPath}
cp -rf ${HttpSrcPath}/* ${HttpDesPath}

echo "done"