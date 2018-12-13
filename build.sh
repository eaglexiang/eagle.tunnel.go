source ./scripts/env_build.sh

mkdir -p ${PublishPath}

if [ $# -gt 0 ]; then
    case $1 in
        "clean")
        echo "begin to clean tmp files for build"
        if [ -d ${PublishPath} ]; then
            rm -rf ${PublishPath}
        else
            echo "nothing to clean"
        fi
        echo "build clean done"
        exit
        ;;
        *)
        source ${ScriptPath}/buildgo.sh $*
        ;;
    esac
else
    source ${ScriptPath}/buildgo.sh $*
fi

source ${ScriptPath}/copy-http.sh
source ${ScriptPath}/copy-whitelist.sh
source ${ScriptPath}/copy-hosts.sh
source ${ScriptPath}/copy-services.sh
source ${ScriptPath}/copy-conf.sh
cp -n ./LICENSE ${PublishPath}