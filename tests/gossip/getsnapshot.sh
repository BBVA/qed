#!/bin/bash

raw=$(curl -s "http://localhost:8888/snapshot?v=$1")

history=$(echo $raw | jq -r '.Snapshot.HistoryDigest' | base64 -d | xxd -c64 -g64 | cut -f2 -d' ')
hyper=$(echo $raw | jq -r '.Snapshot.HyperDigest' | base64 -d | xxd -c64 -g64 | cut -f2 -d' ')

echo HyperDigest: $hyper
echo HistoryDigest: $history

