# Advanced Usage

## Overview

Besides the standalone example given in the [README](../README.md), QED is also designed 
to be a production-ready cluster. Here you can find some detailed examples.

<p align="center"><img width="100%" src="./architecture/full_architecture.png" alt="Architecture overview"/></p>

## QED cluster

In order to guarantee reliability and high availabity, QED servers include
[hashicorp's raft](https://github.com/hashicorp/raft) consensus protocol implementation.
An architectural perspective can be found at [raft](architecture/raft.md) doc. file. 

To have identified the leader beforehand (demo purpose), launch first a single 
cluster-ready server, and then some disposable followers.

### Starting cluster mode

```bash
go run main.go start \
    -k my-key \
    -p $(mktemp -d /var/tmp/demo.XXX) \
    --raftpath $(mktemp -d /var/tmp/demo.XXX) \
    -y ~/.ssh/qed_ed25519 \
    --http-addr :8800 \
    --raft-addr :8500 \
    --mgmt-addr :8700 \
    -l error
```

### Starting two followers
```bash
FOLLOWERS=2
for i in $(seq 1 $FOLLOWERS); do
    go run main.go start \
        -k my-key \
        -p $(mktemp -d /var/tmp/demo.XXX) \
        --raftpath $(mktemp -d /var/tmp/demo.XXX) \
        -y ~/.ssh/qed_ed25519 \
        --http-addr :808$i \
        --join-addr :8700 \
        --raft-addr :900$i \
        --mgmt-addr :809$i \
        --node-id node$i \
        -l error
done
```

Know events must be added **ONLY** in the leader, but events can be verified in
any follower (and it's the way to go).

A Quick example could be use the README standalone client example changing the
endpoint port `--endpoint http://localhost:8081` in the verify event command.


## Agents

In order to allow public `auditors`, we need to ensure a public storage in which 
QED server sends the snapshot information. `publisher` agents are designed to do 
it. Finally `monitors` agents are check internal consitency.

QED sends signed snapshots to, at least, one agent of each type. 
QED uses Gossip protocol for message passing between server and agents, and between
agents themselves.
An architectural perspective can be found at [gossip](architecture/gossip.md) doc. file. 


### Aux service

For demo purposes, QED provides an auxiliary service which uses 
an in-memory structure to store signed snapshots that acts as a public log-storage.

Moreover, it provides an alert endpoint to allow agents register its alerts.

```bash
go run testutils/notifierstore.go
```

To be production-ready, both services must be developed and deployed separatelly.

### Publisher

```bash
# this variables will be used in all the publiser, auditor and monitor examples.
export masterEndpoint="127.0.0.1:8100"
export publisherEndpoint="http://127.0.0.1:8888"
export alertsEndpoint="http://127.0.0.1:8888"
export qedEndpoint="http://127.0.0.1:8800"
```

```bash
go run main.go agent \
    --alertsUrls $alertsEndpoint \
    publisher \
    -k my-key \
    --bind 127.0.0.1:8300 \
    --join $masterEndpoint \
    --endpoints $publisherEndpoint \
    --node publisher0 \
    -l info
```

### Auditor

```bash
go run main.go agent \
    --alertsUrls $alertsEndpoint \
    auditor \
    -k my-key \
    --bind 127.0.0.1:8100 \
    --join $masterEndpoint \
    --qedUrls $qedEndpoint \
    --pubUrls $publisherEndpoint \
    --node auditor0 \
    -l info
```

### Monitor

```bash
go run main.go agent \
    --alertsUrls $alertsEndpoint \
    monitor \
    -k my-key \
    --bind 127.0.0.1:8200 \
    --join $masterEndpoint \
    --endpoints $qedEndpoint \
    --node monitor0 \
    -l info
```
