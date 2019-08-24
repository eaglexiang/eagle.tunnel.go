###
# @Author: EagleXiang
# @LastEditors: EagleXiang
# @Email: eagle.xiang@outlook.com
# @Github: https://github.com/eaglexiang
# @Date: 2019-08-24 10:42:03
# @LastEditTime: 2019-08-24 10:42:03
###

#!/bin/bash

source ./scripts/env_build.sh

# 清理旧的临时文件
if [ -d ${PublishPath} ]; then
    echo "begin to clean tmp files for build"
    rm -rf ${PublishPath}
fi

if [ $# -gt 0 ]; then
    case $1 in
    "clean")
        echo "build clean done"
        exit
        ;;
    *)
        mkdir -p ${PublishPath}
        source ${ScriptPath}/buildgo.sh $*
        ;;
    esac
else
    mkdir -p ${PublishPath}
    source ${ScriptPath}/buildgo.sh $*
fi

source ${ScriptPath}/copy-clearlist.sh
source ${ScriptPath}/copy-hosts.sh
source ${ScriptPath}/copy-services.sh
source ${ScriptPath}/copy-conf.sh
cp -n ./LICENSE ${PublishPath}
