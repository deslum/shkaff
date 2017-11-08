#!/bin/bash
export GOPATH=$PWD
export GOARCH=amd64
export GOOS=linux
export GOBIN=$(go env GOPATH)/bin
go install "src/shkaff/shkaff.go"
