#!/bin/bash

mkdir -p ./data

export GOOS=linux
export GOARCH=amd64
go build -o ./data/storage ../../../../tests/gossip/test_service.go
