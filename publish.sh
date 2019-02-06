#!/bin/bash

RootPath=$(pwd)
ScriptPath="${RootPath}/scripts"
binPath="./zip"

if [ $# -gt 0 ]; then
    case $1 in
        "clean")
        echo "begin to clean tmp files for publish"
        if [ -d ${binPath} ]; then
            rm -rf ${binPath}
        else
            echo "nothing to clean"
        fi
        echo "publish clean done"
        ;;
        "all")
        source ${ScriptPath}/publishzip.sh linux amd64
        source ${ScriptPath}/publishzip.sh linux 386
        source ${ScriptPath}/publishzip.sh linux arm
        source ${ScriptPath}/publishzip.sh darwin amd64
        source ${ScriptPath}/publishzip.sh darwin 386
        source ${ScriptPath}/publishzip.sh windows amd64
        source ${ScriptPath}/publishzip.sh windows 386
        ;;
        *)
        source ${ScriptPath}/publishzip.sh $*
        ;;
    esac
else
    source ${ScriptPath}/publishzip.sh $*
fi