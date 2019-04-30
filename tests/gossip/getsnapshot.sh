#!/bin/bash

raw=$(curl -s "http://localhost:8888/snapshot?v=$1")

if [ $(uname) == "Darwin" ]; then
    history=$(echo $raw | jq -r '.Snapshot.HistoryDigest' | base64 -D | xxd -c64 -g64 | cut -f2 -d' ')
	hyper=$(echo $raw | jq -r '.Snapshot.HyperDigest' | base64 -D | xxd -c64 -g64 | cut -f2 -d' ')
else
    history=$(echo $raw | jq -r '.Snapshot.HistoryDigest' | base64 -d | xxd -c64 -g64 | cut -f2 -d' ')
	hyper=$(echo $raw | jq -r '.Snapshot.HyperDigest' | base64 -d | xxd -c64 -g64 | cut -f2 -d' ')
fi

echo HyperDigest: $hyper
echo HistoryDigest: $history

