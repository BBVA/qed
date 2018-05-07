#!/usr/bin/env bash
vegeta attack -connections 1 -workers 1 -rate 1500 -duration=10s -targets=stress/targets.txt > result.bin 
cat result.bin | vegeta report -reporter='hist[10ms,100ms,500ms,1s,10s,1m]'
cat result.bin | vegeta report -reporter='plot' > plot.html