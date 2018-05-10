
# Server

Run a Qed server for the first time:

```bash
path=/var/tmp/balloon.db
apiKey=server-api-key
mkdir -p ${path}
go run cmd/server/server.go -k ${apiKey} -p ${path}
```

The Qed server has the current flags *(apiKey is the only required):
```txt
Flags:
  -k, --apikey string     Server api key
  -c, --cache uint        Initialize and reserve custom cache size. (default 33554432)
  -e, --endpoint string   Endpoint for REST requests on (host:port) (default "0.0.0.0:8080")
  -h, --help              help for server
  -l, --log string        Choose between log levels: silent, error, info and debug (default "error")
  -p, --path string       Set default storage path. (default "/var/tmp/balloon.db")
  -f, --profiling         Enable profiling server in 6060 port
  -s, --storage string    Choose between different storage backends. Eg badge|bolt (default "badger")
```

# Client

Using a client to pass along simple text messages:
```bash
endpoint=http://localhost:8080
apiKey=server-api-key # this should be the same as the server

echo > input.log
tail -f input.log | go run cmd/cli/qed.go client -k ${apiKey} -e ${endpoint}
```

Echoing a simple line will return a snapshot to stdout:
```bash
echo 'This is my first event' >> input.log
```
```bash
{"HyperDigest":"cM/FhkbOpo5xgyx7+KjW8eaxFLeiVFPIpndNB0geBR8=","HistoryDigest":"qVbKxzBgeoX1fq7Fgodx3PoQOYqQl8czwIJXlt/fTNo=","Version":5,"Event":"VGhpcyBpcyBteSBmaXJzdCBldmVudA=="}
```

# Auditor

Using a auditor to validate snapshots:
```bash
endpoint=http://localhost:8080
apiKey=server-api-key # this should be the same as the server

echo > snapshots.log
echo > input.log
tail -f input.log | go run cmd/cli/qed.go client -k ${apiKey} -e ${endpoint} > snapshots.log &
tail -f snapshots.log | go run cmd/cli/qed.go auditor -k ${apiKey} -e ${endpoint}
```

Echoing a simple line will return a Validation stage to stdout:
```txt
Verify: OK
```
