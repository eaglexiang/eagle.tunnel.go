source ./scripts/env.sh

if [ $# -gt 0 ]; then
    des=$1
else
    des="linux"
fi
echo "des: ${des}"

mkdir -p ${BinPath}

case ${des} in
    "linux")
    source ${ScriptPath}/build4Linux.sh
    ;;
    "windows")
    source ${ScriptPath}/build4Windows.sh
    ;;
    "mac")
    source ${ScriptPath}/build4Mac.sh
    ;;
    "clean")
    rm -rf ${BinPath}
    exit
    ;;
    *)
    exit
    ;;
esac

source ${ScriptPath}/copy-http.sh
source ${ScriptPath}/copy-whitelist.sh
source ${ScriptPath}/copy-hosts.sh
source ${ScriptPath}/copy-services.sh