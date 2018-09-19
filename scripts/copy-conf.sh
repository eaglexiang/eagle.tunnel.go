#!/bin/bash

echo "begin to copy service files..."

mkdir -p ${ConfDesPath}
cp -n ${ConfSrcPath}/* ${ConfDesPath}

echo "done"