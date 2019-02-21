#! /bin/bash

echo "get parameters..."

RootPath=$(pwd)
PublishPath="${RootPath}/publish"
SrcPath="${RootPath}/src"
ScriptPath="${RootPath}/scripts"
EtSrcPath="${SrcPath}/service"
HttpSrcPath="${EtSrcPath}/http"
WhiteListSrcPath="${SrcPath}/whitelistdomains"
HostsSrcPath="${SrcPath}/hosts"
ServiceSrcPath="${SrcPath}/services"
ConfSrcPath="${SrcPath}/config"
SrcMainPath="${SrcPath}/main"

HttpDesPath="${PublishPath}/http"
ServiceDesPath="${PublishPath}/services"
ConfDesPath="${PublishPath}/config"
HostsDesPath="${ConfDesPath}/hosts"

export GOPATH=$GOPATH:${RootPath}

echo "done"