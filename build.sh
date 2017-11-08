#!/bin/bash
export GOPATH=$PWD
export GOARCH=amd64
export GOOS=linux
export GOBIN=$(go env GOPATH)/bin
echo $(go env GOBIN)
go install "src/shkaff/shkaff.go"
