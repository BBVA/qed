# QED

*([quod erat demostrandum](https://en.wikipedia.org/wiki/Q.E.D.))*

## Background 

Forensics investigations can be flawed for many causes, such as that they can
lack any real evidence of an incident. For that reasons, we have the demand
for an immutable tamper-evident log of everything that happens in ay complex
system.


By the beginning of this research we set an ambitious scope in order to 
accomplish this objective. This is keep as guide for keep the development in
track, to find a efficient prototype capable of fulfilling the following 
requirements:

 - [x] To allow the massive ingestion of heterogeneous logs or data
 - [ ] To have the capability to index data by different fields.
 - [x] To enable an efficient and painless verification process.
 - [x] To allow for a periodic check or snapshot to guarantee immutability
 against third-party agents or audit processes.

# About this project

This project is an implementation of a system that can be used to verify large
amounts of data for:

 * Prove inclusion of value
 * Prove non-inclusion of value
 * Retrieve provable value for key
 * Retrieve provable current value for key
 * Prove append-only
 * Enumerate all entries
 * Prove correct operation
 * Enable detection of split-view

 A glossary of terms can be found [here](docs/glossary.md).
 
 ## Environment
 
 We use the [Go](https://golang.org) programming language and environment as
 described in their  [documentation](https://golang.org/doc/code.html)
 
 
 ## Testing http api
 
 [Document](http://blog.questionable.services/article/testing-http-handlers-go/)
 
 
 ## Guide
 
     $ godoc -http=:6060 # http://localhost:6060/pkg/verifiabledata/
     
     go test verifiabledata/store
     
     go test -v verifiabledata/balloon/history
     go test -v verifiabledata/balloon/hyper
     
     go test -v verifiabledata/balloon
 
     go test -bench="BenchmarkAdd" verifiabledata/balloon/hyper
     go test -bench="BenchmarkAdd" verifiabledata/balloon/history
     
     go test -bench="." -v verifiabledata/balloon


## Server - client test

- start server
    go run server.go -k my-key
	
- add event
    go run qed.go -k my-key -e http://localhost:8080 add --key "test event" --value "2"
	
- membership event
    go run qed.go -k my-key -e http://localhost:8080 membership --historyDigest 444f6e7eee66986752983c1d8952e2f0998488a5b038bed013c55528551eaafa --hyperDigest a45fe00356dfccb20b8bc9a7c8331d5c0f89c4e70e43ea0dc0cb646a4b29e59b --version 0 --key "test event"
	
- verify event
    go run qed.go -k my-key -e http://localhost:8080 membership --historyDigest 444f6e7eee66986752983c1d8952e2f0998488a5b038bed013c55528551eaafa --hyperDigest a45fe00356dfccb20b8bc9a7c8331d5c0f89c4e70e43ea0dc0cb646a4b29e59b --version 0 --key "test event" --verify

### Go profiling

    go run  -cpuprofile cpu.out -memprofile mem.out program.go
    go test -v -bench="BenchmarkAdd" -cpuprofile cpu.out -memprofile mem.out verifiabledata/balloon/hyper -run ^$
     
    go tool pprof hyper.test cpu.out 
    go tool pprof hyper.test cpu.out mem.out
     
The server spawns an http server in port 6060 with the pprof api as described in https://golang.org/pkg/net/http/pprof/

## Acme section

    uncomment Edit .s/^(	| *)\/\//\1/g
    comment Edit .s/^(	| *)/\1\/\//g

