#!/usr/bin/env bash

set -eu

apikey=$1; shift
echo "starting with $apikey"

tmppipe=$(mktemp -u)
mkfifo -m 600 "${tmppipe}"


generate_body() {
	i=0
	while true ; do
		echo -e "{\"message\": \"Load Test $i\"}" > ${tmppipe}
		i=$(( $i + 1 ))
		echo generating event $i >> /dev/stderr
	done
}

generate_targets() {
	while true; do
		echo generating target >> /dev/stderr
		cat <<- EOF  
		POST http://localhost:8080/events
		Content-Type: application/json
		Api-key: ${apikey}
		@${tmppipe}
		EOF
	done
}

generate_body &
generate_targets  | exec vegeta -cpus 1 attack -lazy -workers 1 -timeout 10s -rate 1500 -duration=0 > result.bin  

cat result.bin | vegeta report
cat result.bin | vegeta report -reporter='plot' > plot.html
rm ${tmppipe}
