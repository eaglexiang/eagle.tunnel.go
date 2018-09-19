echo "begin to build golang code for Linux"

go build -o ${BinPath}/et.go.linux ${SrcPath}/main.go

echo "done"