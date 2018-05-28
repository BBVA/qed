[![Build Status](https://travis-ci.org/BBVA/qed.svg?branch=master)](https://travis-ci.org/BBVA/qed)
[![Coverage](https://codecov.io/gh/BBVA/qed/branch/master/graph/badge.svg)](https://codecov.io/gh/BBVA/qed)

<p align="center"><a href="https://en.wikipedia.org/wiki/Q.E.D."><img src="./qed_logo.png" alt="Quod Erat Demonstrandum"/><br/>(quod erat demostrandum)</a></p>


## Overview

***qed*** is a software to test the scalability of authenticated data structures. Our mission is to design a system which, even when deployed into a non-trusted server, allows one to verify the integrity of a chain of events and detect modifications of single events or parts of its history.

This software is experimental and part of the research being done at BBVA Labs. We will eventually publish our research work, analysis and the experiments for anyone to reproduce.

## Motivation
The use of a technology that allows to verify the information it stores is quite broad. Audit logs are a common tool for forensic investigations and legal proceedings due to its utility for detecting database tampering. Malicious users, including insiders with high-level access, may perform unlogged activities or tamper with the recorded history. The evidence one seeks in these sorts of investigations often takes the form of statements of existence and order. But this kind of tamper-evident logs have also been used for other use cases: building versioned filesystems like version control systems, p2p protocols or as a mechanism to detect conflicts in distributing systems, like data inconsistencies between replicas.

All of these use cases share something in common: the proof of order and integrity is fulfilled building data structures based on the concept of hash chaining. This technique allows to establish a provable order between entries, and comes with the benefit of tamper-evidence, ensuring that any commitment to a given state of the log is implicitly a commitment to all prior states. Therefore, any subsequent attempt to remove or alter some log entries will invalidate the hash chain.

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
    cd "$GOPATH"/src/github.com/bbva/qed
    mkdir /var/tmp/db_path
    go run cmd/server/server.go -k key -p /var/tmp/db_path -l info
 ```

 - Using the client

     - add event

     ```
	go run				\
		cmd/cli/qed.go		\
		-k my-key			\
		-e http://localhost:8080	\
		add				\
		--key 'test event'		\
		--value 2
     ```

     - membership event

    ```
	go run												\
		cmd/cli/qed.go										\
		-k my-key										\
		-e http://localhost:8080								\
		membership										\
		--historyDigest 444f6e7eee66986752983c1d8952e2f0998488a5b038bed013c55528551eaafa	\
		--hyperDigest a45fe00356dfccb20b8bc9a7c8331d5c0f89c4e70e43ea0dc0cb646a4b29e59b		\
		--version 0										\
		--key 'test event'
    ```

     - verify event

    ```
	go run												\
		cmd/cli/qed.go										\
		-k my-key										\
		-e http://localhost:8080								\
		membership										\
		--historyDigest 444f6e7eee66986752983c1d8952e2f0998488a5b038bed013c55528551eaafa	\
		--hyperDigest a45fe00356dfccb20b8bc9a7c8331d5c0f89c4e70e43ea0dc0cb646a4b29e59b		\
		--version 0										\
		--key 'test event'									\
		--verify
    ```
    See [usage](docs/usage.md) for the gory details.

## Useful commands

- Go documentation server

  ```
     $ godoc -http=:6061 # http://localhost:6061/pkg/qed/
  ```

- Test everything

  ```
     go test -v "$GOPATH"/src/github.com/bbva/qed/...
  ```
- Go profiling

  ```
	go run
		-cpuprofile cpu.out						\
		-memprofile mem.out						\
		program.go

	go test								\
		-v								\
		-bench=BenchmarkAdd						\
		-cpuprofile cpu.out						\
		-memprofile mem.out						\
		qed/balloon/hyper						\
		-run ^$

    go tool pprof hyper.test cpu.out
    go tool pprof hyper.test cpu.out mem.out
  ```

The server spawns an http server on port 6060 with the pprof api as described in https://golang.org/pkg/net/http/pprof/

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
