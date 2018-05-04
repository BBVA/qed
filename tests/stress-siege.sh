#!/usr/bin/env bash

docker run --rm -t --net=host yokogawa/siege -b -c1 -t60s -H 'Content-Type: application/json' -H 'Api-Key: this-is-my-api-key' 'http://localhost:8080/events POST {"Message": "Test event"}'
