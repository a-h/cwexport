#!/bin/sh

GOOS="linux" GOARCH="amd64" go build -o cdk/lambda/processor/lambda -ldflags "-s -w" cdk/lambda/processor/main.go

go build