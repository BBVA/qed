## Environment

We use the [Go](https://golang.org) programming language and set up the
environment as described in its [documentation](https://golang.org/doc/code.html)

## Self-signed certificate

```
cd ~/.ssh/

# Generate private key (.key)
openssl genrsa -out server.key 2048


# Generation of self-signed(x509) public key (PEM-encodings .pem|.crt) based on the private (.key)
openssl req -new -x509 -sha256 -key server.key -out server.crt -days 3650
```

## Useful commands

- Go [documentation server](http://localhost:6061/pkg/github.com/bbva/qed/)

```
godoc -http=:6061 # http://localhost:6061/pkg/github.com/bbva/qed/
```

- Test everything

```
go test -v "${GOPATH}"/src/github.com/bbva/qed/...
```
