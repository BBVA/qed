# Advanced Usage

## Overview

Besides the standalone example given in the README.md, QED are created to be a
production-ready cluster. Here you can find some detailed examples.

## QED cluster

In order to garantee reliability and High Availabity QED storage servers are 
created around hashicorp's [raft](https://github.com/hashicorp/raft) implementation.

### Starting cluster mode

To have identified the leader beforehand, we launch a server and then some
disposable followers.

```bash
go run main.go start \
    -k my-key \
    -p $(mktemp -d /var/tmp/demo.XXX) \
    --raftpath $(mktemp -d /var/tmp/demo.XXX) \
    -y ~/.ssh/id_ed25519-qed \
    --http-addr :8080 \
    --raft-addr :9000 \
    --mgmt-addr :8090 \
    -l error
```

### Starting two followers
```bash
CLUSTER_SIZE=2
for i in $(seq 1 $CLUSTER_SIZE); do
    go run main.go start \
        -k my-key \
        -p $(mktemp -d /var/tmp/demo.XXX) \
        --raftpath $(mktemp -d /var/tmp/demo.XXX) \
        -y ~/.ssh/id_ed25519-qed \
        --http-addr :808$i \
        --join-addr :8090 \
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

## test_service
To use a demo public log-storage QED provides a small helper wich uses a in-memory
structure to store signed snapshots.

```bash
go run tests/e2e/gossip/test_service.go
```

### Publisher

```bash
# this variables will be used in all the publiser, auditor and monitor examples.
export masterEndpoint="127.0.0.1:9100"
export publisherEndpoint="http://127.0.0.1:8888"
export alertsEndpoint="http://127.0.0.1:8888"
export qedEndpoint="http://127.0.0.1:8080"
```

```bash
go run main.go agent \
    --alertsUrls $alertsEndpoint \
    publisher \
    -k my-key \
    --bind 127.0.0.1:9300 \
    --join $masterEndpoint \
    --endpoints $publisherEndpoint \
    --node publisher0 \
    -l info
```

### auditor
```bash
go run main.go agent \
    --alertsUrls $alertsEndpoint \
    auditor \
    -k my-key \
    --bind 127.0.0.1:9100 \
    --join $masterEndpoint \
    --qedUrls $qedEndpoint \
    --pubUrls $publisherEndpoint \
    --node auditor0 \
    -l info
```

### monitor
```bash
go run main.go agent \
    --alertsUrls $alertsEndpoint \
     monitor \
     -k my-key \
     --bind 127.0.0.1:920$i \
     --join $masterEndpoint \
     --endpoints $qedEndpoint \
     --node monitor0 \
     -l info
```
