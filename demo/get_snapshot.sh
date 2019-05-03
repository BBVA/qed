#!/usr/bin/env bash

curl -s "127.0.0.1:8888/snapshot?v=0" | python -m json.tool