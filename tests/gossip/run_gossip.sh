#!/bin/bash

master="127.0.0.1:9000"
go run $GOPATH/src/github.com/bbva/qed/main.go agent  -k key -l silent --bind $master  --node 0 --role server &
pids[0]=$!
sleep 1s

for i in `seq 1 $1`;
do
	xterm -hold -e "go run $GOPATH/src/github.com/bbva/qed/main.go agent -k key -l silent --bind 127.0.0.1:900$i --join $master --node $i --role monitor" &
	pids[${i}]=$!
done 

for pid in ${pids[*]}; do
	echo waiting for pid $pid
	wait $pid
done
