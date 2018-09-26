#!/bin/bash

source ./scripts/env_install.sh

echo "installing libs"
mkdir -p ${LibDesPath}
cp -f ${PublishPath}/et.go.linux ${LibDesPath}
cp -f ${PublishPath}/run.sh ${LibDesPath}
ln -sf ${LibDesPath}/run.sh /usr/bin/et

echo "installing systemd units"
mkdir -p ${ServiceDesPath}
cp -f ${ServiceSrcPath}/* ${ServiceDesPath}

echo "installing config files"
mkdir -p ${ConfDesPath}
if [ -f ${ConfDesPath}/client.conf ]; then
    echo "found old client.conf, new template will be named client.conf_new"
    cp -f ${ConfSrcPath}/client.conf ${ConfDesPath}/client.conf_new
else
    cp ${ConfSrcPath}/client.conf ${ConfDesPath}
fi
if [ -f ${ConfDesPath}/server.conf ]; then
    echo "found old server.conf, new template will be named server.conf_new"
    cp -f ${ConfSrcPath}/server.conf ${ConfDesPath}/server.conf_new
else
    cp ${ConfSrcPath}/server.conf ${ConfDesPath}
fi
cp -f ${ConfSrcPath}/whitelist_domain.txt ${ConfDesPath}
cp -rf ${ConfSrcPath}/hosts ${ConfDesPath}

./scripts/after-install.sh