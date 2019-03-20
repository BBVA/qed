#!/bin/bash


go run $GOPATH/src/github.com/bbva/qed/main.go 	\
	--apikey my-key \
	client \
		--log info \
		--endpoints http://${QED_LEADER}:8800 \
		membership --key key$1 \
		--version $1 --verify
