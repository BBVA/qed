#!/usr/bin/env bash
# echo Clean old data
# rm /tmp/targets.txt
#
# echo Generate event json
# for i in {1..100000}; do echo -e "{\"message\": \"Load Test $i\"}" > /tmp/event$i.json;done
#
# echo Generate targets file
# for i in {1..100000}; do echo -e "POST http://localhost:8080/events\nContent-Type: application/json\nApi-key: this-is-my-dummy-key\n@/tmp/event$i.json\n" >> /tmp/targets.txt;done
#
# echo Execute Vegeta benchmark
# vegeta attack -timeout 100s -rate 1500 -duration=60s -targets=/tmp/targets.txt > result.bin

go run attack_add.go -k key -r 250 -d 120s > result.bin

cat result.bin | vegeta report
cat result.bin | vegeta report -reporter='plot' > plot.html
