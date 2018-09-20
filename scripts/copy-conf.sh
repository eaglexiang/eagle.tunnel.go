#!/bin/bash

echo "begin to copy config files..."

mkdir -p ${ConfDesPath}
cp -n ${ConfSrcPath}/* ${ConfDesPath}

echo "done"