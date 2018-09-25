source ./scripts/env_build.sh

mkdir -p ${PublishPath}

if [ $# -gt 0 ]; then
    case $1 in
        "clean")
        rm -rf ${PublishPath}
        exit
        ;;
        *)
        source ${ScriptPath}/buildgo.sh $*
        exit
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