#!/usr/bin/env bash

msg=$(cat project/go.mod | xargs )
version="v1.3.0"
hash=$(sha1sum project/go.mod | cut -d' ' -f1)

echo "
{
	\"msg\": \"$msg\",
	\"version\": \"$version\",
	\"hash\": \"$hash\"
}
" > event0.json
go run ../main.go client --api-key key --insecure add --event $(cat event0.json) --log info
