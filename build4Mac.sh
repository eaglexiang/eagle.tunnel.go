source ./scripts/env.sh

echo "begin to build golang code"
mkdir -p ${BinPath}
cd ${SrcPath} && CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o ${BinPath}/et.go.linux main.go
echo "done"

source ${ScriptPath}/copy-http.sh
source ${ScriptPath}/copy-whitelist.sh