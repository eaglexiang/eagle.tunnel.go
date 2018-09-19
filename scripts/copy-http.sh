#! /bin/bash

echo "copy http files..."

echo "from ${HttpSrcPath} to ${HttpDesPath}"

mkdir -p ${HttpDesPath}
cp -rf ${HttpSrcPath}/* ${HttpDesPath}

echo "done"