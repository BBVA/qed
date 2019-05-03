#!/usr/bin/env bash

go run ../main.go client membership --api-key key --insecure add --event $(cat event0.json) --log info
