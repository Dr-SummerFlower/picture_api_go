#!/d/git/usr/bin/zsh

CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o ./build/picture_api-windows-amd64.exe ./go_server/main.go
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./build/picture_api-linux-amd64 ./go_server/main.go
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o ./build/picture_api-linux-arm64 ./go_server/main.go

