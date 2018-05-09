## QED

*([quod erat demostrandum](https://en.wikipedia.org/wiki/Q.E.D.))*

## Background 

***qed*** is a software to test the scalability of autenticated data structures. Our mission is to design a system which, even deployed to a non-trusted sever, allows to verify  the integrity of a chain of events and detect modifications of single events or parts of the history.

This software is experimental and part of the hyperscale research being done in BBVA. We will eventually publish our research work, analisys and the experiments for anyone to reproduce. 

 ## Environment
 
 We use the [Go](https://golang.org) programming language and environment as
 described in their  [documentation](https://golang.org/doc/code.html)
 
 ## Starting guide
 
 - Download software
 ```
    go get github.com/bbva/qed
    go get github.com/dgraph-io/badger
    go get github.com/coreos/bbolt
    go get github.com/google/btree
 ```  
 - Start the server
 
 ```
    cd $GOPATH/src/github.com/bbva/qed/cmd/server
    go run server.go -k key -p /var/tmp/db_path -l info
 ```
 
 - Using the client
 
     - add event
 
     ```
        go run qed.go -k my-key -e http://localhost:8080 add --key "test event" --value "2"
     ```
 
     - membership event
 
    ```
        go run qed.go -k my-key -e http://localhost:8080 membership --historyDigest 444f6e7eee66986752983c1d8952e2f0998488a5b038bed013c55528551eaafa --hyperDigest a45fe00356dfccb20b8bc9a7c8331d5c0f89c4e70e43ea0dc0cb646a4b29e59b --version 0 --key "test event"
    ```

     - verify event

    ```
        go run qed.go -k my-key -e http://localhost:8080 membership --historyDigest 444f6e7eee66986752983c1d8952e2f0998488a5b038bed013c55528551eaafa --hyperDigest a45fe00356dfccb20b8bc9a7c8331d5c0f89c4e70e43ea0dc0cb646a4b29e59b --version 0 --key "test event" --verify
    ```

## Useful commands

- Go documentation server

  ```
     $ godoc -http=:6061 # http://localhost:6061/pkg/qed/
  ```
  
- Test everything
 
  ```
     go test -v github.com/bbva/qed/...
  ```   



### Go profiling

  ```  
    go run  -cpuprofile cpu.out -memprofile mem.out program.go
    go test -v -bench="BenchmarkAdd" -cpuprofile cpu.out -memprofile mem.out qed/balloon/hyper -run ^$
     
    go tool pprof hyper.test cpu.out 
    go tool pprof hyper.test cpu.out mem.out
  ```

The server spawns an http server in port 6060 with the pprof api as described in https://golang.org/pkg/net/http/pprof/


