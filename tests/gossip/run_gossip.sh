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
go run $GOPATH/src/github.com/bbva/qed/main.go start  -k key -l debug  --node-id server0 --gossip-addr $master --raft-addr 127.0.0.1:9000 -y $HOME/.ssh/id_ed25519 &
pids[0]=$!
sleep 1s

for i in `seq 1 $1`;
do
	xterm -hold -e "go run $GOPATH/src/github.com/bbva/qed/main.go agent -k key -l debug --bind 127.0.0.1:910$i --join $master --node auditor$i --role auditor" &
	pids+=($!)
done 

for i in `seq 1 $2`;
do
	xterm -hold -e "go run $GOPATH/src/github.com/bbva/qed/main.go agent -k key -l debug --bind 127.0.0.1:920$i --join $master --node monitor$i --role monitor" &
	pids+=($!)
done 

for pid in ${pids[*]}; do
	echo waiting for pid $pid
	wait $pid
done
