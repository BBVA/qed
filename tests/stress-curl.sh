#!/usr/bin/env bash
set -a
set -e

time for i in {1..1000}
do
        curl -s localhost:8080/events     \
        -H "Api-Key: this-is-my-key"      \
        -d '{"Message": "Test event $x"}' \
        -o /dev/null
done

