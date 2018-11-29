#!/bin/bash

go run $GOPATH/src/github.com/bbva/qed/main.go client add -k key -e http://127.0.0.1:8080 --key key$1 --value value$1
