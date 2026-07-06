#!/bin/sh
cd /src
go mod tidy
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o server .
echo "Build exit: $?"
