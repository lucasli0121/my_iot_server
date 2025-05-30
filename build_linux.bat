SET CGO_ENABLED=0
SET GOOS=linux
rem ****支持X86******
SET GOARCH=amd64
rem ******************
rem ****支持arm架构*******
rem SET GOARCH=arm
rem SET GOARM=7
rem **********************
SET GIN_MODE=release

go build -o my_iot_server_%1 main.go