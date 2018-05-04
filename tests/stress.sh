#!/usr/bin/env bash
vegeta attack -connections 1 -workers 1 -rate 100 -duration=60s -targets=stress/targets.txt | vegeta report -reporter='hist[10ms,100ms,500ms,1s,10s,1m]'
