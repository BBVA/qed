#!/usr/bin/env bash

go run ../main.go client membership --api-key key --insecure --event $(cat event0.json) --verify --log info