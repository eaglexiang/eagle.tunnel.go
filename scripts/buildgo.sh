echo "begin to build golang code"

if [ $# -gt 0 ]; then
    os=$1
else
    os="linux"
fi

if [ $# -gt 1 ];then
    arch=$2
else
    arch="amd64"
fi

case ${os} in
    "linux")
    bin="et.go.${arch}.linux"
    ;;
    "mac")
    bin="et.go.${arch}.mac"
    os="darwin"
    ;;
    "windows")
    bin="et.go.${arch}.exe"
    ;;
    *)
    bin="et.go"
    ;;
esac

echo "OS: ${os}"
echo "Arch: ${arch}"
echo "BinFile: ${bin}"

CGO_ENABLED=0 GOOS=${os} GOARCH=${arch} go build -o ${PublishPath}/${bin} ${SrcPath}/main.go

echo "done"