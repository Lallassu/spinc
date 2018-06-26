#!/bin/zsh
source ~/repository/crosscompile.bash
go fmt
VERSION=`git log | head -n 1 | cut  -f 2 -d ' '`
VERSION=${VERSION:0:6}
go build -ldflags "-s -w -X main.version=$VERSION" 
go-linux-amd64 build -ldflags "-s -w -X main.version=$VERSION" -o Linux/spinc
go build -ldflags "-s -w -X main.version=$VERSION" -o MacOS/spinc
