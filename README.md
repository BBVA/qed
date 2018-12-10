[![Build Status](https://travis-ci.org/BBVA/qed.svg?branch=master)](https://travis-ci.org/BBVA/qed)
[![Coverage](https://codecov.io/gh/BBVA/qed/branch/master/graph/badge.svg)](https://codecov.io/gh/BBVA/qed)
[![GoReport](https://goreportcard.com/badge/github.com/bbva/qed)](https://goreportcard.com/report/github.com/bbva/qed)
[![GoDoc](https://godoc.org/github.com/bbva/qed?status.svg)](https://godoc.org/github.com/bbva/qed)

<p align="center"><a href="https://en.wikipedia.org/wiki/Q.E.D."><img width="55%" src="./qed_logo.png" alt="Quod Erat Demonstrandum"/><br/>(quod erat demonstrandum)</a></p>
<p align="center"><img width="90%" src="./qed_whiteboard.png" alt="Whiteboard depicting a use case for qed"/></p>


## Overview

***qed*** is a software to test the scalability of authenticated data structures. Our mission is to design a system which, even when deployed into a non-trusted server, allows one to verify the integrity of a chain of events and detect modifications of single events or parts of its history.

This software is experimental and part of the research being done at BBVA Labs. We will eventually publish our research work, analysis and the experiments for anyone to reproduce.

## Motivation
The use of a technology that allows to verify the information it stores is quite broad. Audit logs are a common tool for forensic investigations and legal proceedings due to its utility for detecting database tampering. Malicious users, including insiders with high-level access, may perform unlogged activities or tamper with the recorded history. The evidence one seeks in these sorts of investigations often takes the form of statements of existence and order. But this kind of tamper-evident logs have also been used for other use cases: building versioned filesystems like version control systems, p2p protocols or as a mechanism to detect conflicts in distributing systems, like data inconsistencies between replicas.

All of these use cases share something in common: the proof of order and integrity is fulfilled building data structures based on the concept of hash chaining. This technique allows to establish a provable order between entries, and comes with the benefit of tamper-evidence, ensuring that any snapshot to a given state of the log is implicitly a snapshot to all prior states. Therefore, any subsequent attempt to remove or alter some log entries will invalidate the hash chain.

In order to prove that an entry has been included in the information storage, and that it has not been modified in an inconsistent way we need:

* **Proof of inclusion**, answering the question about if a given entry is in the log or not.
* **Proof of consistency**, answering the question about if a given entry is consistent with the prior ones. This ensures the recorded history has not been altered.
* **Proof of deletion**, so we are able to know when a log has been tampered with at its source location.

Some of the systems that can be built upon these technologies are:

* Verifiable Application Log - verifiable operation activity
* Verifiable Security Audit - verifiable authentication / authorization activity
* Verifiable Transaction Log - verifiable business activity
* Verifiable Data Blocks - verifiable HDFS blocks

A number of hash data structures have been proposed for storing data in a tamper-evident fashion (see references [below](#other-projects-papers-and-references)). All of them have at their core a Merkle tree or some variant.

Our work draws strongly from the **Balloon proposals**, with some modifications of our own that aim to improve scalability.

## Environment

 We use the [Go](https://golang.org) programming language and set up the environment as
 described in its [documentation](https://golang.org/doc/code.html)

## Getting started

 - Download the software and its dependencies
 ```
 go get -v -u -d github.com/bbva/qed/...
 ```
 - Start the server

 ```
 cd "$GOPATH/src/github.com/bbva/qed"
 rm -rf /var/tmp/qed
 mkdir -p /var/tmp/qed/{data,raft}
 ssh-keygen -t ed25519 -P '' -f ~/.ssh/id_ed25519-qed
 go run main.go start --apikey key --keypath ~/.ssh/id_ed25519-qed --log info
 ```

 - Using the client

     - add event

    ```
    go run \
    main.go \
    --apikey my-key \
    client \
    --endpoint http://localhost:8080 \
    add \
    --key 'test event' \
    --value 2 \
    --log info
    ```

     - membership event

    ```
    go run \
    main.go \
    --apikey my-key \
    client \
    --endpoint http://localhost:8080 \
    membership \
    --hyperDigest 3ec11c37f0a53ff5c4cfc3cf2573c33a9721cd25d8e670a3b2be0fda5724bb5c \
    --historyDigest 776b33eab8ed829ecffab3d579bf7ccbcc126b94bac1aaca7d5d8b0a2687bdec \
    --version 0 \
    --key 'test event' \
    --log info
    ```

     - verify event

    ```
    go run \
    main.go \
    --apikey my-key \
    client \
    --endpoint http://localhost:8080 \
    membership \
    --hyperDigest 3ec11c37f0a53ff5c4cfc3cf2573c33a9721cd25d8e670a3b2be0fda5724bb5c \
    --historyDigest 776b33eab8ed829ecffab3d579bf7ccbcc126b94bac1aaca7d5d8b0a2687bdec \
    --version 0 \
    --key 'test event' \
    --log info \
    --verify
    ```

## Useful commands

- Go [documentation server](http://localhost:6061/pkg/github.com/bbva/qed/)

```
godoc -http=:6061 # http://localhost:6061/pkg/github.com/bbva/qed/
```

- Test everything

```
go test -v "$GOPATH"/src/github.com/bbva/qed/...
```
## Other projects, papers and references

- github related projects
   - [Balloon](https://github.com/pylls/balloon)
   - [GoSMT](https://github.com/pylls/gosmt)
   - [Trillian](https://github.com/google/trillian)
   - [Continusec](https://github.com/continusec/verifiabledatastructures)

 - related papers
   - https://github.com/google/trillian/blob/master/docs/VerifiableDataStructures.pdf
   - http://tamperevident.cs.rice.edu/papers/paper-treehist.pdf
   - http://kau.diva-portal.org/smash/get/diva2:936353/FULLTEXT01.pdf
   - http://www.links.org/files/sunlight.html
   - http://www.links.org/files/RevocationTransparency.pdf
   - https://eprint.iacr.org/2015/007.pdf
   - https://eprint.iacr.org/2016/683.pdf

## Contributions

Contributions are very welcome, see [CONTRIBUTING.md](https://github.com/BBVA/qed/blob/master/CONTRIBUTING.md)
or skim [existing tickets](https://github.com/BBVA/qed/issues) to see where you could help out.

## License

***qed*** is Open Source and available under the Apache 2 License.
