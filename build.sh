#!/bin/bash
export GOPATH=$PWD
export GOARCH=amd64
export GOOS=linux
export GOBIN=$(go env GOPATH)/bin
go get github.com/tools/godep
cp ./bin/godep /usr/bin
cd src/shkaff
godep restore
go install "shkaff.go"
