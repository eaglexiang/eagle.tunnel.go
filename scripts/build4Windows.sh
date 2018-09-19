echo "begin to build golang code for Windows"

CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o ${BinPath}/et.go.exe ${SrcPath}/main.go

echo "done"