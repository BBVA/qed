#!/bin/bash

mkdir -p ./data

go build -o ./data/storage ../../../../tests/gossip/test_service.go
