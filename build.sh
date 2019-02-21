source ./scripts/env_build.sh

# 清理旧的临时文件
echo "begin to clean tmp files for build"
if [ -d ${PublishPath} ]; then
    rm -rf ${PublishPath}
else
    echo "nothing to clean"
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

source ${ScriptPath}/copy-whitelist.sh
source ${ScriptPath}/copy-hosts.sh
source ${ScriptPath}/copy-services.sh
source ${ScriptPath}/copy-conf.sh
cp -n ./LICENSE ${PublishPath}