
# QED benchmarking tools

## Requirements:
* Dialog
* Gnuplot
* Vegeta
* Wrk

## Latency test
```
./stress-latency-60s
```
This will run an latency benchmark with Vegeta for 60s by default and output plot.html.

## Throughput test
```
./stress-throughput-60s
```
This script run an throughtput benchmark using Wrk for 60 by default and outputs the results in your terminal.

## Endurance test
```
PROFILING=true ./stress -add -n 1000000
```
It will bring up the server and execute and insert 1M events with default concurrency of 10 connections. Also will start a go profiling tool to collect all internal data of QED server an save it in 'results' directory.
The througput will be also displayed every second. 
The graph is saved in 'results/graph-Add-$number_events.png (results/graph-Add-1000000.png)'.

This script use 'riot.go'. It's a tool we've develop to stress QED. It's also capable of generating charts, start with specific offset or even run Membership|Incremental Proofs for the events into QED.

If you install Dialog and Gnuplot you can generate usage graphs for each function collected by Golang pprof tool. ei:
```
./make_graph results
```







