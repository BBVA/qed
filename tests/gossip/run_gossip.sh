#!/bin/bash

#  Copyright 2018 Banco Bilbao Vizcaya Argentaria, S.A.
#  Licensed under the Apache License, Version 2.0 (the "License");
#  you may not use this file except in compliance with the License.
#  You may obtain a copy of the License at
#      http://www.apache.org/licenses/LICENSE-2.0
#  Unless required by applicable law or agreed to in writing, software
#  distributed under the License is distributed on an "AS IS" BASIS,
#  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
#  See the License for the specific language governing permissions and
#  limitations under the License.

master="127.0.0.1:9100"
publisher="http://127.0.0.1:8888"
qed="http://127.0.0.1:8080"
echo Create id_ed25519 key
echo -e 'y\n' | ssh-keygen -t rsa -N '' -f /var/tmp/id_ed25519
go run $GOPATH/src/github.com/bbva/qed/main.go start  -k key -l silent  --node-id server0 --gossip-addr $master --raft-addr 127.0.0.1:9000 &
pids[0]=$!
sleep 1s

for i in `seq 1 $1`;
do
	xterm -hold -e "go run $GOPATH/src/github.com/bbva/qed/main.go agent auditor -k key -l info --bind 127.0.0.1:910$i --join $master --endpoints $qed --node auditor$i" &
	pids+=($!)
done 

for i in `seq 1 $2`;
do
	xterm -hold -e "go run $GOPATH/src/github.com/bbva/qed/main.go agent monitor -k key -l info --bind 127.0.0.1:920$i --join $master --endpoints $qed --node monitor$i" &
	pids+=($!)
done 

for i in `seq 1 $3`;
do
	xterm -hold -e "go run $GOPATH/src/github.com/bbva/qed/main.go agent publisher -k key -l info --bind 127.0.0.1:930$i --join $master --endpoints $publisher --node publisher$i" &
	pids+=($!)
done 

for pid in ${pids[*]}; do
	echo waiting for pid $pid
	wait $pid
done
