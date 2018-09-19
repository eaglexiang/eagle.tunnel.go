echo "begin to build golang code for Linux"

CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ${PublishPath}/et.go.linux ${SrcPath}/main.go

echo "done"