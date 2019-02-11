#!/bin/bash

export GOOS=linux
export GOARCH=amd64

mkdir -p ./data

go build -o ./data/riot ../../../../tests/riot.go
