#!/bin/bash

set -e 

# QED="go run ../cmd/cli/qed.go -k path -e http://localhost:8080"
QED="qed -l info -k pepe -e http://localhost:8080"

add_event(){
	local event="$1"; shift
	local value="$1"; shift
	$QED  add --key "${event}" --value "${value}"
}


#Adding key [ test event ] with value [ 2 ]
#test event
#Received snapshot with values: 
#	Event: test event
#	HyperDigest: a45fe00356dfccb20b8bc9a7c8331d5c0f89c4e70e43ea0dc0cb646a4b29e59b
#	HistoryDigest: 444f6e7eee66986752983c1d8952e2f0998488a5b038bed013c55528551eaafa
#	Version: 0

verify_event() {
	local commitment="$1"; shift
	echo "${commitment}"
	local event=$(echo "${commitment}" | grep "Event: " | awk -F': ' '{print $2;}')
	local history=$(echo "${commitment}" | grep "HistoryDigest" | awk -F': ' '{print $2;}')
	local hyper=$(echo "${commitment}" | grep "HyperDigest: " | awk -F': ' '{print $2;}')
	local version=$(echo "${commitment}" | grep "Version: " | awk -F': ' '{print $2;}')
	
	
	$QED membership --historyDigest ${history}   --hyperDigest ${hyper}  --version ${version} --key ${event} --verify
}


for i in $(seq 1 1000); do
	event=$(cat /dev/urandom | tr -dc 'a-zA-Z0-9' | fold -w 32 | head -n 1)
	commitment=$(add_event "${event}" "42")
	verify_event "${commitment}"
done

