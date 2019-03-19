#!/bin/bash

#  Copyright 2018-2019 Banco Bilbao Vizcaya Argentaria, S.A.
#  Licensed under the Apache License, Version 2.0 (the "License");
#  you may not use this file except in compliance with the License.
#  You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
#  Unless required by applicable law or agreed to in writing, software
#  distributed under the License is distributed on an "AS IS" BASIS,
#  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
#  See the License for the specific language governing permissions and
#  limitations under the License.

qedGossipEndpoint="127.0.0.1:8400"
snapshotStoreEndpoint="http://127.0.0.1:8888"
alertsStoreEndpoint="http://127.0.0.1:8888"
qedHTTPEndpoint="http://127.0.0.1:8800"
keyFile="/var/tmp/id_ed25519"
QED="go run $GOPATH/src/github.com/bbva/qed/main.go"

if [ ! -f "$keyFile" ]; then 
	echo Create id_ed25519 key
	echo -e 'y\n' | ssh-keygen -t ed25519 -N '' -f /var/tmp/id_ed25519
fi

xterm -hold -e "$QED start  -k key  -l debug  -p $(mktemp -d)  --node-id server0  --raft-addr 127.0.0.1:8500 --gossip-addr 127.0.0.1:8400 --mgmt-addr 127.0.0.1:8700 --metrics-addr 127.0.0.1:8600 --http-addr 127.0.0.1:8800 --keypath $keyFile" &
pids[0]=$!

sleep 3s

xterm -hold -e "$QED start -k key -l debug -p $(mktemp -d) --node-id server1 --gossip-addr 127.0.0.2:8401 --raft-addr 127.0.0.2:8501 --keypath $keyFile --join-addr 127.0.0.1:8700 --gossip-join-addr 127.0.0.1:8400 --http-addr 127.0.0.2:8801 --mgmt-addr 127.0.0.2:8701 --metrics-addr 127.0.0.2:8601" &
pids+=($!)

sleep 2s

for i in `seq 1 $1`;
do
	xterm -hold -e "$QED agent --metrics 127.0.0.2:1810$i  auditor -k key -l debug --bind 127.0.0.1:810$i --join $qedGossipEndpoint --qedUrls $qedHTTPEndpoint --pubUrls $snapshotStoreEndpoint --node auditor$i --alertsUrls $alertsStoreEndpoint" &
	pids+=($!)
done 

for i in `seq 1 $2`;
do
	xterm  -hold -e "$QED agent --metrics 127.0.0.2:1820$i --alertsUrls $alertsStoreEndpoint monitor -k key -l debug --bind 127.0.0.1:820$i --join $qedGossipEndpoint --qedUrls $qedHTTPEndpoint  --node monitor$i " &
	pids+=($!)
done 

for i in `seq 1 $3`;
do
	xterm -hold -e "$QED agent --metrics 127.0.0.2:1830$i --alertsUrls $alertsStoreEndpoint publisher -k key -l debug --bind 127.0.0.1:830$i --join $qedGossipEndpoint --pubUrls $snapshotStoreEndpoint --node publisher$i " &
	pids+=($!)
done 

for pid in ${pids[*]}; do
	echo waiting for pid $pid
	wait $pid
done
