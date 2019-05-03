#!/usr/bin/env bash

echo Starting services
../tests/start_server &
sleep 5
../tests/start_agent publisher &
sleep 5
go run ../testutils/notifierstore.go &
