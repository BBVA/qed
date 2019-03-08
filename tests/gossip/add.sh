#!/bin/bash

go run $GOPATH/src/github.com/bbva/qed/main.go client add --apikey foo -e http://127.0.0.1:8800 --key key$1
