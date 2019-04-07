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

QED="CGO_LDFLAGS_ALLOW='.*' go run $GOPATH/src/github.com/bbva/qed/main.go"

# Agent options
AGENT_CONFIG=()
AGENT_CONFIG+=('--agent-log debug')
AGENT_CONFIG+=('--agent-bind-addr 127.0.0.1:810${i}')
AGENT_CONFIG+=('--agent-metrics-addr 127.0.0.2:1810${i}')
AGENT_CONFIG+=('--agent-start-join 127.0.0.1:8400')


# Notifier options
NOTIFIER_CONFIG=()
NOTIFIER_CONFIG+=('--notifier-servers http://127.0.0.1:8888')

# Snapshot store options
STORE_CONFIG=()
STORE_CONFIG+=('--store-servers http://127.0.0.1:8888')

# Task manager options
TASKS_CONFIG=()
TASKS_CONFIG+=("")

# QED client options
QED_CONFIG=()
QED_CONFIG+=("--qed-endpoints http://127.0.0.1:8800")


MONITOR_CONFIG=("${AGENT_CONFIG[@]}" "${NOTIFIER_CONFIG[@]}" "${STORE_CONFIG[@]}" "${TASKS_CONFIG[@]}" "${QED_CONFIG[@]}")
MONITOR_CONFIG+=("--agent-role monitor")
MONITOR_CONFIG+=("--agent-node-name monitor${i}")

pids=()

n="$1"
if [ -z "$n" ]; then
	n=1
fi

for id in $(seq 0 1 $n); do
	monitor=$(echo ${MONITOR_CONFIG[@]} | i=$id envsubst )
	xterm -hold -e "$QED agent monitor $monitor " &
	pids+=($!)
	sleep 3s
done


for pid in ${pids[*]}; do
	echo waiting for pid $pid
	wait $pid
done
